package server

import (
	"database/sql"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/cyc-ttn/pos"
	"github.com/jmoiron/sqlx"
)


type Origin struct {
	Id   string
	Name string
}

type Service struct {
	Id       string
	OriginId string
	Name     string
	LastSeen time.Time `db:"last_seen"`
	SigHash  []byte    `db:"sig_hash"`

	// Filled by join
	OriginName string `db:"origin_name"`
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
	if err := pos.GetForStruct(db, &s, ViewService, where); err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return &s, nil
}

func GetServiceByHash(db sqlx.Queryer, hash []byte) (*Service, error) {
	return GetService(db, squirrel.Eq{"sig_hash": hash})
}

func GetServiceByName(db sqlx.Queryer, name string) (*Service, error) {
	return GetService(db, squirrel.Eq{"name": name})
}

// Creates an origin or returns an existing origin.
const createOriginQuery = `
INSERT INTO origin(name) VALUES($1)
ON CONFLICT DO UPDATE SET name=EXCLUDED.name 
RETURNING  id, name
`

func CreateOrReturnOrigin(db sqlx.Ext, origin string) (*Origin, error) {
	var o Origin
	if err := db.QueryRowx(createOriginQuery, origin).Scan(&o.Id, &o.Name); err != nil {
		return nil, err
	}
	return &o, nil
}

// Creates a service from origin and service combo.
// It will look for the origin if it exists. If it doesn't,
// it will create it.
const createServiceQuery = `
INSERT INTO service(name, last_seen, sig_hash) VALUES($1, NOW(), $2)
ON CONFLICT DO UPDATE SET last_seen=NOW() 
RETURNING id, name, last_seen, sig_hash
`

func CreateService(db sqlx.Ext, originName, service string, sigHash []byte) (*Service, error) {
	origin, err := CreateOrReturnOrigin(db, originName)
	if err != nil {
		return nil, err
	}
	s := Service{
		OriginId:   origin.Id,
		OriginName: origin.Name,
	}
	if err := db.QueryRowx(createServiceQuery, service, sigHash).Scan(&s.Id, &s.Name, &s.LastSeen, &s.SigHash); err != nil {
		return nil, err
	}
	return &s, nil
}
