package mensa

import (
	"errors"
	"sync"

	"github.com/growerlab/mensa/mensa/common"
)

type Result struct {
	HttpCode    int
	HttpMessage string
	Err         error
}

type ServerHandler func(ctx *common.Context) *Result

type Server interface {
	// 启动并监听服务
	// 	当有新的链接时，将调用cb方法
	ListenAndServe(cb ServerHandler) error
	// 停止服务
	Shutdown() error
}

type Manager struct {
	servers []Server

	entry Entryer
}

func NewManager(entry Entryer) *Manager {
	return &Manager{
		entry: entry,
	}
}

func (m *Manager) RegisterServer(srv Server) {
	m.servers = append(m.servers, srv)
}

// 允许server 并 等待
func (m *Manager) Run() {
	var wg sync.WaitGroup
	for _, s := range m.servers {
		wg.Add(1)
		go func(srv Server) {
			defer wg.Done()
			err := srv.ListenAndServe(m.ServerHandler)
			if err != nil {
				panic(err)
			}
		}(s)
	}
	wg.Wait()
}

func (m *Manager) ServerHandler(ctx *common.Context) *Result {
	if m.entry != nil {
		err := m.entry.Prep(ctx)
		if err != nil {
			return &Result{
				HttpCode:    m.entry.HttpStatus(),
				HttpMessage: m.entry.HttpStatusMessage(),
				Err:         errors.New(m.entry.HttpStatusMessage()),
			}
		}
	}
	return nil
}