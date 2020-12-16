package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/monstercat/gologx/logxhost"
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

	srvcsArr := func() []string {
		if services != "" {
			return strings.Split(strings.ReplaceAll(services, " ", ""), ",")
		} else {
			return make([]string, 0)
		}
	}()

	srvcs, err := logxhost.SelectServices(db, srvcsArr)
	if err != nil {
		return err
	}

	if len(srvcs) == 0 {
		fmt.Println(styleWarning("No services were found"))
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

	year, month, day, hour, min, _ := timeDiff(time.Now(), date)
	//If heartbeat is bigger than 1 minute, it is inactive.
	if year == 0 && month == 0 && day == 0 && hour == 0 && min == 0 {
		return styleSuccess("Active")
	}

	return styleDanger("Inactive")
}

func timeDiff(a, b time.Time) (year, month, day, hour, min, sec int) {
	if a.Location() != b.Location() {
		b = b.In(a.Location())
	}
	if a.After(b) {
		a, b = b, a
	}
	y1, M1, d1 := a.Date()
	y2, M2, d2 := b.Date()

	h1, m1, s1 := a.Clock()
	h2, m2, s2 := b.Clock()

	year = int(y2 - y1)
	month = int(M2 - M1)
	day = int(d2 - d1)
	hour = int(h2 - h1)
	min = int(m2 - m1)
	sec = int(s2 - s1)

	// Normalize negative values
	if sec < 0 {
		sec += 60
		min--
	}
	if min < 0 {
		min += 60
		hour--
	}
	if hour < 0 {
		hour += 24
		day--
	}
	if day < 0 {
		// days in month:
		t := time.Date(y1, M1, 32, 0, 0, 0, 0, time.UTC)
		day += 32 - t.Day()
		month--
	}
	if month < 0 {
		month += 12
		year--
	}

	return
}
