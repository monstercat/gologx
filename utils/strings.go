package utils

import (
	"fmt"
	"os"
	"strings"
)

func ParseFlagStrToArray(str string, separator string) []string {
	if separator == "" {
		fmt.Println(StyleDanger("separator is required."))
		os.Exit(1)
	}

	if str != "" {
		return strings.Split(strings.ReplaceAll(str, " ", ""), separator)
	} else {
		return make([]string, 0)
	}
}
