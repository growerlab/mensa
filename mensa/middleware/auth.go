package middleware

import (
	"net/http"

	"github.com/growerlab/mensa/mensa/common"
)

func Authenticate(ctx *common.Context) (httpCode int, err error) {

	return http.StatusOK, nil
}
