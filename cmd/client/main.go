package main

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	cmd "github.com/tmathews/commander"
)

func main() {
	var args []string

	if len(os.Args) >= 2 {
		args = os.Args[1:]
	}

	fmt.Printf("\n\n%s\n", TxtHighlight("Welcome to Monstercat LogX!\n"))

	err := cmd.Exec(args, cmd.DefaultHelper, cmd.M{
		"help":     showHelp,
		"status":   showStatus,
		"search":   showSearch,
		"services": showServices,
	})
	if err != nil {
		switch v := err.(type) {
		case cmd.Error:
			fmt.Print(v.Help())
			os.Exit(2)
		default:
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}
}

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