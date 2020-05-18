package app

import (
	"context"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/growerlab/mensa/app/common"
	"github.com/growerlab/mensa/app/conf"
	"github.com/pkg/errors"
)

// TODO 平滑重启

const (
	DefaultIdleTimeout = 120  // 链接最大闲置时间
	DefaultDeadline    = 3600 // git 的默认执行时间，最长1小时
)

const (
	GitReceivePack   = "git-receive-pack"
	GitUploadPack    = "git-upload-pack"
	GitUploadArchive = "git-upload-archive"

	ReceivePack   = "receive-pack"
	UploadPack    = "upload-pack"
	UploadArchive = "upload-archive"
)

var AllowedCommandMap = map[string]string{
	GitReceivePack:   ReceivePack,
	GitUploadPack:    UploadPack,
	GitUploadArchive: UploadArchive,
}

func NewGitSSHServer(cfg *conf.Config) *GitSSHServer {
	deadline := DefaultDeadline
	idleTimeout := DefaultIdleTimeout

	if cfg.Deadline > 0 {
		deadline = cfg.Deadline
	}
	if cfg.IdleTimeout > 0 {
		idleTimeout = cfg.IdleTimeout
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
	handler ServerHandler

	srv         *ssh.Server
	gitBinPath  string   // bin git
	gitUser     string   // default "git"
	listen      string   // listen addr
	hostKeys    []string // host keys
	deadline    int      // default 3600
	idleTimeout int      // default 120
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
func (g *GitSSHServer) ListenAndServe(handler ServerHandler) error {
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
	var err error
	defer func() {
		session.Close()
	}()

	ctx, err := common.BuildContextFromSSH(session)
	if err != nil {
		log.Printf("[ssh] %v\n", err)
		return
	}
	log.Println("[ssh] git handler commands: ", ctx.RawCommands, ctx.RepoDir)

	result := g.handler(ctx)
	if result != nil {
		_, _ = session.Write([]byte(result.HttpMessage))
		return
	}

	service, ok := AllowedCommandMap[ctx.RawCommands[0]]
	if !ok {
		log.Printf("[ssh] invalid service: %s\n", ctx.RawCommands[0])
		return
	}

	// deadline
	cmdCtx, cancel := context.WithTimeout(context.Background(), time.Duration(g.deadline)*time.Second)
	defer cancel()

	args := []string{service, ctx.RepoDir}
	cmd := exec.CommandContext(cmdCtx, g.gitBinPath, args...)
	cmd.Dir = ctx.RepoDir
	cmd.Stdin = session
	cmd.Stdout = session
	err = cmd.Run()
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

	publicKeyHanderOption := ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
		if g.gitUser != ctx.User() {
			return false
		}
		return true
	})

	defaultOption := func(server *ssh.Server) error {
		server.IdleTimeout = time.Duration(g.idleTimeout) * time.Second
		server.MaxTimeout = 30 * time.Second
		server.Version = UA
		return nil
	}

	g.srv = &ssh.Server{
		Handler: g.sessionHandler,
		Addr:    g.listen,
	}
	g.srv.SetOption(publicKeyHanderOption)
	g.srv.SetOption(passwordOption)
	g.srv.SetOption(defaultOption)
	// for _, k := range g.hostKeys {
	// 	g.srv.SetOption(ssh.HostKeyFile(k))
	// }
	err := g.srv.ListenAndServe()
	return errors.WithStack(err)
}
