package logxhost

import (
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	dbutil "github.com/monstercat/golib/db"
)

type Service struct {
	Id       string `setmap:"ignore"`
	Machine  string
	Name     string
	LastSeen time.Time `db:"last_seen"`
	SigHash  []byte    `db:"sig_hash"`
}

var (
	ColsService = dbutil.GetColumnsList(&Service{}, "")
)

func (s *Service) Insert(tx *sqlx.Tx) error {
	return psql.Insert(TableService).
		SetMap(dbutil.SetMap(s, true)).
		Suffix("RETURNING id").
		RunWith(tx).
		QueryRow().
		Scan(&s.Id)
}

func (s *Service) Update(tx *sqlx.Tx) error {
	if s.Id == "" {
		return ErrInvalidId
	}

	s.LastSeen = time.Now()
	_, err := psql.Update(TableService).
		SetMap(dbutil.SetMap(s, false)).
		Where("id=?", s.Id).
		RunWith(tx).
		Exec()
	return err
}

func (s *Service) UpdateHash(db sqlx.Ext) error {
	if s.Id == "" {
		return ErrInvalidId
	}
	_, err := db.Exec(`UPDATE service SET sig_hash=$2 WHERE id=$1`, s.Id, s.SigHash)
	return err
}

func (s *Service) UpdateLastSeen(db sqlx.Ext) error {
	if s.Id == "" {
		return ErrInvalidId
	}
	_, err := db.Exec(`UPDATE service SET last_seen=NOW() WHERE id=$1`, s.Id)
	return err
}

func GetService(db sqlx.Queryer, where interface{}) (*Service, error) {
	var s Service
	if err := dbutil.Get(db, &s, psql.Select(ColsService...).From(TableService).Where(where)); err != nil {
		return nil, err
	}
	return &s, nil
}

func GetServiceByHash(db sqlx.Queryer, hash []byte) (*Service, error) {
	return GetService(db, squirrel.Eq{"sig_hash": hash})
}

func GetServiceByName(db sqlx.Queryer, machine, name string) (*Service, error) {
	return GetService(db, squirrel.Eq{"name": name, "machine": machine})
}
