package middleware

import (
	"log"
	"net/http"
	"os"

	"github.com/growerlab/backend/app/common/permission"
	"github.com/growerlab/mensa/app/common"
	"github.com/growerlab/mensa/app/service"
)

// Authenticate 鉴权
func Authenticate(ctx *common.Context) (httpCode int, appendText string, err error) {
	httpCode = http.StatusOK
	noAuth := os.Getenv("NOAUTH")
	if len(noAuth) > 0 {
		return
	}

	if err = checkPermission(ctx); err != nil {
		httpCode = http.StatusUnauthorized
		appendText = err.Error()
		log.Printf("%s, unauthorized: %v\n", ctx.Desc(), err)
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
	if ctx.IsReadAction() {
		var nsID int64
		var err error
		if !ctx.Operator.IsEmptyUser() {
			nsID, err = service.GetNamespaceByOperator(ctx.Operator)
			if err != nil {
				return err
			}
		}
		return permission.CheckCloneRepository(&nsID, repoID)
	} else {
		nsID, err := service.GetNamespaceByOperator(ctx.Operator)
		if err != nil {
			return err
		}
		return permission.CheckPushRepository(nsID, repoID)
	}
}
