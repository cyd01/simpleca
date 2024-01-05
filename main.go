package simpleca

import (
	"fmt"
	"os"

	"simpleca/internal/acmeca"
	"simpleca/internal/ca"
	"simpleca/internal/cert"
	"simpleca/internal/csr"
	"simpleca/internal/key"
	"simpleca/internal/web"
)

func usage() {
	fmt.Print(`
Usage:  simpleca COMMAND [OPTIONS]

A simple tools to manage SSL stuff

Commands:
  acme             Start an ACME certificate authority web server
  ca               Manage certificate authority
  cert             Manage server certificates
  csr              Manage server certificate signing request
  key              Manage keys
  web              Start an automatic certificate authority web server

Run 'simpleca COMMAND -h' for more informations on a command

`)
}

func Main(args []string) {
	if len(os.Args) <= 1 {
		usage()
	} else {
		argsWithoutProg := os.Args[1:]
		switch cmd := argsWithoutProg[0]; cmd {
		case "acme":
			acmeca.Main(argsWithoutProg)
		case "ca":
			ca.Main(argsWithoutProg)
		case "crt":
			cert.Main(argsWithoutProg)
		case "cert":
			cert.Main(argsWithoutProg)
		case "csr":
			csr.Main(argsWithoutProg)
		case "key":
			key.Main(argsWithoutProg)
		case "web":
			web.Main(argsWithoutProg)

		case "info":
			printInfo()

		default:
			fmt.Fprintln(os.Stderr, "Unknown command "+cmd)
			usage()
			os.Exit(1)
		}
	}
}

var (
	BuildTime = "Undefined"
	Author    = "Unknown"
	Revision  = "Undefined"
)

func printInfo() {
	fmt.Println("Build at", BuildTime, "by", Author, "on revision", Revision)
}
