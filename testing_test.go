package logx

import (
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/monstercat/logx/logxhost"
)


func getPostgresConnection(url string) (*sqlx.DB, error) {
	connStr, err := pq.ParseURL(url)
	if err != nil {
		return nil, errors.Wrapf(err, "error with postgres url '%s'", url)
	}
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, errors.Wrap(err, "error opening postgres")
	}
	return db, nil
}

func DefaultTestPostgres() *sqlx.DB {
	v := os.Getenv("TEST_PG")
	if v == "" {
		panic("environment variable TEST_PG not set")
	}
	db, err := getPostgresConnection(v)
	if err != nil {
		panic(err)
	}
	return db
}
