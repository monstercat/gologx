package main

import (
	sqlx "github.com/jmoiron/sqlx"
	errors "github.com/pkg/errors"
)

func getPostgres() (*sqlx.DB, error) {
	connStr := "user=logx dbname=logx password=itsasumofthings host=159.203.34.152 sslmode=disable"
	db, err := sqlx.Open("postgres", connStr)

	if err != nil {
		return nil, errors.Wrap(err, "error opening postgres")
	}

	return db, nil
}
