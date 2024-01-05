package csr

import (
	"fmt"
	"os"
	"simpleca/flags"
)

var (
	f = flags.NewFlag("simpleca")
)

func usage() {
	fmt.Print(`
Usage:  simpleca csr COMMAND

Manage server certificate signing request

Commands:
  create           Create a new certificate signing request
  read             Read a certificate signing request

`)
}

func Main(args []string) {
	if len(args) <= 1 {
		usage()
	} else {
		argsWithoutProg := args[1:]
		switch cmd := argsWithoutProg[0]; cmd {
		case "create":
			Create(argsWithoutProg)
		case "new":
			Create(argsWithoutProg)
		case "read":
			Read(argsWithoutProg)
		default:
			fmt.Fprintln(os.Stderr, "Unknown command "+cmd)
			usage()
			os.Exit(1)
		}
	}
}
