package main

import (
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	dbUtil "github.com/monstercat/golib/db"
)

type service struct {
	machine   string
	name      string
	last_seen time.Time
}

func showServices(name string, args []string) error {

	fmt.Printf("%s", txtDanger("Should Show Services"))

	qry := sq.Select("machine", "name", "last_seen").From("services").OrderBy("machine DESC")
	db, _ := getPostgres()

	var services []*service

	if err := dbUtil.Select(db, services, qry); err != nil {
		panic(err)
	}

	return nil
}
