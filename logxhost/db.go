package logxhost

import (
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	errs "github.com/pkg/errors"
)

var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

const (
	TableService = "service"
	TableLog     = "log"
	ViewLog      = "log_view"
)

var (
	ErrInvalidId = errors.New("invalid id")
)

func GetPostgresConnection(url string) (*sqlx.DB, error) {
	connStr, err := pq.ParseURL(url)
	if err != nil {
		return nil, errs.Wrapf(err, "error with postgres url '%s'", url)
	}
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, errs.Wrap(err, "error opening postgres")
	}
	return db, nil
}
