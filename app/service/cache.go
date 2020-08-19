package service

import (
	"github.com/growerlab/backend/app/common/errors"
	"github.com/growerlab/backend/app/model/db"
)

type setFunc func() (value string, err error)

type Cache struct {
	memDB *db.MemDBClient
}

func NewCache() *Cache {
	return &Cache{memDB: db.MemDB}
}

func (c *Cache) GetOrSet(key, field string, getFunc setFunc) (string, error) {
	key = c.memDB.KeyBuilder.PartMaker().Append(key).String()

	if c.memDB.HExists(key, field).Val() {
		return c.memDB.HGet(key, field).Val(), nil
	} else {
		value, err := getFunc()
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
