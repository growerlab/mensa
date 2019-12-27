package mensa

import (
	"io"
	"log"
	"os"

	"github.com/growerlab/mensa/mensa/conf"
	"github.com/growerlab/mensa/mensa/middleware"
)

// var mids *middleware.Middleware
var manager *Manager
var logger io.Writer = os.Stdout

func initialize() {
	// 初始化日志输出
	log.SetPrefix("[MENSA] ")
	log.SetOutput(logger)

	// 初始化依赖顺序的「初始化」
	startInit(conf.LoadConfig)
}

func startInit(fn func() error) {
	if err := fn(); err != nil {
		panic(err)
	}
}

func Run() {
	initialize()

	// 初始化中间件
	entry := new(middleware.Middleware)
	entry.Add(middleware.Authenticate)

	// 初始化管理器
	manager = NewManager(entry)
	manager.RegisterServer(NewGitHttpServer(conf.GetConfig()))
	manager.RegisterServer(NewGitSSHServer(conf.GetConfig()))
	manager.Run()
}
