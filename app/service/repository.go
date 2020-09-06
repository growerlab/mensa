package service

import (
	"strconv"

	dbModel "github.com/growerlab/backend/app/model/db"
	repoModel "github.com/growerlab/backend/app/model/repository"
	"github.com/growerlab/mensa/app/db"
	"github.com/pkg/errors"
)

func RepositoryID(repoOwner, repoName string) (int64, error) {
	repoOwnerNS, err := GetUserNamespaceByUsername(repoOwner)
	if err != nil {
		return 0, err
	}

	key := dbModel.MemDB.KeyMaker().Append("repository", "id", "namespace").String()
	field := dbModel.MemDB.KeyMakerNoNS().Append(repoOwner, repoName).String()

	// 仓库的公开状态可能变动，所以这里仅缓存仓库id
	repoIDRaw, err := NewCache().GetOrSet(
		key,
		field,
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
		return 0, err
	}

	repoID, err := strconv.ParseInt(repoIDRaw, 10, 64)
	return repoID, errors.WithStack(err)
}
