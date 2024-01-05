package ca

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
Usage:  simpleca ca COMMAND

Manage certificate authority

Commands:
  create           Create or renew a certficate authority
  sign             Sign a certificate with a certificate authority previously created

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
		case "sign":
			Sign(argsWithoutProg)
		default:
			fmt.Fprintln(os.Stderr, "Unknown command "+cmd)
			usage()
			os.Exit(1)
		}
	}
}
