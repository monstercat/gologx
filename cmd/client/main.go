package main

import (
	"fmt"
	"os"

	cmd "github.com/tmathews/commander"
)

func main() {
	var args []string

	if len(os.Args) >= 2 {
		args = os.Args[1:]
		fmt.Println(" ")
	} else {
		fmt.Printf("\n\n%s\n", txtHighlight("Welcome to Monstercat LogX!\n"))
		return
	}

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
