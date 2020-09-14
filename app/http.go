package app

import (
	"compress/gzip"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/growerlab/mensa/app/common"
	"github.com/growerlab/mensa/app/conf"
	"github.com/pkg/errors"
)

const (
	RpcUploadPack  = "upload-pack"
	RpcReceivePack = "receive-pack"
)

func NewGitHttpServer(cfg *conf.Config) *GitHttpServer {
	deadline := DefaultDeadline
	idleTimeout := DefaultIdleTimeout

	if cfg.Deadline > 0 {
		deadline = cfg.Deadline
	}
	if cfg.IdleTimeout > 0 {
		idleTimeout = cfg.IdleTimeout
	}

	server := &GitHttpServer{
		listen:      cfg.HttpListen,
		gitBinPath:  cfg.GitPath,
		deadline:    deadline,
		idleTimeout: idleTimeout,
	}

	server.services = map[string]service{
		"(.*?)/git-upload-pack$":                       {"POST", RpcUploadPack, server.serviceRpc},
		"(.*?)/git-receive-pack$":                      {"POST", RpcReceivePack, server.serviceRpc},
		"(.*?)/info/refs$":                             {"GET", "", server.getInfoRefs},
		"(.*?)/HEAD$":                                  {"GET", "", server.getTextFile},
		"(.*?)/objects/info/alternates$":               {"GET", "", server.getTextFile},
		"(.*?)/objects/info/http-alternates$":          {"GET", "", server.getTextFile},
		"(.*?)/objects/info/packs$":                    {"GET", "", server.getInfoPacks},
		"(.*?)/objects/info/[^/]*$":                    {"GET", "", server.getTextFile},
		"(.*?)/objects/[0-9a-f]{2}/[0-9a-f]{38}$":      {"GET", "", server.getLooseObject},
		"(.*?)/objects/pack/pack-[0-9a-f]{40}\\.pack$": {"GET", "", server.getPackFile},
		"(.*?)/objects/pack/pack-[0-9a-f]{40}\\.idx$":  {"GET", "", server.getIdxFile},
	}
	return server
}

type requestContext struct {
	w    http.ResponseWriter
	r    *http.Request
	Rpc  string
	Dir  string
	File string
}

type service struct {
	Method string
	Rpc    string
	Do     func(*requestContext) error
}

type GitHttpServer struct {
	// 服务器的监听地址(eg. host:port)
	listen string
	// git bin path
	gitBinPath string
	// logger
	// logger io.Writer
	// 最长执行时间
	deadline int
	// 限制最大时间
	idleTimeout int
	// services
	services map[string]service
	// 回调
	middlewareHandler MiddlewareHandler
}

