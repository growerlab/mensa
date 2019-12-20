package main

import (
	"github.com/gliderlabs/ssh"
)

// ServerVars config
type ServerVars struct {
	Listen      string   `json:"Listen"`
	HostKeys    []string `json:"HostKeys"`
	Deadline    int      `json:"Deadline,omitempty"`    // default 3600
	IdleTimeout int      `json:"IdleTimeout,omitempty"` // default 120
}

// Mensa-1.0

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
