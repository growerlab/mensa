package mensa

import (
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/growerlab/mensa/mensa/common"
	"github.com/pkg/errors"
)

// TODO 平滑重启

const (
	DefaultIdleTimeout = 120  // 链接最大闲置时间
	DefaultDeadline    = 3600 // git 的默认执行时间，最长1小时
)

var AllowedCommandMap = map[string]string{
	"git-receive-pack":   "receive-pack",
	"git-upload-pack":    "upload-pack",
	"git-upload-archive": "upload-archive",
}

func RunGitSSHServer(listen, hostKey string, entryer Entryer) {
	gitServer := &GitSSHServer{
		entryer:     entryer,
		logger:      nil,
		gitUser:     "git",
		listen:      listen,
		hostKey:     hostKey,
		deadline:    DefaultDeadline,
		idleTimeout: DefaultIdleTimeout,
	}
	err := gitServer.Start()
	if err != nil {
		panic(err)
	}
}

// MultiServers multi instance
type GitSSHServer struct {
	entryer Entryer

	logger io.Writer

	srv *ssh.Server

	gitBinPath  string // bin git
	gitUser     string // default "git"
	listen      string // listen addr
	hostKey     string // host key
	deadline    int    // default 3600
	idleTimeout int    // default 120
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
func (g *GitSSHServer) Start() error {
	if err := g.validate(); err != nil {
		return err
	}
	if err := g.run(); err != nil {
		return err
	}
	return nil
}

func (g *GitSSHServer) handler(session ssh.Session) {
	var err error

	ctx, err := common.BuildContextFromSSH(session)
	if err != nil {
		log.Printf("%v\n", err)
		return
	}
	log.Println("git handler commands: ", ctx.RawCommands)

	err = g.entryer.Prep(ctx)
	if err != nil {
		g.entryer.Fail(err)
		return
	}

	service, ok := AllowedCommandMap[ctx.RawCommands[0]]
	if !ok {
		log.Printf("invalid service: %s\n", ctx.RawCommands[0])
		return
	}

	// deadline
	cmdCtx, cancel := context.WithTimeout(context.Background(), time.Duration(g.deadline)*time.Second)
	defer cancel()

	args := []string{service, ctx.RepoPath}
	cmd := exec.CommandContext(cmdCtx, g.gitBinPath, args...)
	cmd.Dir = ctx.RepoPath
	// TODO 这里使用pipe还是stdin，可能还有待测试
	cmd.Stdin = session
	cmd.Stdout = session
	err = cmd.Run()
	if err != nil {
		log.Printf("git was err on running: %v\n", err)
	}
}

func (g *GitSSHServer) validate() error {
	if g.gitUser == "" {
		return errors.New("git user is required")
	}
	if !strings.Contains(g.listen, ":") {
		return errors.New("invalid listen addr")
	}
	if _, err := os.Stat(g.hostKey); os.IsNotExist(err) {
		return errors.Errorf("%s is not exist", g.hostKey)
	}
	return nil
}

func (g *GitSSHServer) prepre() {
	log.SetOutput(g.logger)
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

	hostKeyOption := ssh.HostKeyFile(g.hostKey)

	defaultOption := func(server *ssh.Server) error {
		if g.idleTimeout == 0 {
			g.idleTimeout = DefaultIdleTimeout
		}
		server.IdleTimeout = time.Duration(g.idleTimeout) * time.Second
		server.MaxTimeout = 30 * time.Second
		server.Version = UA
		return nil
	}

	g.srv = &ssh.Server{Handler: g.handler}
	g.srv.SetOption(publicKeyHanderOption)
	g.srv.SetOption(passwordOption)
	g.srv.SetOption(hostKeyOption)
	g.srv.SetOption(defaultOption)
	g.srv.SetOption(ssh.NoPty())
	err := g.srv.ListenAndServe()
	return errors.WithStack(err)
}