func (g *GitHttpServer) ListenAndServe(handler MiddlewareHandler) error {
	log.Printf("[http] git listen and serve: %v\n", g.listen)
	g.middlewareHandler = handler

	if err := g.validate(); err != nil {
		return err
	}

	g.prepre()

	if err := http.ListenAndServe(g.listen, g); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (g *GitHttpServer) validate() error {
	if g.listen == "" {
		return errors.New("addr is required")
	}
	if !strings.Contains(g.listen, ":") {
		return errors.Errorf("addr is invalid: %s", g.listen)
	}
	return nil
}

func (g *GitHttpServer) prepre() {

}

// TODO 平滑重启
func (g *GitHttpServer) Shutdown() error {
	return nil
}

func (g *GitHttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	begin := time.Now()
	preLog := fmt.Sprintf("METHOD: %s IP: %s URL: %s [%s]", r.Method, r.RemoteAddr, r.URL.String(), begin.Format(time.RFC3339))
	defer func() {
		end := time.Now()
		log.Printf("[http] %s-[%s] TAKE: %s\n", preLog, end.Format(time.RFC3339), end.Sub(begin))
	}()

	ctx, err := common.BuildContextFromHTTP(w, r)
	if err != nil {
		g.httpRender(w, http.StatusBadRequest, "bad request")
		return
	}

	result := g.middlewareHandler(ctx)
	if result != nil {
		log.Printf("[http] middleware err: %+v \nresult:%d %s\n", result.Err, result.HttpCode, result.HttpMessage)
		g.httpRender(w, result.HttpCode, "")
		return
	}

	// git服务
	for match, service := range g.services {
		re, err := regexp.Compile(match)
		if err != nil {
			log.Println(errors.Errorf("does not match service: %s", match))
			return
		}

		if m := re.FindStringSubmatch(r.URL.Path); m != nil {
			log.Println("[http] git handler commands: ", m)
			if service.Method != r.Method {
				g.httpRender(w, http.StatusMethodNotAllowed, "invalid method: "+r.Method)
				return
			}

			file := strings.Replace(r.URL.Path, m[1]+"/", "", 1)
			req := &requestContext{
				w:    w,
				r:    r,
				Rpc:  service.Rpc,
				Dir:  ctx.RepoDir,
				File: file,
			}

			err = service.Do(req)
			if err != nil {
				log.Printf("[http] service.Do was err: %+v\n", err)
			}
			return
		}
	}
	g.httpRender(w, http.StatusBadRequest, "invalid command")
}

func (g *GitHttpServer) httpRender(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	_, _ = w.Write([]byte(message))
}

func (g *GitHttpServer) getInfoPacks(ctx *requestContext) error {
	g.hdrCacheForever(ctx.w)
	g.sendFile("text/plain; charset=utf-8", ctx)
	return nil
}

func (g *GitHttpServer) getLooseObject(ctx *requestContext) error {
	g.hdrCacheForever(ctx.w)
	g.sendFile("application/x-git-loose-object", ctx)
	return nil
}

func (g *GitHttpServer) getPackFile(ctx *requestContext) error {
	g.hdrCacheForever(ctx.w)
	g.sendFile("application/x-git-packed-objects", ctx)
	return nil
}

func (g *GitHttpServer) getIdxFile(ctx *requestContext) error {
	g.hdrCacheForever(ctx.w)
	g.sendFile("application/x-git-packed-objects-toc", ctx)
	return nil
}

func (g *GitHttpServer) getTextFile(ctx *requestContext) error {
	g.hdrNocache(ctx.w)
	g.sendFile("text/plain", ctx)
	return nil
}

func (g *GitHttpServer) serviceRpc(ctx *requestContext) error {
	w, r, rpc, dir := ctx.w, ctx.r, ctx.Rpc, ctx.Dir

	w.Header().Set("Content-Type", fmt.Sprintf("application/x-git-%s-result", rpc))
	w.WriteHeader(http.StatusOK)

	var body = r.Body
	defer body.Close()

	if r.Header.Get("Content-Encoding") == "gzip" {
		body, _ = gzip.NewReader(r.Body)
	}

	args := []string{rpc, "--stateless-rpc", dir}

	// deadline
	cmdCtx, cancel := context.WithTimeout(context.Background(), time.Duration(g.deadline)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, g.gitBinPath, args...)
	cmd.Dir = dir
	cmd.Stdin = body
	cmd.Stdout = w
	cmd.Stderr = w

	err := cmd.Run()
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (g *GitHttpServer) getInfoRefs(ctx *requestContext) error {
	w, r, dir := ctx.w, ctx.r, ctx.Dir
	serviceName := g.getServiceType(r)

	access := g.hasAccess(r, dir, serviceName, false)
	if access {
		args := []string{serviceName, "--stateless-rpc", "--advertise-refs", "."}
		refs, err := g.gitCommand(dir, args...)
		if err != nil {
			return err
		}

		g.hdrNocache(w)
		w.Header().Set("Content-Type", fmt.Sprintf("application/x-git-%s-advertisement", serviceName))
		w.WriteHeader(http.StatusOK)
		w.Write(g.packetWrite("# service=git-" + serviceName + "\n"))
		w.Write(g.packetFlush())
		w.Write([]byte(refs))
	} else {
		g.updateServerInfo(dir)
		g.hdrNocache(w)
		g.sendFile("text/plain; charset=utf-8", ctx)
	}
	return nil
}

func (g *GitHttpServer) packetFlush() []byte {
	return []byte("0000")
}

func (g *GitHttpServer) packetWrite(str string) []byte {
	s := strconv.FormatInt(int64(len(str)+4), 16)

	if len(s)%4 != 0 {
		s = strings.Repeat("0", 4-len(s)%4) + s
	}
	return []byte(s + str)
}

func (g *GitHttpServer) sendFile(contentType string, ctx *requestContext) {
	w, r := ctx.w, ctx.r
	reqFile := path.Join(ctx.Dir, ctx.File)

	f, err := os.Stat(reqFile)
	if os.IsNotExist(err) {
		g.httpRender(w, http.StatusNotFound, "not found")
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", f.Size()))
	w.Header().Set("Last-Modified", f.ModTime().Format(http.TimeFormat))
	http.ServeFile(w, r, reqFile)
}

func (g *GitHttpServer) hasAccess(r *http.Request, dir string, rpc string, checkContentType bool) bool {
	if checkContentType {
		if r.Header.Get("Content-Type") != fmt.Sprintf("application/x-git-%s-request", rpc) {
			return false
		}
	}

	if !(rpc == "upload-pack" || rpc == "receive-pack") {
		return false
	}
	if rpc == "receive-pack" {
		// return g.config.ReceivePack
		return true
	}
	if rpc == "upload-pack" {
		// return g.config.UploadPack
		return true
	}

	return g.getConfigSetting(rpc, dir)
}

func (g *GitHttpServer) getServiceType(r *http.Request) string {
	serviceType := r.FormValue("service")

	if s := strings.HasPrefix(serviceType, "git-"); !s {
		return ""
	}
	return strings.Replace(serviceType, "git-", "", 1)
}

func (g *GitHttpServer) getConfigSetting(serviceName string, dir string) bool {
	serviceName = strings.Replace(serviceName, "-", "", -1)
	setting, err := g.getGitConfig("http."+serviceName, dir)
	if err != nil {
		log.Printf("get git config was err: %v", err)
		return false
	}

	if serviceName == "uploadpack" {
		return setting != "false"
	}
	return setting == "true"
}

func (g *GitHttpServer) getGitConfig(configName string, dir string) (string, error) {
	args := []string{"config", configName}
	out, err := g.gitCommand(dir, args...)
	if err != nil {
		return "", err
	}
	return out, nil
}

func (g *GitHttpServer) updateServerInfo(dir string) (string, error) {
	args := []string{"update-server-info"}
	return g.gitCommand(dir, args...)
}

func (g *GitHttpServer) gitCommand(dir string, args ...string) (string, error) {
	cmd := exec.Command(g.gitBinPath, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return string(out), errors.WithStack(err)
}

func (g *GitHttpServer) hdrNocache(w http.ResponseWriter) {
	w.Header().Set("Expires", "Fri, 01 Jan 1980 00:00:00 GMT")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Cache-Control", "no-cache, max-age=0, must-revalidate")
}

func (g *GitHttpServer) hdrCacheForever(w http.ResponseWriter) {
	now := time.Now().Unix()
	expires := now + 31536000
	w.Header().Set("Date", fmt.Sprintf("%d", now))
	w.Header().Set("Expires", fmt.Sprintf("%d", expires))
	w.Header().Set("Cache-Control", "public, max-age=31536000")
}
