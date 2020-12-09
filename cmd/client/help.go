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
		title:       "services",
		usage:       "logx services",
		description: "Prints out a list of services for log querying.",
	},
	{
		title:       "status",
		usage:       "logx status [service names...]",
		description: "Will print the status of matched services (last seen, alive, etc.).\n   It can have 0 - N service names. No service name input will output all services.",
	},
	{
		title:       "search",
		usage:       "logx search [service name] [text match] [date (before/after)]",
		description: "Allows to search for logs matching any the the criteria",
	},
}

func showHelp(name string, args []string) error {
	fmt.Printf("%s\n", TxtHighlight("List of commands:"))
	for _, cmd := range commands {
		fmt.Printf(">> %s\n   %s: %s\n   %s\n\n", TxtHighlight(cmd.title), TxtWhiteBold("Usage"), TxtWhite(cmd.usage), TxtWhite(cmd.description))
	}
	return nil
}
