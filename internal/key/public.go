package key

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"simpleca/tools"
)

func PublicUsage() {
	fmt.Println(`
Usage:  simpleca key pub [OPTIONS] FILENAME

Get public key from private key

Options:`)
	f.PrintDefaults()
	os.Exit(0)
}

func Public(args []string) {

	out := f.String("out", "-", "Output file (- for standard output)")
	passphrase := f.String("passphrase", "", "Private key passphrase")

	f.SetUsage(PublicUsage)
	f.Parse(args[1:])
	if f.NArg() != 1 {
		PublicUsage()
	} else {
		filename := f.Arg(0)

		if filename != "-" {
			if b, _ := tools.Exists(filename); !b {
				fmt.Fprintln(os.Stderr, "Private key file does not exist")
				os.Exit(1)
			}
		}

		if key, err := LoadPrivateKeyFile(filename, *passphrase); err == nil {
			publicKeyBlock, err := GetPublicKey(key)
			if err != nil {
				fmt.Printf("error when dumping publickey: %s\n", err)
				os.Exit(1)
			}
			var publicPem *os.File
			if *out == "-" {
				publicPem = os.Stdout
			} else {
				publicPem, err = os.Create(*out)
				if err != nil {
					fmt.Printf("Error when create %s: %s\nn", *out, err)
					os.Exit(1)
				}
				defer publicPem.Close()
			}
			err = pem.Encode(publicPem, publicKeyBlock)
			if err != nil {
				fmt.Printf("Error when encode public pem: %s\nn", err)
				os.Exit(1)
			}
		} else {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func GetPublicKey(key any) (*pem.Block, error) {
	publicKey := key.(crypto.Signer).Public()
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, errors.New("error when dumping publickey: " + err.Error())
	}
	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	return publicKeyBlock, nil
}
