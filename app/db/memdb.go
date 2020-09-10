// KeyDB / Redis 配置

package db

import (
	"github.com/growerlab/backend/app/model/db"
	"github.com/growerlab/backend/app/utils/conf"
	selfConf "github.com/growerlab/mensa/app/conf"
)

var MemDB *db.MemDBClient
var PermissionDB *db.MemDBClient

func InitMemDB() (err error) {
	c := selfConf.GetConfig().Redis

	redisConf := &conf.Redis{
		Host:        c.Host,
		Port:        c.Port,
		Namespace:   c.Namespace,
		MaxIdle:     c.MaxIdle,
		MaxActive:   c.MaxActive,
		IdleTimeout: c.IdleTimeout,
	}

	MemDB, err = db.DoInitMemDB(redisConf, 0)
	if err != nil {
		return err
	}

	redisConf.Namespace = c.PermissionNamespace

	PermissionDB, err = db.DoInitMemDB(redisConf, 0)
	if err != nil {
		return err
	}
	return nil
}
