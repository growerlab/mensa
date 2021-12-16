package app

import (
	"compress/gzip"
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/growerlab/mensa/app/common"
	"github.com/growerlab/mensa/app/conf"
	"github.com/pkg/errors"
)

func NewGitHttpServer(cfg *conf.Config) *GitHttpServer {
	deadline := DefaultDeadline * time.Second
	idleTimeout := DefaultIdleTimeout * time.Second

	if cfg.Deadline > 0 {
		deadline = time.Duration(cfg.Deadline) * time.Second
	}
	if cfg.IdleTimeout > 0 {
		idleTimeout = time.Duration(cfg.IdleTimeout) * time.Second
	}

	server := &GitHttpServer{
		listen:      cfg.HttpListen,
		gitBinPath:  cfg.GitPath,
		deadline:    deadline,
		idleTimeout: idleTimeout,
	}

	engine := gin.Default()
	engine.Use(server.handlerBuildRequestContext)
	engine.GET("/:path/:repo_name/info/refs", server.handlerGetInfoRefs)
	engine.POST("/:path/:repo_name/:rpc", server.handlerGitPack)

	server.server = &http.Server{
		Handler:      engine,
		Addr:         cfg.HttpListen,
		WriteTimeout: deadline,
		IdleTimeout:  idleTimeout,
	}
	return server
}

type requestContext struct {
	c   *gin.Context
	w   http.ResponseWriter
	r   *http.Request
	Rpc string
	Dir string
}

type GitHttpServer struct {
	// engine for http git
	server *http.Server
	// 服务器的监听地址(eg. host:port)
	listen string
	// git bin path
	gitBinPath string
	// logger io.Writer
	// 最长执行时间
	deadline time.Duration
	// 限制最大时间
	idleTimeout time.Duration

	MiddlewareHandler MiddlewareHandler
}

