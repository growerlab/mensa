package service

import (
	"github.com/go-redis/redis/v7"
	"github.com/growerlab/backend/app/common/errors"
	"github.com/growerlab/backend/app/model/db"
	selfDB "github.com/growerlab/mensa/app/db"
)

type getFunc func() (value string, err error)

type Cache struct {
	memDB *db.MemDBClient
}

func NewCache() *Cache {
	return &Cache{memDB: selfDB.MemDB}
}

func (c *Cache) GetOrSet(key, field string, getf getFunc) (string, error) {
	key = c.memDB.KeyMaker().Append(key).String()

	cmd := c.memDB.HGet(key, field)
	if cmd.Err() != redis.Nil {
		return cmd.Val(), nil
	} else {
		value, err := getf()
		if err != nil {
			return "", err
		}
		err = c.memDB.HSet(key, field, value).Err()
		if err != nil {
			return "", errors.Trace(err)
		}
		return value, nil
	}
}
