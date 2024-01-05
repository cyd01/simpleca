package key

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"fmt"
	"os"
	"simpleca/tools"
)

func CryptUsage() {
	fmt.Println(`
Usage:  simpleca key crypt [OPTIONS] FILENAME

Add/change a passphrase to an existing private key

Options:`)
	f.PrintDefaults()
	os.Exit(0)
}

func Crypt(args []string) {

	out := f.String("out", "-", "Output file (- for standard output)")
	outpass := f.String("outpass", "", "New passphrase")
	passphrase := f.String("passphrase", "", "Private key passphrase")

	f.SetUsage(CryptUsage)
	f.Parse(args[1:])
	if f.NArg() != 1 {
		CryptUsage()
	} else {
		filename := f.Arg(0)
		if filename != "-" {
			if b, _ := tools.Exists(filename); !b {
				fmt.Fprintln(os.Stderr, "Private key file does not exist")
				os.Exit(1)
			}
		}

		if key, err := LoadPrivateKeyFile(filename, *passphrase); err == nil {
			switch key.(type) {
			case *rsa.PrivateKey:
				if err = WriteRSAKeyFile(key.(*rsa.PrivateKey), *outpass, *out); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			case *ecdsa.PrivateKey:
				fmt.Fprintln(os.Stderr, "ECDSA private key not available")
				os.Exit(1)
			case *ed25519.PrivateKey:
				fmt.Fprintln(os.Stderr, "ED25519 private key not available")
				os.Exit(1)
			default:
				fmt.Fprintln(os.Stderr, "Unknown private key type")
				os.Exit(1)
			}
		} else {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
