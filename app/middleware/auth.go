package middleware

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/growerlab/backend/app/common/permission"
	"github.com/growerlab/mensa/app/common"
	"github.com/growerlab/mensa/app/service"
)

func Authenticate(ctx *common.Context) (httpCode int, appendText string, err error) {
	httpCode = http.StatusOK
	noAuth := os.Getenv("NOAUTH")
	if len(noAuth) > 0 {
		return
	}

	if err = checkPermission(ctx); err != nil {
		httpCode = http.StatusUnauthorized
		var sb strings.Builder
		sb.Write([]byte("\n"))
		sb.WriteString("----- Power by GrowerLab.net -----")
		appendText = sb.String()
		log.Printf("unauthorized: %v\n", err)
		return
	}
	return
}

// 检查是否有读取、推送权限
//	公共项目：可读、只有项目成员可写
// 	私有项目：项目成员可读/写
//
func checkPermission(ctx *common.Context) error {
	repoID, err := service.RepositoryID(ctx.RepoOwner, ctx.RepoName)
	if err != nil {
		return err
	}
	if ctx.IsRead() {
		return permission.CheckCloneRepository(nil, repoID)
	} else {
		userID, err := service.GetNamespaceByOperator(ctx.Operator)
		if err != nil {
			return err
		}
		return permission.CheckPushRepository(userID, repoID)
	}
}
