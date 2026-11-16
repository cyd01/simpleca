package key

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

func CreateUsage() {
	fmt.Println(`
Usage:  simpleca key create [OPTIONS]

Create a new private key

Options:`)
	f.PrintDefaults()
	os.Exit(0)
}

func Create(args []string) {

	size := f.Int("size", 2048, "Private key size (in bits")
	out := f.String("out", "-", "Output file (- for standard output)")
	passphrase := f.String("passphrase", "", "Private key passphrase")
	ktype := f.String("type", "rsa", "Private key type (rsa, or ecdsa)")

	f.SetUsage(CreateUsage)
	f.Parse(args[1:])
	if f.NArg() != 0 {
		CreateUsage()
	} else {
		switch *ktype {
		case "rsa":
			if err := GenerateRSAKeyFile(*size, *passphrase, *out); err != nil {
				fmt.Fprint(os.Stderr, err.Error())
				os.Exit(1)
			}
		case "ecdsa":
			if err := GenerateECDSAKeyFile(*passphrase, *out); err != nil {
				fmt.Fprint(os.Stderr, err.Error())
				os.Exit(1)
			}
		default:
			fmt.Fprintf(os.Stderr, "Unknown private key type")
			os.Exit(1)
		}
	}
}

// createKeyFile opens a private key file for writing, readable by its owner
// only, including when it replaces an existing file.
func createKeyFile(filename string) (*os.File, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return nil, err
	}
	if err := file.Chmod(0o600); err != nil {
		file.Close()
		return nil, err
	}
	return file, nil
}

func GenerateRSAKey(size int) (*rsa.PrivateKey, error) {
	if size < MinRSABits {
		return nil, fmt.Errorf("RSA key size must be at least %d bits", MinRSABits)
	}
	if privatekey, err := rsa.GenerateKey(rand.Reader, size); err != nil {
		return nil, errors.New("Can not generate RSA private key")
	} else {
		return privatekey, nil
	}
}

func GenerateECDSAKey() (*ecdsa.PrivateKey, error) {
	if privatekey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader); err != nil {
		return nil, errors.New("Can not generate ECDSA private key")
	} else {
		return privatekey, nil
	}
}

func WriteRSAKeyStream(privatekey *rsa.PrivateKey, passphrase string, file *os.File) error {
	if privateKeyBlock, err := ConvertRSAKeyToBlock(privatekey, passphrase); err != nil {
		return err
	} else {
		err = pem.Encode(file, privateKeyBlock)
		if err != nil {
			return errors.New("Error when encode private key to : " + err.Error())
		}
		return nil
	}
}

func WriteECDSAKeyStream(privatekey *ecdsa.PrivateKey, passphrase string, file *os.File) error {
	if privateKeyBlock, err := ConvertECDSAKeyToBlock(privatekey, passphrase); err != nil {
		return err
	} else {
		err = pem.Encode(file, privateKeyBlock)
		if err != nil {
			return errors.New("Error when encode private key to pem: " + err.Error())
		}
		return nil
	}
}

func WriteRSAKeyFile(privatekey *rsa.PrivateKey, passphrase, filename string) error {
	var file *os.File
	var err error
	if filename == "-" {
		file = os.Stdout
	} else {
		file, err = createKeyFile(filename)
		if err != nil {
			return errors.New("Error when creating file")
		}
		defer file.Close()
	}
	return WriteRSAKeyStream(privatekey, passphrase, file)
}

func WriteECDSAKeyFile(privatekey *ecdsa.PrivateKey, passphrase, filename string) error {
	var file *os.File
	var err error
	if filename == "-" {
		file = os.Stdout
	} else {
		file, err = createKeyFile(filename)
		if err != nil {
			return errors.New("Error when creating file")
		}
		defer file.Close()
	}
	return WriteECDSAKeyStream(privatekey, passphrase, file)
}

func ConvertRSAKeyToBlock(privatekey *rsa.PrivateKey, passphrase string) (*pem.Block, error) {
	if passphrase == "" {
		return &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privatekey),
		}, nil
	}
	return encryptKey(privatekey, passphrase)
}

func ConvertECDSAKeyToBlock(privatekey *ecdsa.PrivateKey, passphrase string) (*pem.Block, error) {
	if passphrase != "" {
		return encryptKey(privatekey, passphrase)
	}
	privateKeyBytes, err := x509.MarshalECPrivateKey(privatekey)
	if err != nil {
		return nil, errors.New("Can not convert ECDSA to Block: " + err.Error())
	}
	return &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privateKeyBytes,
	}, nil
}

func encryptKey(privatekey any, passphrase string) (*pem.Block, error) {
	der, err := x509.MarshalPKCS8PrivateKey(privatekey)
	if err != nil {
		return nil, errors.New("Unable to encode private key: " + err.Error())
	}
	block, err := encryptPKCS8(der, passphrase)
	if err != nil {
		return nil, errors.New("Unable to encrypt private key: " + err.Error())
	}
	return block, nil
}

func GenerateRSAKeyBlock(size int, passphrase string) (*pem.Block, error) {
	privatekey, err := GenerateRSAKey(size)
	if err != nil {
		return nil, err
	} else {
		if privateKeyBlock, err := ConvertRSAKeyToBlock(privatekey, passphrase); err != nil {
			return nil, err
		} else {
			return privateKeyBlock, nil
		}
	}
}

func GenerateECDSAKeyBlock(passphrase string) (*pem.Block, error) {
	privatekey, err := GenerateECDSAKey()
	if err != nil {
		return nil, err
	} else {
		if privateKeyBlock, err := ConvertECDSAKeyToBlock(privatekey, passphrase); err != nil {
			return nil, err
		} else {
			return privateKeyBlock, nil
		}
	}
}

func GenerateRSAKeyStream(size int, passphrase string, file *os.File) error {
	if privateKeyBlock, err := GenerateRSAKeyBlock(size, passphrase); err != nil {
		return err
	} else {
		err = pem.Encode(file, privateKeyBlock)
		if err != nil {
			return errors.New("Error when encode private pem: " + err.Error())
		}
		return nil
	}
}

func GenerateECDSAKeyStream(passphrase string, file *os.File) error {
	if privateKeyBlock, err := GenerateECDSAKeyBlock(passphrase); err != nil {
		return err
	} else {
		err = pem.Encode(file, privateKeyBlock)
		if err != nil {
			return errors.New("Error when encode private pem: " + err.Error())
		}
		return nil
	}
}

func GenerateRSAKeyFile(size int, passphrase string, filename string) error {
	var privatePem *os.File
	var err error
	if filename == "-" {
		privatePem = os.Stdout
	} else {
		privatePem, err = createKeyFile(filename)
		if err != nil {
			return errors.New("Error when creating file")
		}
		defer privatePem.Close()
	}
	return GenerateRSAKeyStream(size, passphrase, privatePem)
}

func GenerateECDSAKeyFile(passphrase string, filename string) error {
	var privatePem *os.File
	var err error
	if filename == "-" {
		privatePem = os.Stdout
	} else {
		privatePem, err = createKeyFile(filename)
		if err != nil {
			return errors.New("Error when creating file")
		}
		defer privatePem.Close()
	}
	return GenerateECDSAKeyStream(passphrase, privatePem)
}
