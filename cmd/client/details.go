package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/monstercat/gologx/logxhost"
	"github.com/monstercat/gologx/utils"
)

func showDetails(name string, args []string) error {
	var postgres string
	var id string

	set := flag.NewFlagSet(name, flag.ExitOnError)
	set.StringVar(&postgres, "postgres", "", "Postgres database")
	set.StringVar(&id, "id", "", "Log ID for detailed info")

	if err := set.Parse(args); err != nil {
		return err
	}

	if id == "" {
		fmt.Println(utils.StyleDanger("--id flag is required."))
		os.Exit(1)
	}

	db, err := logxhost.GetPostgresConnection(postgres)
	if err != nil {
		return err
	}
	defer db.Close()

	log, err := logxhost.GetLog(db, id)
	if err != nil {
		return err
	}

	if log == nil {
		fmt.Println(utils.StyleWarning("No logs were found"))
		os.Exit(1)
	}

	msg := func() string {
		if log.Message != "" {
			return log.Message
		} else {
			return "null"
		}
	}()

	fmt.Printf("%s\n", utils.StyleWarning(log.LogType+" - "+log.LogTime.Format("2006-01-02 15:04:05")))
	fmt.Printf("%s %s %s %s\n", utils.StyleWhiteBold("Machine:"), utils.StyleWhite(log.Machine), utils.StyleWhiteBold(" |  Service:"), utils.StyleWhite(log.Service))
	fmt.Printf("%s %s\n", utils.StyleWhiteBold("Message:"), utils.StyleWhite(msg))
	fmt.Printf("%s %s\n", utils.StyleWhiteBold("Context:"), utils.StyleWhite(log.Context))
	return nil
}