func (g *GitHttpServer) ListenAndServe(handler MiddlewareHandler) error {
	log.Printf("[http] git listen and serve: %v\n", g.listen)

	if err := g.validate(); err != nil {
		return err
	}

	g.MiddlewareHandler = handler

	if err := g.server.ListenAndServe(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (g *GitHttpServer) handlerBuildRequestContext(c *gin.Context) {
	r := c.Request
	w := c.Writer
	// file := r.URL.Path
	_, _, repoDir, err := common.BuildRepoInfoByPath(r.URL.Path)
	if err != nil {
		log.Printf("build repo info was err: %+v\n", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	rpc := g.getServiceType(c)

	req := &requestContext{
		c:   c,
		w:   w,
		r:   r,
		Rpc: rpc,
		Dir: repoDir,
	}
	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), "request_context", req))
}

func (g *GitHttpServer) handlerGitPack(c *gin.Context) {
	reqContext, ok := c.Request.Context().Value("request_context").(*requestContext)
	if !ok {
		log.Println("handlerGitPack: 'request_context' must exist in context")
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}

	err := g.serviceRpc(reqContext)
	if err != nil {
		log.Printf("git rpc err: %+v\n", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
}

func (g *GitHttpServer) handlerGetInfoRefs(c *gin.Context) {
	reqContext, ok := c.Request.Context().Value("request_context").(*requestContext)
	if !ok {
		log.Println("handlerGetInfoRefs: 'request_context' must exist in context")
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}

	err := g.getInfoRefs(reqContext)
	if err != nil {
		log.Printf("get info refs was err: %v\n", err)
		return
	}
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

// 平滑重启
func (g *GitHttpServer) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return g.server.Shutdown(ctx)
}

func (g *GitHttpServer) runMiddlewares(ctx *common.Context) error {
	result := g.MiddlewareHandler(ctx)
	if result.HttpCode > http.StatusCreated {
		ctx.Resp.WriteHeader(result.HttpCode)
		ctx.Resp.Header().Set("WWW-Authenticate", "Basic") // fmt.Sprintf("Basic realm=%s charset=UTF-8"))
	}

	if result.Err != nil {
		log.Printf("[http] middleware err: %+v \nresult:%d %s\n", result.Err, result.HttpCode, result.HttpMessage)
	}
	return result.Err
}

func (g *GitHttpServer) serviceRpc(ctx *requestContext) error {
	w, r, rpc, dir := ctx.w, ctx.r, ctx.Rpc, ctx.Dir

	var body = r.Body
	defer body.Close()

	commonCtx, err := common.BuildContextFromHTTP(ctx.w, ctx.r)
	if err != nil {
		return errors.WithStack(err)
	}

	w.Header().Set("Content-Type", fmt.Sprintf("application/x-git-%s-result", rpc))

	if r.Header.Get("Content-Encoding") == "gzip" {
		body, _ = gzip.NewReader(r.Body)
	}

	args := []string{rpc, "--stateless-rpc", "."}
	if rpc == ReceivePack {
		for _, op := range GitReceivePackOptions {
			args = append(args, op.Name, op.Args)
		}
	}

	err = gitCommand(body, w, dir, args, commonCtx.Env())
	return errors.WithStack(err)
}

func (g *GitHttpServer) getInfoRefs(ctx *requestContext) error {
	w, r, rpc, dir := ctx.w, ctx.r, ctx.Rpc, ctx.Dir

	access := g.hasAccess(r, dir, rpc, false)
	if access {
		commonCtx, err := common.BuildContextFromHTTP(ctx.w, ctx.r)
		if err != nil {
			return errors.WithStack(err)
		}
		err = g.runMiddlewares(commonCtx)
		if err != nil {
			return err
		}

		g.hdrNocache(w)
		w.Header().Set("Content-Type", fmt.Sprintf("application/x-git-%s-advertisement", rpc))
		_, _ = w.Write(g.packetWrite("# service=git-" + rpc + "\n"))
		_, _ = w.Write(g.packetFlush())

		args := []string{rpc, "--stateless-rpc", "--advertise-refs", "."}
		err = gitCommand(r.Body, w, dir, args, commonCtx.Env())
		if err != nil {
			return err
		}
	} else {
		// g.updateServerInfo(dir)
		// g.hdrNocache(w)
		log.Printf("can't access %s %s\n", dir, rpc)
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

// hasAccess 是否有访问权限
// 这里之后可能要改成从数据库、权限验证中心来确认
func (g *GitHttpServer) hasAccess(r *http.Request, dir string, rpc string, checkContentType bool) bool {
	if checkContentType {
		if r.Header.Get("Content-Type") != fmt.Sprintf("application/x-git-%s-request", rpc) {
			return false
		}
	}

	if !(rpc == UploadPack || rpc == ReceivePack) {
		return false
	}
	if rpc == ReceivePack {
		// return g.config.ReceivePack
		return true
	}
	if rpc == UploadPack {
		// return g.config.UploadPack
		return true
	}

	return true
}

func (g *GitHttpServer) getServiceType(c *gin.Context) string {
	serviceType := c.Request.FormValue("service")
	if len(serviceType) == 0 {
		serviceType = c.Param("rpc")
	}

	if s := strings.HasPrefix(serviceType, "git-"); !s {
		return ""
	}
	return strings.Replace(serviceType, "git-", "", 1)
}

func (g *GitHttpServer) getGitConfig(configName string, dir string) (string, error) {
	var args = []string{"config", configName}
	var out strings.Builder
	err := gitCommand(nil, &out, dir, args, nil)
	return out.String(), errors.WithStack(err)
}

// func (g *GitHttpServer) updateServerInfo(dir string) (string, error) {
// 	var args = []string{"update-server-info"}
// 	var out strings.Builder
// 	err := gitCommand(nil, &out, dir, args)
// 	return out.String(), errors.WithStack(err)
// }

func (g *GitHttpServer) hdrNocache(w http.ResponseWriter) {
	w.Header().Set("Expires", "Fri, 01 Jan 1980 00:00:00 GMT")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Cache-Control", "no-cache, max-age=0, must-revalidate")
}
