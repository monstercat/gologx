package main

import (
	"fmt"
)

type command struct {
	title       string
	usage       string
	description string
}

var commands = []command{
	{
		title:       "help",
		usage:       "logx help",
		description: "Prints out a list of commands you can use.",
	},
	{
		title:       "status",
		usage:       "logx status -postgres postgresql://login:password@127.0.0.1/dbname --services serviceA, serviceB",
		description: "Will print the status of matched services (last seen, alive, etc.).\n   It can have 0 - N service names. No service name input will output all services.",
	},
	{
		title:       "search",
		usage:       "logx search --postgres postgresql://login:password@127.0.0.1/dbname --service serviceA --dateafter 2010-01-01 00:00:00 -datebefore 2010-01-05 00:00:00 --limit 100 --matcher machinename --orderby columnA, columnB",
		description: "Allows to search for logs matching considering the search criteria. Service flag is mandatory. Default limit is 50.",
	},
	{
		title:       "details",
		usage:       "logx details --postgres postgresql://login:password@127.0.0.1/dbname --id 12345",
		description: "Allows log details searching by ID. Id flag is mandatory.",
	},
}

func showHelp(name string, args []string) error {
	fmt.Printf("%s\n", styleHighlight("List of commands:"))
	for _, cmd := range commands {
		fmt.Printf(">> %s\n   %s: %s\n   %s\n\n", styleHighlight(cmd.title), styleWhiteBold("Usage"), styleWhite(cmd.usage), styleWhite(cmd.description))
	}
	return nil
}
