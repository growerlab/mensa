package common

import (
	"github.com/growerlab/backend/app/common/permission"
	"github.com/growerlab/mensa/app/db"
)

func InitPermission() error {
	err := permission.InitPermissionHub(db.DB, db.PermissionDB)
	return err
}
