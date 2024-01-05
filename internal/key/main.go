package key

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"simpleca/flags"
)

var (
	f = flags.NewFlag("simpleca")
)

func usage() {
	fmt.Print(`
Usage:  simpleca key COMMAND

Manage keys

Commands:
  create           Create a new private key
  crypt            Add/change a passphrase to an existing private key
  pub              Get public key from private key
  read             Print key informations

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
		case "crypt":
			Crypt(argsWithoutProg)
		case "new":
			Create(argsWithoutProg)
		case "pub":
			Public(argsWithoutProg)
		case "public":
			Public(argsWithoutProg)
		case "read":
			Read(argsWithoutProg)
		default:
			fmt.Fprintln(os.Stderr, "Unknown command "+cmd)
			usage()
			os.Exit(1)
		}
	}
}

func PublicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func PemBlockForKey(priv interface{}) *pem.Block {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
			os.Exit(2)
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	default:
		return nil
	}
}
