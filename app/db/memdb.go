// KeyDB / Redis 配置

package db

import (
	"net"
	"strconv"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/growerlab/backend/app/model/db"
	dbModel "github.com/growerlab/backend/app/model/db"
	"github.com/growerlab/mensa/app/conf"
	"github.com/pkg/errors"
)

var MemDB *db.MemDBClient
var PermissionDB *db.MemDBClient

func InitMemDB() error {
	var config = conf.GetConfig().Redis
	MemDB = newPool(config, config.Namespace, 0)
	PermissionDB = newPool(config, config.PermissionNamespace, 0)

	// Test
	reply, err := MemDB.Ping().Result()
	if err != nil || reply != "PONG" {
		return errors.New("memdb not ready")
	}
	return err
}

func newPool(cfg *conf.Redis, namespace string, db int) *dbModel.MemDBClient {
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	idleTimeout := time.Duration(cfg.IdleTimeout) * time.Second

	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		DB:           db,
		PoolSize:     cfg.MaxActive,
		MinIdleConns: cfg.MaxIdle,
		IdleTimeout:  idleTimeout,
	})

	memDB := &dbModel.MemDBClient{
		client,
		dbModel.NewKeyBuilder(namespace),
	}
	return memDB
}
