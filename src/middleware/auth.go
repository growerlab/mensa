package middleware

import (
	"net/http"
	"os"

	"github.com/growerlab/mensa/src/common"
)

func Authenticate(ctx *common.Context) (httpCode int, err error) {
	httpCode = http.StatusOK
	noAuth := os.Getenv("NOAUTH")
	if len(noAuth) > 0 {
		return
	}

	if err = checkPermission(ctx); err != nil {
		return
	}

	return
}

// 检查是否有读取、推送权限
func checkPermission(ctx *common.Context) error {
	// ctx.RawCommands
	return nil
}
