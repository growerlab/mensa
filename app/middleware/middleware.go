package middleware

import (
	"net/http"
	"strings"

	"github.com/growerlab/mensa/app/common"
)

type HandleFunc func(*common.Context) (httpCode int, appendText string, err error)

type Middleware struct {
	funcs []HandleFunc

	lastErr        error
	lastStatusCode int
	lastAppendText strings.Builder
}

func (m *Middleware) Add(fn HandleFunc) {
	m.funcs = append(m.funcs, fn)
}

func (m *Middleware) Run(ctx *common.Context) error {
	for _, fn := range m.funcs {
		statusCode, appendText, err := fn(ctx)
		if len(appendText) > 0 {
			m.lastAppendText.WriteString(appendText)
		}
		m.lastStatusCode = statusCode
		if err != nil {
			m.lastErr = err
			return m.lastErr
		}
	}
	return nil
}

func (m *Middleware) Enter(ctx *common.Context) error {
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
	if m.lastAppendText.Len() > 0 {
		sb.WriteString(m.lastAppendText.String())
	}
	sb.WriteString("\n----- Power by GrowerLab.net -----")
	return sb.String()
}

func (m *Middleware) LastErr() error {
	return m.lastErr
}
