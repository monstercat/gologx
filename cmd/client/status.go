package main

import (
	"fmt"
)

func showStatus(name string, args []string) error {
	fmt.Printf("%s", txtDanger("Should show status"))

	return nil
}
