package golang

import (
	"strings"

	"github.com/fatih/camelcase"
)

func CreateReceiver(name string) string {
	nameWords := camelcase.Split(name)
	if len(nameWords) == 0 {
		return "r"
	}

	return strings.ToLower(string(nameWords[len(nameWords)-1][0]))
}
