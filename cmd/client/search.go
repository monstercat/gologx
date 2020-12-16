package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/Masterminds/squirrel"
	pgUtils "github.com/monstercat/golib/db/postgres"
	"github.com/monstercat/gologx/logxhost"
	"github.com/olekukonko/tablewriter"
)

func showSearch(name string, args []string) error {
	var postgres string
	var service string
	var matcher string
	var limit int
	var datebefore string
	var dateafter string
	var orderby string

	limitDefault := 50

	set := flag.NewFlagSet(name, flag.ExitOnError)
	set.StringVar(&postgres, "postgres", "", "Postgres database")
	set.StringVar(&service, "service", "", "Defines a service to be filtered")
	set.StringVar(&matcher, "matcher", "", "Defines a matcher to be filtered. Searches machine and name")
	set.IntVar(&limit, "limit", limitDefault, "Defines the limit of results in the list")
	set.StringVar(&datebefore, "datebefore", "", "Filters before a specific date time")
	set.StringVar(&dateafter, "dateafter", "", "Filters after a specific date time")
	set.StringVar(&orderby, "orderby", "", "Orders list by a defined column")

	if err := set.Parse(args); err != nil {
		return err
	}

	if service == "" {
		fmt.Println(styleDanger("service flag is required"))
		os.Exit(1)
	}

	dateRegex := regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
	if valid := dateRegex.MatchString(datebefore); datebefore != "" && !valid {
		fmt.Println(styleDanger("yyyy/mm/dd format is required for datebefore"))
		os.Exit(1)
	}

	if valid := dateRegex.MatchString(dateafter); dateafter != "" && !valid {
		fmt.Println(styleDanger("yyyy/mm/dd format is required for dateafter"))
		os.Exit(1)
	}

	db, err := logxhost.GetPostgresConnection(postgres)
	if err != nil {
		return err
	}
	defer db.Close()

	search := getQuery(service, matcher, limit, datebefore, dateafter, orderby)

	logs, err := logxhost.SelectLogs(db, search)
	if err != nil {
		return err
	}

	if len(logs) == 0 {
		fmt.Println(styleWarning("No logs were found"))
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

func getQuery(service string, matcher string, limit int, datebefore string, dateafter string, orderby string) squirrel.SelectBuilder {
	q := logxhost.BuildSelectLogs().Where("service LIKE '" + service + "'")

	if matcher != "" {
		some := squirrel.Or{}
		some = append(some, squirrel.Like{"machine": "'" + matcher + "'"})
		some = append(some, squirrel.Like{"message": "'" + matcher + "'"})
		q = q.Where(some)
	}

	if datebefore != "" {
		q = q.Where("log_time < '" + datebefore + "'")
	}

	if dateafter != "" {
		q = q.Where("log_time >= '" + dateafter + "'")
	}

	if orderby != "" {
		orderByArr := func() []string {
			if orderby != "" {
				return strings.Split(strings.ReplaceAll(orderby, " ", ""), ",")
			} else {
				return make([]string, 0)
			}
		}()

		pgUtils.ApplySortStrings(orderByArr, &q, map[string]string{"logTime": "log_time", "created": "created", "logType": "log_type", "service": "service", "machine": "machine"})
	}

	q = q.Limit(uint64(limit))

	log.Print(q.ToSql())

	return q
}
