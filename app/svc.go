package app

import (
	"context"
	"net/http"
	"time"

	"github.com/growerlab/mensa/app/conf"
	"github.com/growerlab/mensa/svc/model"
	"github.com/growerlab/mensa/svc/service"
)

type SvcServer struct {
	server *http.Server
	config *conf.Config
}

func NewSvcServer(config *conf.Config) *SvcServer {

	model.ReposDir = config.GitRepoDir

	deadline := time.Duration(config.Deadline) * time.Second
	idleTimeout := time.Duration(config.IdleTimeout) * time.Second

	engine := service.BuildEngine(config)
	server := &http.Server{
		Handler:      engine,
		Addr:         config.SvcAddr,
		WriteTimeout: deadline,
		IdleTimeout:  idleTimeout,
	}

	return &SvcServer{
		config: config,
		server: server,
	}
}

func (s *SvcServer) ListenAndServe(handler MiddlewareHandler) error {
	return s.server.ListenAndServe()
}

// 停止服务
func (s *SvcServer) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}
