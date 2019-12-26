package middleware

import (
	"net/http"
	"os"

	"github.com/growerlab/mensa/mensa/common"
)

func Authenticate(ctx *common.Context) (httpCode int, err error) {
	noauth := os.Getenv("NOAUTH")
	if len(noauth) > 0 {
		return http.StatusOK, nil
	}

	return http.StatusOK, nil
}
