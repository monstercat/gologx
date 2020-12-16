package logxhost

import (
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	dbutil "github.com/monstercat/golib/db"
	pgUtils "github.com/monstercat/golib/db/postgres"
	"github.com/monstercat/gologx/utils"
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
	ColsLogs     = dbutil.GetColumnsList(&Log{}, "")
	SortConfig   = map[string]string{"logTime": "log_time", "created": "created", "logType": "log_type", "service": "service", "machine": "machine"}
	SelectLogQry = psql.Select(ColsLogs...).From(ViewLog)
)

func SelectLogs(db sqlx.Queryer, q squirrel.SelectBuilder) ([]*Log, error) {
	var l []*Log
	if err := dbutil.Select(db, &l, q); err != nil {
		return nil, err
	}
	return l, nil
}

func GetLog(db sqlx.Queryer, id string) (*Log, error) {
	var q = psql.Select(ColsLogs...).From(ViewLog).Where("id = '" + id + "'")

	var l Log
	if err := dbutil.Get(db, &l, q); err != nil {
		return nil, err
	}

	return &l, nil
}

func GetLogSearchQuery(service string, matcher string, limit int, datebefore string, dateafter string, orderby string) squirrel.SelectBuilder {
	q := SelectLogQry.Where("service ILIKE '" + service + "'")

	if matcher != "" {
		q = q.Where(squirrel.Or{
			squirrel.ILike{"machine": "'" + matcher + "'"},
			squirrel.ILike{"message": "'" + matcher + "'"},
		})
	}

	if datebefore != "" {
		q = q.Where("log_time < '" + datebefore + "'")
	}

	if dateafter != "" {
		q = q.Where("log_time >= '" + dateafter + "'")
	}

	if orderby != "" {
		orderByArr := utils.ParseFlagStrToArray(orderby, ",")

		pgUtils.ApplySortStrings(orderByArr, &q, SortConfig)
	}

	q = q.Limit(uint64(limit))

	return q
}
