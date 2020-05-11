package db

import (
	"fmt"
	"log"
	"runtime/debug"

	sq "github.com/Masterminds/squirrel"
	"github.com/growerlab/mensa/src/conf"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

var DB *sqlx.DB

func InitDatabase() error {
	var config = conf.GetConfig()
	var err error

	DB, err = sqlx.Connect("pgx", config.DBUrl)
	if err != nil {
		panic(err)
	}

	// pgsql placeholder
	sq.StatementBuilder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	return nil
}

func Transact(txFn func(*sqlx.Tx) error) (err error) {
	tx := DB.MustBegin()

	defer func() {
		if p := recover(); p != nil {
			log.Printf("%s: %s", p, debug.Stack())
			switch x := p.(type) {
			case error:
				err = x
			default:
				err = fmt.Errorf("%s", x)
			}
		}
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = errors.WithStack(tx.Commit())
	}()

	return txFn(tx)
}
