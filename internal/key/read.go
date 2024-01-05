package key

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
	"simpleca/tools"
)

func ReadUsage() {
	fmt.Println(`
Usage:  simpleca key read [OPTIONS] FILENAME

Read a key

Options:`)
	f.PrintDefaults()
	os.Exit(0)
}

var (
	ECDSA   = errors.New("ECDSA")
	ED25519 = errors.New("ED25519")
)

func Read(args []string) {

	passphrase := f.String("passphrase", "", "Private key passphrase")

	f.SetUsage(ReadUsage)
	f.Parse(args[1:])
	if f.NArg() != 1 {
		ReadUsage()
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
				if bytes, err := EncodeRSAPrivateKeyToPEM(key.(*rsa.PrivateKey), ""); err == nil {
					fmt.Println(string(bytes))
				} else {
					fmt.Fprintln(os.Stderr, err)
				}
			case *ecdsa.PrivateKey:
				if bytes, err := EncodECPrivateKeyToPEM(key.(*ecdsa.PrivateKey), ""); err == nil {
					fmt.Println(string(bytes))
				} else {
					fmt.Fprintln(os.Stderr, err)
				}
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

func LoadPrivateKeyFile(filename, passphrase string) (any, error) {
	if filename == "-" {
		return LoadPrivateKeyStream(os.Stdin, passphrase)
	} else {
		if file, err := os.Open(filename); err == nil {
			defer file.Close()
			return LoadPrivateKeyStream(file, passphrase)
		} else {
			return nil, errors.New("Can not open filename " + filename)
		}
	}
}

func LoadPrivateKeyStream(file *os.File, passphrase string) (any, error) {
	reader := bufio.NewReader(file)
	if bytes, err := io.ReadAll(reader); err != nil {
		return nil, errors.New("Can not read private key file")
	} else {
		return LoadPrivateKey(bytes, passphrase)
	}
}

func LoadPrivateKey(bytes []byte, passphrase string) (any, error) {
	if key, err := LoadRSAKey(bytes, passphrase); err == nil {
		return key, nil
	} else if err == ECDSA {
		if key, err := LoadECDSAKey(bytes, passphrase); err == nil {
			return key, nil
		} else {
			return nil, err
		}
	} else if err == ED25519 {
		return nil, err
	} else {
		return nil, err
	}
}

func LoadECDSAKeyFile(filename, passphrase string) (*ecdsa.PrivateKey, error) {
	if filename == "-" {
		return LoadECDSAKeyStream(os.Stdin, passphrase)
	} else {
		if file, err := os.Open(filename); err == nil {
			defer file.Close()
			return LoadECDSAKeyStream(file, passphrase)
		} else {
			return nil, errors.New("Can not open filename " + filename)
		}
	}
}

func LoadECDSAKeyStream(file *os.File, passphrase string) (*ecdsa.PrivateKey, error) {
	reader := bufio.NewReader(file)
	if bytes, err := io.ReadAll(reader); err != nil {
		return nil, errors.New("Can not read private key file")
	} else {
		return LoadECDSAKey(bytes, passphrase)
	}
}

func LoadECDSAKey(bytes []byte, passphrase string) (*ecdsa.PrivateKey, error) {
	var key *ecdsa.PrivateKey = nil
	var err error
	block, _ := pem.Decode(bytes)
	if block == nil {
		return key, errors.New("Can not decode private key")
	} else {
		switch block.Type {
		case "EC PRIVATE KEY":
			if key, err = x509.ParseECPrivateKey(block.Bytes); err == nil {
			} else {
				return key, errors.New("Can not decode private key")
			}
		default:
			return key, errors.New("Unknown private key type")
		}
	}
	return key, nil
}

func LoadRSAKeyFile(filename, passphrase string) (*rsa.PrivateKey, error) {
	if filename == "-" {
		return LoadRSAKeyStream(os.Stdin, passphrase)
	} else {
		if file, err := os.Open(filename); err == nil {
			defer file.Close()
			return LoadRSAKeyStream(file, passphrase)
		} else {
			return nil, errors.New("Can not open filename " + filename)
		}
	}
}

func LoadRSAKeyStream(file *os.File, passphrase string) (*rsa.PrivateKey, error) {
	reader := bufio.NewReader(file)
	if bytes, err := io.ReadAll(reader); err != nil {
		return nil, errors.New("Can not read private key file")
	} else {
		return LoadRSAKey(bytes, passphrase)
	}
}

func LoadRSAKey(bytes []byte, passphrase string) (*rsa.PrivateKey, error) {
	var key *rsa.PrivateKey = nil
	var err error
	block, _ := pem.Decode(bytes)
	if block == nil {
		return key, errors.New("Can not decode private key")
	} else {
		var privateKeyBytes []byte
		switch block.Type {
		case "PRIVATE KEY":
			if mkey, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
				switch mkey.(type) {
				case *rsa.PrivateKey:
					return mkey.(*rsa.PrivateKey), nil
				case *ecdsa.PrivateKey:
					return nil, ECDSA
				case *ed25519.PrivateKey:
					return nil, errors.New("ED25519")
				default:
					return nil, errors.New("Unknown private key type")
				}
			} else {
				return key, errors.New("Can not decode private key")
			}
		case "RSA PRIVATE KEY":
			if x509.IsEncryptedPEMBlock(block) {
				if privateKeyBytes, err = x509.DecryptPEMBlock(block, []byte(passphrase)); err != nil {
					return key, errors.New("Can not decrypt private key with passphrase")
				}
			} else {
				privateKeyBytes = block.Bytes
			}
			if key, err = x509.ParsePKCS1PrivateKey(privateKeyBytes); err != nil {
				return key, errors.New("Can not parse private key")
			}
		case "EC PRIVATE KEY":
			return nil, ECDSA
		default:
			return key, errors.New("Unknown private key type")
		}
	}
	return key, nil
}
