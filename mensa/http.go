package mensa

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/growerlab/mensa/mensa/common"
	"github.com/pkg/errors"
)

func RunGitHttpServer(addr, gitPath string, logger io.Writer, entryer Entryer) {
	server := &GitHttpServer{
		addr:    addr,
		entryer: entryer,
		gitPath: gitPath,
		logger:  logger,
	}
	err := server.Start()
	if err != nil {
		panic(err)
	}
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
	Do     func(*requestContext) error
	Rpc    string
}

type GitHttpServer struct {
	// 当有新的连接时，先执行该『关卡』
	entryer Entryer
	// 服务器的监听地址(eg. host:port)
	addr string
	// git bin path
	gitPath string
	// logger
	logger io.Writer

	// services
	services map[string]service
}

func (g *GitHttpServer) Start() error {
	if err := g.validate(); err != nil {
		return errors.WithStack(err)
	}

	g.prepre()

	if err := http.ListenAndServe(g.addr, g); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (g *GitHttpServer) validate() error {
	if g.entryer == nil {
		return errors.New("entryer is required")
	}
	if g.addr == "" {
		return errors.New("addr is required")
	}
	if !strings.Contains(g.addr, ":") {
		return errors.Errorf("addr is invalid: %s", g.addr)
	}
	return nil
}

func (g *GitHttpServer) prepre() {
	if g.logger != nil {
		log.SetPrefix("MENSA")
		log.SetOutput(g.logger)
	}

	g.services = map[string]service{
		"(.*?)/git-upload-pack$":                       service{"POST", g.serviceRpc, "upload-pack"},
		"(.*?)/git-receive-pack$":                      service{"POST", g.serviceRpc, "receive-pack"},
		"(.*?)/info/refs$":                             service{"GET", g.getInfoRefs, ""},
		"(.*?)/HEAD$":                                  service{"GET", g.getTextFile, ""},
		"(.*?)/objects/info/alternates$":               service{"GET", g.getTextFile, ""},
		"(.*?)/objects/info/http-alternates$":          service{"GET", g.getTextFile, ""},
		"(.*?)/objects/info/packs$":                    service{"GET", g.getInfoPacks, ""},
		"(.*?)/objects/info/[^/]*$":                    service{"GET", g.getTextFile, ""},
		"(.*?)/objects/[0-9a-f]{2}/[0-9a-f]{38}$":      service{"GET", g.getLooseObject, ""},
		"(.*?)/objects/pack/pack-[0-9a-f]{40}\\.pack$": service{"GET", g.getPackFile, ""},
		"(.*?)/objects/pack/pack-[0-9a-f]{40}\\.idx$":  service{"GET", g.getIdxFile, ""},
	}
}

func (g *GitHttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	begin := time.Now()
	preLog := fmt.Sprintf("IP: %s URL: %s [%s]", r.RemoteAddr, r.URL.String(), begin.Format(time.RFC3339))
	defer func() {
		end := time.Now()
		log.Printf("%s-[%s] TAKE: %s\n", preLog, end.Format(time.RFC3339), end.Sub(begin))
	}()

	ctx, err := common.BuildContextFromHTTP(r.URL)
	if err != nil {
		g.httpRender(w, http.StatusBadRequest, "bad request")
	}

	err = g.entryer.Prep(ctx)
	if err != nil {
		g.entryer.Fail(err)
		return
	}

	// git服务
	for match, service := range g.services {
		re, err := regexp.Compile(match)
		if err != nil {
			log.Print(err)
		}

		if m := re.FindStringSubmatch(r.URL.Path); m != nil {
			if service.Method != r.Method {
				g.httpRender(w, http.StatusMethodNotAllowed, "invalid method: "+r.Method)
				return
			}

			rpc := service.Rpc
			file := strings.Replace(r.URL.Path, m[1]+"/", "", 1)
			dir := ctx.RepoPath

			err = service.Do(&requestContext{w, r, rpc, dir, file})
			if err != nil {
				log.Printf("service.Do was err:%v\n", err)
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

	var body io.ReadCloser = r.Body
	defer body.Close()

	if r.Header.Get("Content-Encoding") == "gzip" {
		body, _ = gzip.NewReader(r.Body)
	}

	args := []string{rpc, "--stateless-rpc", dir}

	// TODO 这里是否应该使用 ContextCommand 来给命令一个超时时间？
	cmd := exec.Command(g.gitPath, args...)
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
		fmt.Println(args)
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

func (g *GitHttpServer) sendFile(content_type string, ctx *requestContext) {
	w, r := ctx.w, ctx.r
	reqFile := path.Join(ctx.Dir, ctx.File)

	f, err := os.Stat(reqFile)
	if os.IsNotExist(err) {
		g.httpRender(w, http.StatusNotFound, "not found")
		return
	}
	w.Header().Set("Content-Type", content_type)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", f.Size()))
	w.Header().Set("Last-Modified", f.ModTime().Format(http.TimeFormat))
	http.ServeFile(w, r, reqFile)
}

func (g *GitHttpServer) hasAccess(r *http.Request, dir string, rpc string, check_content_type bool) bool {
	if check_content_type {
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
	cmd := exec.Command(g.gitPath, args...)
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
