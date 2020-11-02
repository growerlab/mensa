package app

import (
	"log"
	"strings"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/growerlab/mensa/app/common"
	"github.com/growerlab/mensa/app/conf"
	"github.com/pkg/errors"
)

func NewGitSSHServer(cfg *conf.Config) *GitSSHServer {
	deadline := DefaultDeadline * time.Second
	idleTimeout := DefaultIdleTimeout * time.Second

	if cfg.Deadline > 0 {
		deadline = time.Duration(cfg.Deadline) * time.Second
	}
	if cfg.IdleTimeout > 0 {
		idleTimeout = time.Duration(cfg.IdleTimeout) * time.Second
	}

	gitServer := &GitSSHServer{
		gitUser: cfg.User,
		listen:  cfg.Listen,
		// hostKeys:    cfg.HostKeys,
		gitBinPath:  cfg.GitPath,
		deadline:    deadline,
		idleTimeout: idleTimeout,
	}
	return gitServer
}

type GitSSHServer struct {
	handler MiddlewareHandler

	srv         *ssh.Server
	gitBinPath  string        // bin git
	gitUser     string        // default "git"
	listen      string        // listen addr
	deadline    time.Duration // default 3600
	idleTimeout time.Duration // default 120
}

// Shutdown close all server and wait.
func (g *GitSSHServer) Shutdown() error {
	var err error
	if g.srv != nil {
		err = g.srv.Close()
	}
	return errors.WithStack(err)
}

// Start server
func (g *GitSSHServer) ListenAndServe(handler MiddlewareHandler) error {
	log.Printf("[ssh] git listen and serve: %v\n", g.listen)
	g.handler = handler

	if err := g.validate(); err != nil {
		return err
	}
	if err := g.run(); err != nil {
		return err
	}
	return nil
}

func (g *GitSSHServer) sessionHandler(session ssh.Session) {
	defer session.Close()

	ctx, err := common.BuildContextFromSSH(session)
	if err != nil {
		log.Printf("[ssh] %v\n", err)
		return
	}
	log.Println("[ssh] git handler commands: ", ctx.RawCommands, ctx.RepoDir)

	result := g.handler(ctx)
	if result.Err != nil {
		_, _ = session.Write([]byte(result.HttpMessage))
		return
	}

	service, ok := AllowedCommandMap[ctx.RawCommands[0]]
	if !ok {
		log.Printf("[ssh] invalid service: %s\n", ctx.RawCommands[0])
		return
	}

	args := []string{service, ctx.RepoDir}

	err = gitCommand(session, session, ctx.RepoDir, args, ctx.Env())
	if err != nil {
		log.Printf("[ssh] git was err on running: %v\n", err)
	}
}

func (g *GitSSHServer) validate() error {
	if g.gitUser == "" {
		return errors.New("git user is required")
	}
	if !strings.Contains(g.listen, ":") {
		return errors.New("invalid listen addr")
	}
	return nil
}

func (g *GitSSHServer) prepre() {
}

func (g *GitSSHServer) run() error {
	passwordOption := ssh.PasswordAuth(func(_ ssh.Context, _ string) bool {
		return false
	})

	publicKeyOption := ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
		if g.gitUser != ctx.User() {
			return false
		}
		return true
	})

	defaultOption := func(server *ssh.Server) error {
		server.IdleTimeout = g.idleTimeout
		server.MaxTimeout = g.deadline
		server.Version = UA
		return nil
	}

	g.srv = &ssh.Server{
		Handler: g.sessionHandler,
		Addr:    g.listen,
	}
	g.srv.SetOption(publicKeyOption)
	g.srv.SetOption(passwordOption)
	g.srv.SetOption(defaultOption)
	// for _, k := range g.hostKeys {
	// 	g.srv.SetOption(ssh.HostKeyFile(k))
	// }
	err := g.srv.ListenAndServe()
	return errors.WithStack(err)
}
