package main

import (
	"github.com/gliderlabs/ssh"
)

// MultiServers multi instance
type MultiServers struct {
	srvs []*ssh.Server
}

// Shutdown close all server and wait.
func (ms *MultiServers) Shutdown() error {

	return nil
}

// Start start all servers
func (ms *MultiServers) Start() error {

	return nil
}
