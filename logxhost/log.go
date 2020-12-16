package logxhost

import (
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	dbutil "github.com/monstercat/golib/db"
)

type Log struct {
	Id      string `setmap:"ignore"`
	Machine string
	Service string
	Context string
	Message string
	LogType string    `db:"log_type"`
	LogTime time.Time `db:"log_time"`
	Created time.Time
}

var (
	ColsLogs = dbutil.GetColumnsList(&Log{}, "")
)

func BuildSelectLogs() squirrel.SelectBuilder {
	return psql.Select(ColsLogs...).From(TableLogs)
}

func SelectLogs(db sqlx.Queryer, q squirrel.SelectBuilder) ([]*Log, error) {
	var l []*Log
	if err := dbutil.Select(db, &l, q); err != nil {
		return nil, err
	}
	return l, nil
}

func GetLog(db sqlx.Queryer, id string) (*Log, error) {
	var q = psql.Select(ColsLogs...).From(TableLogs).Where("id = '" + id + "'")

	var l Log
	if err := dbutil.Get(db, &l, q); err != nil {
		return nil, err
	}

	return &l, nil
}
