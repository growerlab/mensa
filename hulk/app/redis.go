package app

import (
	"net"
	"strconv"
	"time"

	"github.com/go-redis/redis/v7"
)

var RedisClient *redis.Client

func InitRedis() error {
	addr := net.JoinHostPort(Conf.Redis.Host, strconv.Itoa(Conf.Redis.Port))
	idleTimeout := time.Duration(Conf.Redis.IdleTimeout) * time.Second
	RedisClient = redis.NewClient(&redis.Options{
		Addr:         addr,
		DB:           0,
		PoolSize:     Conf.Redis.MaxActive,
		MinIdleConns: Conf.Redis.MaxIdle,
		IdleTimeout:  idleTimeout,
	})
	return nil
}
