package middleware

import (
	"log"
	"net/http"
	"os"

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
	if ctx.IsRead() {
		var userID *int64
		if !ctx.Operator.IsEmptyUser() {
			uid, err := service.GetNamespaceByOperator(ctx.Operator)
			if err != nil {
				return err
			}
			if uid > 0 {
				userID = &uid
			}
		}
		return permission.CheckCloneRepository(userID, repoID)
	} else {
		userID, err := service.GetNamespaceByOperator(ctx.Operator)
		if err != nil {
			return err
		}
		return permission.CheckPushRepository(userID, repoID)
	}
}
