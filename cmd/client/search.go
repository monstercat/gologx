package main

import (
	"fmt"
)

func showSearch(name string, args []string) error {
	fmt.Printf("%s", TxtDanger("Should show search"))

	return nil
}
