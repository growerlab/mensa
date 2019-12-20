package middleware

import (
	"github.com/growerlab/mensa/mensa/common"
)

type MiddlewareFunc func(*common.Context) error

type Middleware struct {
	midFuncs []MiddlewareFunc
}

func (m *Middleware) Add(fn MiddlewareFunc) {
	m.midFuncs = append(m.midFuncs, fn)
}

func (m *Middleware) Run(ctx *common.Context) error {
	for _, fn := range m.midFuncs {
		err := fn(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
