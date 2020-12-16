package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/monstercat/gologx/logxhost"
	"github.com/monstercat/gologx/utils"
	"github.com/olekukonko/tablewriter"
)

func showStatus(name string, args []string) error {
	var postgres string
	var services string

	set := flag.NewFlagSet(name, flag.ExitOnError)
	set.StringVar(&postgres, "postgres", "", "Postgres database")
	set.StringVar(&services, "services", "", "List of services")

	if err := set.Parse(args); err != nil {
		return err
	}

	db, err := logxhost.GetPostgresConnection(postgres)
	if err != nil {
		return err
	}
	defer db.Close()

	srvcsArr := utils.ParseFlagStrToArray(services, ",")

	srvcs, err := logxhost.SelectServices(db, srvcsArr)
	if err != nil {
		return err
	}

	if len(srvcs) == 0 {
		fmt.Println(utils.StyleWarning("No services were found"))
		os.Exit(1)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Id", "Machine", "Status", "Name", "Last Seen"})

	for _, s := range srvcs {
		row := []string{s.Id, s.Machine, getStatus(s.LastSeen), s.Name, s.LastSeen.Format("2006-01-02 15:04:05")}
		table.Append(row)
	}

	table.Render()

	return nil
}

func getStatus(date time.Time) string {

	if time.Now().Sub(date) <= 60*time.Second {
		return utils.StyleSuccess("Active")
	}

	return utils.StyleDanger("Inactive")
}
