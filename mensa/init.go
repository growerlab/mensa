package mensa

import (
	"io"
	"log"
	"os"

	"github.com/growerlab/mensa/mensa/middleware"
)

var mids *middleware.Middleware
var logger io.Writer = os.Stdout

func init() {
	mids = new(middleware.Middleware)
	mids.Add(middleware.Authenticate)

	log.SetPrefix("MENSA")
	log.SetOutput(logger)
}

func Run() error {
	go RunGitHttpServer(":8080", "git", nil, nil)
	go RunGitSSHServer(":8022", "~/.ssh/id_rsa", nil)
	select {}
}
