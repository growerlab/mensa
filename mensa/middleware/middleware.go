package middleware

import (
	"net/http"

	"github.com/growerlab/mensa/mensa/common"
	"github.com/pkg/errors"
)

type MiddlewareFunc func(*common.Context) (httpCode int, err error)

type Middleware struct {
	midFuncs []MiddlewareFunc

	lastErr        error
	lastStatusCode int
}

func (m *Middleware) Add(fn MiddlewareFunc) {
	m.midFuncs = append(m.midFuncs, fn)
}

func (m *Middleware) Run(ctx *common.Context) error {
	for _, fn := range m.midFuncs {
		m.lastStatusCode, m.lastErr = fn(ctx)
		if m.lastErr != nil {
			return errors.WithStack(m.lastErr)
		}
	}
	return nil
}

func (m *Middleware) Prep(ctx *common.Context) error {
	err := m.Run(ctx)
	return err
}

func (m *Middleware) HttpStatus() int {
	return m.lastStatusCode
}

func (m *Middleware) HttpStatusMessage() string {
	if m.lastStatusCode == 0 {
		m.lastStatusCode = http.StatusOK
	}
	return http.StatusText(m.lastStatusCode)
}
