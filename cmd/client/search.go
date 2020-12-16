package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"

	"github.com/monstercat/gologx/logxhost"
	"github.com/monstercat/gologx/utils"
	"github.com/olekukonko/tablewriter"
)

func showSearch(name string, args []string) error {
	var postgres string
	var service string
	var matcher string
	var limit int
	var dateBefore string
	var dateAfter string
	var orderby string

	limitDefault := 50

	set := flag.NewFlagSet(name, flag.ExitOnError)
	set.StringVar(&postgres, "postgres", "", "Postgres database")
	set.StringVar(&service, "service", "", "Defines a service to be filtered")
	set.StringVar(&matcher, "matcher", "", "Defines a matcher to be filtered. Searches machine and name")
	set.IntVar(&limit, "limit", limitDefault, "Defines the limit of results in the list")
	set.StringVar(&dateBefore, "datebefore", "", "Filters before a specific date time")
	set.StringVar(&dateAfter, "dateafter", "", "Filters after a specific date time")
	set.StringVar(&orderby, "orderby", "", "Orders list by a defined column")

	if err := set.Parse(args); err != nil {
		return err
	}

	if service == "" {
		fmt.Println(utils.StyleDanger("service flag is required"))
		os.Exit(1)
	}

	dateRegex := regexp.MustCompile("((19|20)\\d\\d)-(0?[1-9]|1[012])-(0?[1-9]|[12][0-9]|3[01])")
	if valid := dateRegex.MatchString(dateBefore); !valid && dateBefore != "" {
		fmt.Println(utils.StyleDanger("yyyy-mm-dd format is required for datebefore"))
		os.Exit(1)
	}

	if valid := dateRegex.MatchString(dateAfter); !valid && dateAfter != "" {
		fmt.Println(utils.StyleDanger("yyyy-mm-dd format is required for dateafter"))
		os.Exit(1)
	}

	db, err := logxhost.GetPostgresConnection(postgres)
	if err != nil {
		return err
	}
	defer db.Close()

	search := logxhost.GetLogSearchQuery(service, matcher, limit, dateBefore, dateAfter, orderby)

	logs, err := logxhost.SelectLogs(db, search)
	if err != nil {
		return err
	}

	if len(logs) == 0 {
		fmt.Println(utils.StyleWarning("No logs were found"))
		os.Exit(1)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Id", "Machine", "Service", "Log Type", "Message", "Log Time"})

	for _, l := range logs {
		row := []string{l.Id, l.Machine, l.Service, l.LogType, l.Message, l.LogTime.Format("2006-01-02 15:04:05")}
		table.Append(row)
	}

	table.Render()

	return nil
}
