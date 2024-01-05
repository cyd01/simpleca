package cert

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
Usage:  simpleca cert COMMAND

Manage server certificates

Commands:
  read             Read a certficate
  self             Create a self-signed certificate

`)
}

func Main(args []string) {
	if len(args) <= 1 {
		usage()
	} else {
		argsWithoutProg := args[1:]
		switch cmd := argsWithoutProg[0]; cmd {
		case "read":
			Read(argsWithoutProg)
		case "self":
			Self(argsWithoutProg)
		default:
			fmt.Fprintln(os.Stderr, "Unknown command "+cmd)
			usage()
			os.Exit(1)
		}
	}
}
