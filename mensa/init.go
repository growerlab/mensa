package mensa

import "github.com/growerlab/mensa/mensa/middleware"

var mids *middleware.Middleware

func init() {
	mids = new(middleware.Middleware)
	mids.Add(middleware.Authenticate)
}

func Run() error {
	go RunGitHttpServer(":8080", "git", nil, nil)
	return nil
}
