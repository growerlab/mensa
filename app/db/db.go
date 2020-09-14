package db

import (
	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/growerlab/backend/app/model/db"
	"github.com/growerlab/mensa/app/conf"
)

var DB *db.DBQuery

func InitDatabase() error {
	var err error
	var conf = conf.GetConfig()
	DB, err = db.DoInitDatabase(conf.DBUrl, conf.Debug)
	if err != nil {
		return err
	}

	db.DB, err = db.DoInitDatabase(conf.DBUrl, conf.Debug)
	return err
}
