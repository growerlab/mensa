package middleware

import (
	"net/http"
	"strings"

	"github.com/growerlab/mensa/app/common"
)

type MiddlewareError string

func (m MiddlewareError) Error() string {
	return string(m)
}

type HandleResult struct {
	status       int
	lastErrorMsg strings.Builder
	lastError    error
}

func (h *HandleResult) HttpStatus() int {
	return h.status
}

// 当进入失败时，应返回http错误的信息
func (h *HandleResult) HttpStatusMessage() string {
	h.lastErrorMsg.WriteString(http.StatusText(h.status))
	h.lastErrorMsg.WriteString("\n----- Power by GrowerLab.net -----")
	return h.lastErrorMsg.String()
}

// 错误码
func (h *HandleResult) LastErr() error {
	return h.lastError
}

type HandleFunc func(*common.Context) (httpCode int, appendText string, err error)

type Middleware struct {
	funcs []HandleFunc
}

func (m *Middleware) Add(fn HandleFunc) {
	m.funcs = append(m.funcs, fn)
}

func (m *Middleware) Run(ctx *common.Context) *HandleResult {
	result := &HandleResult{}

	for _, fn := range m.funcs {
		statusCode, appendText, err := fn(ctx)
		if len(appendText) > 0 {
			result.lastErrorMsg.WriteString(appendText)
		}
		result.status = statusCode
		if err != nil {
			result.lastError = err
			return result
		}
	}
	return result
}

func (m *Middleware) Enter(ctx *common.Context) *HandleResult {
	return m.Run(ctx)
}
