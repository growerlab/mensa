package service

import (
	"github.com/go-redis/redis/v7"
	"github.com/growerlab/backend/app/common/errors"
	"github.com/growerlab/mensa/app/db"
)

type setFunc func() (value string, err error)

type Cache struct {
	c *redis.Client
}

func NewCache() *Cache {
	return &Cache{c: db.MemDB}
}

func (c *Cache) GetOrSet(key, field string, getFunc setFunc) (string, error) {
	if c.c.HExists(key, field).Val() {
		return c.c.HGet(key, field).Val(), nil
	} else {
		value, err := getFunc()
		if err != nil {
			return "", err
		}
		err = c.c.HSet(key, field, value).Err()
		if err != nil {
			return "", errors.Trace(err)
		}
		return value, nil
	}
}
