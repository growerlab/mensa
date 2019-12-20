package mensa

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"

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

type requestHandler struct {
	w    http.ResponseWriter
	r    *http.Request
	Rpc  string
	Dir  string
	File string
}

type service struct {
	Method  string
	Handler func(requestHandler)
	Rpc     string
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
	ctx, err := common.BuildContextFromHTTP(r.URL)
	if err != nil {
		g.httpRender(w, http.StatusBadRequest, "bad request")
	}

	err = g.entryer.Prep(ctx)
	if err != nil {
		g.entryer.Fail(err)
		return
	}

	// TODO git服务
}

func (g *GitHttpServer) httpRender(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	_, _ = w.Write([]byte(message))
}

func (g *GitHttpServer) serviceRpc(hr requestHandler) {
	w, r, rpc, dir := hr.w, hr.r, hr.Rpc, hr.Dir

	w.Header().Set("Content-Type", fmt.Sprintf("application/x-git-%s-result", rpc))
	w.WriteHeader(http.StatusOK)

	args := []string{rpc, "--stateless-rpc", dir}
	cmd := exec.Command(g.gitPath, args...)
	cmd.Dir = dir
	in, err := cmd.StdinPipe()
	if err != nil {
		log.Print(err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Print(err)
	}

	err = cmd.Start()
	if err != nil {
		log.Print(err)
	}

	var reader io.ReadCloser
	switch r.Header.Get("Content-Encoding") {
	case "gzip":
		reader, _ = gzip.NewReader(r.Body)
		defer reader.Close()
	default:
		reader = r.Body
	}
	io.Copy(in, reader)
	in.Close()
	io.Copy(w, stdout)
	cmd.Wait()
}
