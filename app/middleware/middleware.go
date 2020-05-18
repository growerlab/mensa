package middleware

import (
	"net/http"
	"strings"

	"github.com/growerlab/mensa/app/common"
)

type MiddlewareFunc func(*common.Context) (httpCode int, appendText string, err error)

type Middleware struct {
	midFuncs []MiddlewareFunc

	lastErr        error
	lastStatusCode int
	lastAppendText string
}

func (m *Middleware) Add(fn MiddlewareFunc) {
	m.midFuncs = append(m.midFuncs, fn)
}

func (m *Middleware) Run(ctx *common.Context) error {
	for _, fn := range m.midFuncs {
		m.lastStatusCode, m.lastAppendText, m.lastErr = fn(ctx)
		if m.lastErr != nil {
			return m.lastErr
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

	var sb strings.Builder
	sb.WriteString(http.StatusText(m.lastStatusCode))
	if len(m.lastAppendText) > 0 {
		sb.WriteString(" ")
		sb.WriteString(m.lastAppendText)
	}
	return sb.String()
}

func (m *Middleware) LastErr() error {
	return m.lastErr
}
