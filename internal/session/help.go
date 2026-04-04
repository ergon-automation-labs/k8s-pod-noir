package session

import (
	_ "embed"
	"strings"
)

//go:embed testdata/repl_help.golden
var replHelpBody string

func replHelpText() string {
	return strings.TrimSpace(replHelpBody)
}
