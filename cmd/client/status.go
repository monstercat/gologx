package main

import (
	"fmt"
)

func showStatus(name string, args []string) error {
	fmt.Printf("%s", TxtDanger("Should show status"))
	return nil
}
