// KeyDB / Redis 配置

package db

import (
	"net"
	"strconv"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/growerlab/backend/app/model/db"
	"github.com/growerlab/mensa/app/conf"
	"github.com/pkg/errors"
)

var MemDB *redis.Client

func InitMemDB() error {
	var config = conf.GetConfig().Redis
	MemDB = newPool(config, 0)

	// Test
	reply, err := MemDB.Ping().Result()
	if err != nil || reply != "PONG" {
		return errors.New("memdb not ready")
	}
	return err
}

func newPool(cfg *conf.Redis, db int) *redis.Client {
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	idleTimeout := time.Duration(cfg.IdleTimeout) * time.Second

	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		DB:           db,
		PoolSize:     cfg.MaxActive,
		MinIdleConns: cfg.MaxIdle,
		IdleTimeout:  idleTimeout,
	})
	return client
}

func BaseKeyBuilder(s ...string) *db.KeyBuilder {
	return db.NewKeyBuilder(conf.GetConfig().Redis.Namespace).Append(s...)
}
