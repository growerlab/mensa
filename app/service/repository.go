package service

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"

	repoModel "github.com/growerlab/backend/app/model/repository"
	"github.com/growerlab/mensa/app/db"
)

func RepositoryID(repoOwner, repoName string) (int64, error) {
	repoOwnerNS, err := GetUserNamespaceByUsername(repoOwner)
	if err != nil {
		return 0, err
	}

	// 仓库的公开状态可能变动，所以这里不缓存
	repoIDRaw, err := NewCache().GetOrSet(
		db.BaseKeyBuilder("repository", "id").String(),
		strings.Join([]string{repoOwner, repoName}, ":"),
		func() (value string, err error) {
			repo, err := repoModel.GetRepositoryByNsWithPath(db.DB, repoOwnerNS, repoName)
			if err != nil {
				return "", err
			}
			if repo == nil {
				return "", errors.Errorf("not found repo: %s/%s", repoOwner, repoName)
			}
			return strconv.FormatInt(repo.ID, 10), nil
		})
	if err != nil {
		return 0, errors.WithStack(err)
	}

	repoID, err := strconv.ParseInt(repoIDRaw, 10, 64)

	return repoID, errors.WithStack(err)
}
