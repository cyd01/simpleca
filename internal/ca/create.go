package ca

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"simpleca/internal/key"
	"simpleca/tools"
	"time"
)

func CreateUsage() {
	fmt.Println(`
Usage:  simpleca ca create [OPTIONS]

Create a new certificate authority

Options:`)
	f.PrintDefaults()
	os.Exit(0)
}

func Create(args []string) {

	privKey := f.StringP("key", "k", "-", "Private key file")
	size := f.IntP("size", "s", 2048, "Private key size")
	passphrase := f.String("passphrase", "", "Private key passphrase")
	days := f.Int("days", 3650, "Not valid after days")
	Country := f.String("C", "FR", "Country name")
	State := f.String("ST", "France", "State")
	Locality := f.String("L", "Paris", "Locality")
	Organization := f.String("O", "MyOrg", "Organization")
	OrganizationalUnit := f.String("OU", "MyUnit", "Unit")
	CommonName := f.String("CN", "MyCA", "Common name")

	out := f.StringP("out", "c", "-", "Output file (- for standard output)")

	f.SetUsage(CreateUsage)
	f.Parse(args[1:])
	if f.NArg() != 0 {
		CreateUsage()
	} else {
		if len(*privKey) == 0 {
			fmt.Fprintf(os.Stderr, "Private key file must be set\n")
			os.Exit(1)
		}
		if len(*out) == 0 {
			fmt.Fprintf(os.Stderr, "Certificate file must be set\n")
			os.Exit(1)
		}

		C := *Country
		ST := *State
		L := *Locality
		O := *Organization
		OU := *OrganizationalUnit
		CN := *CommonName
		SA := ""
		PC := ""

		if b, _ := tools.Exists(*privKey); b {
			fmt.Fprintln(os.Stderr, "Loading CA private key")
			privateKey, err := key.LoadPrivateKeyFile(*privKey, *passphrase)
			if err != nil {
				fmt.Fprintf(os.Stderr, err.Error())
				os.Exit(1)
			}
			fmt.Fprintln(os.Stderr, "Generating CA certificate")
			err = GenerateCACertFile(CN, C, ST, L, O, OU, SA, PC, privateKey, *days, *out)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to generate CA certificate: "+err.Error())
				os.Exit(1)
			}
		} else {
			fmt.Fprintln(os.Stderr, "Generating CA private key")
			privateKey, err := key.GenerateRSAKey(*size)
			if err != nil {
				fmt.Fprintf(os.Stderr, err.Error())
				os.Exit(1)
			}
			err = key.WriteRSAKeyFile(privateKey, *passphrase, *privKey)
			if err != nil {
				fmt.Fprintf(os.Stderr, err.Error())
				os.Exit(1)
			}
			fmt.Fprintln(os.Stderr, "Generating CA certificate")
			err = GenerateCACertFile(CN, C, ST, L, O, OU, SA, PC, privateKey, *days, *out)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to generate CA certificate: "+err.Error())
				os.Exit(1)
			}
		}
	}
}

func GenerateCACert(CN, C, ST, L, O, OU, SA, PC string, key any, days int) (*x509.Certificate, error) {
	if certBytes, err := GenerateCACertBytes(CN, C, ST, L, O, OU, SA, PC, key, days); err != nil {
		return nil, err
	} else {
		return x509.ParseCertificate(certBytes)
	}
}

func ConvertCACertBytes(caBytes []byte) (*x509.Certificate, error) {
	return x509.ParseCertificate(caBytes)
}

func GenerateCACertBytes(CN, C, ST, L, O, OU, SA, PC string, key any, days int) ([]byte, error) {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Country:            []string{},
			Province:           []string{},
			Locality:           []string{},
			Organization:       []string{},
			OrganizationalUnit: []string{},
			StreetAddress:      []string{},
			PostalCode:         []string{},
			CommonName:         CN,
		},
		SignatureAlgorithm: x509.SHA256WithRSA,
		NotBefore:          time.Now(),
		NotAfter:           time.Now().AddDate(0, 0, days),
		KeyUsage:           x509.KeyUsageDigitalSignature | x509.KeyUsageCRLSign | x509.KeyUsageCertSign | x509.KeyUsageKeyAgreement,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageAny,
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageCodeSigning,
			x509.ExtKeyUsageEmailProtection,
			x509.ExtKeyUsageTimeStamping,
			x509.ExtKeyUsageOCSPSigning,
		},
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	if len(C) > 0 {
		ca.Subject.Country = []string{C}
	}
	if len(ST) > 0 {
		ca.Subject.Province = []string{ST}
	}
	if len(L) > 0 {
		ca.Subject.Locality = []string{L}
	}
	if len(O) > 0 {
		ca.Subject.Organization = []string{O}
	}
	if len(OU) > 0 {
		ca.Subject.OrganizationalUnit = []string{OU}
	}
	if len(SA) > 0 {
		ca.Subject.StreetAddress = []string{SA}
	}
	if len(PC) > 0 {
		ca.Subject.PostalCode = []string{PC}
	}
	publicKey := key.(crypto.Signer).Public()
	ca.PublicKey = publicKey
	ca.KeyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageCRLSign | x509.KeyUsageCertSign
	ca.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}

	return x509.CreateCertificate(rand.Reader, ca, ca, publicKey, key)
}

func ConvertCACertBytesToBlock(caBytes []byte) *pem.Block {
	return &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	}
}

func GenerateCACertBlock(CN, C, ST, L, O, OU, SA, PC string, key any, days int) (*pem.Block, error) {
	if certBytes, err := GenerateCACertBytes(CN, C, ST, L, O, OU, SA, PC, key, days); err != nil {
		return nil, err
	} else {
		return ConvertCACertBytesToBlock(certBytes), nil
	}
}

func GenerateCACertStream(CN, C, ST, L, O, OU, SA, PC string, key any, days int, file *os.File) error {
	if certBlock, err := GenerateCACertBlock(CN, C, ST, L, O, OU, SA, PC, key, days); err != nil {
		return err
	} else {
		err = pem.Encode(file, certBlock)
		if err != nil {
			return errors.New("Error when encode private pem: " + err.Error())
		}
		return nil
	}
}

func GenerateCACertFile(CN, C, ST, L, O, OU, SA, PC string, key any, days int, filename string) error {
	var certPem *os.File
	var err error
	if filename == "-" {
		certPem = os.Stdout
	} else {
		certPem, err = os.Create(filename)
		if err != nil {
			return errors.New("Error when creating file")
		}
		defer certPem.Close()
	}
	return GenerateCACertStream(CN, C, ST, L, O, OU, SA, PC, key, days, certPem)
}

func GenerateCAPrivateKeyFile(keyfile, passphrase string, size int) error {
	privateKey, err := key.GenerateRSAKey(size)
	if err != nil {
		return err
	}
	err = key.WriteRSAKeyFile(privateKey, passphrase, keyfile)
	if err != nil {
		return err
	}
	return nil
}
