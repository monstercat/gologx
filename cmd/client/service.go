package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/monstercat/gologx/logxhost"
)

func showServices(name string, args []string) error {
	var postgres string
	set := flag.NewFlagSet(name, flag.ExitOnError)
	set.StringVar(&postgres, "postgres", "", "Postgres database")
	if err := set.Parse(args); err != nil {
		return err
	}

	db, err := getPostgresConnection(postgres)
	if err != nil {
		return err
	}

	fmt.Printf("%s", TxtDanger("Should Show Services"))

	services, err := logxhost.SelectServices(db)
	if err != nil {
		return err
	}

	log.Print(services)

	return nil
}
