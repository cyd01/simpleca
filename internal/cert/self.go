package cert

import (
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
	"simpleca/internal/key"
	"simpleca/tools"
	"strings"
	"time"
)

func SelfUsage() {
	fmt.Println(`
Usage:  simpleca cert self [OPTIONS]

Create a self-signed certificate

Options:`)
	f.PrintDefaults()
	os.Exit(0)
}

func Self(args []string) {
	privKey := f.StringP("key", "k", "-", "Server private key file")
	size := f.IntP("size", "s", 2048, "Private key size")
	passphrase := f.String("passphrase", "", "Server private key passphrase")

	AltNames := f.StringP("alt-names", "a", "", "Coma separated alternate names list")
	IPs := f.StringP("ips", "i", "", "Coma separated IP addresses list")
	days := f.Int("days", 3650, "Not valid after days")
	Country := f.String("C", "FR", "Country name")
	State := f.String("ST", "France", "State")
	Locality := f.String("L", "Paris", "Locality")
	Organization := f.String("O", "MyOrg", "Organization")
	OrganizationalUnit := f.String("OU", "MyUnit", "Unit")
	CommonName := f.String("CN", "", "Common name")

	out := f.StringP("out", "c", "-", "Output file (- for standard output)")

	f.SetUsage(SelfUsage)
	f.Parse(args[1:])
	if f.NArg() != 0 {
		SelfUsage()
	} else {
		altNames := []string{}
		if len(*AltNames) > 0 {
			list := strings.Split(*AltNames, ",")
			for _, l := range list {
				altNames = append(altNames, strings.TrimSpace(l))
			}
		}
		ips := []net.IP{}
		if len(*IPs) > 0 {
			list := strings.Split(*IPs, ",")
			for _, l := range list {
				ips = append(ips, net.ParseIP(strings.TrimSpace(l)))
			}
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
			fmt.Fprintln(os.Stderr, "Loading private key")
			privateKey, err := key.LoadPrivateKeyFile(*privKey, *passphrase)
			if err != nil {
				fmt.Fprintf(os.Stderr, err.Error())
				os.Exit(1)
			}
			fmt.Fprintln(os.Stderr, "Generating self-signed certificate")
			err = GenerateSelfCertFile(CN, C, ST, L, O, OU, SA, PC, altNames, ips, privateKey, *days, *out)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to generate self-signed certificate: "+err.Error())
				os.Exit(1)
			}
		} else {
			fmt.Fprintln(os.Stderr, "Generating private key")
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
			fmt.Fprintln(os.Stderr, "Generating self-signed certificate")
			err = GenerateSelfCertFile(CN, C, ST, L, O, OU, SA, PC, altNames, ips, privateKey, *days, *out)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to generate self-signed certificate: "+err.Error())
				os.Exit(1)
			}
		}
	}
}

func GenerateSelfCertBytes(name, C, ST, L, O, OU, SA, PC string, altNames []string, ips []net.IP, privKey any, days int) ([]byte, error) {
	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Country:            []string{},
			Province:           []string{},
			Locality:           []string{},
			Organization:       []string{},
			OrganizationalUnit: []string{},
			StreetAddress:      []string{},
			PostalCode:         []string{},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24 * 180),
		DNSNames:              altNames,
		IPAddresses:           ips,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	if len(C) > 0 {
		tmpl.Subject.Country = []string{C}
	}
	if len(ST) > 0 {
		tmpl.Subject.Province = []string{ST}
	}
	if len(L) > 0 {
		tmpl.Subject.Locality = []string{L}
	}
	if len(O) > 0 {
		tmpl.Subject.Organization = []string{O}
	}
	if len(OU) > 0 {
		tmpl.Subject.OrganizationalUnit = []string{OU}
	}
	if len(SA) > 0 {
		tmpl.Subject.StreetAddress = []string{SA}
	}
	if len(PC) > 0 {
		tmpl.Subject.PostalCode = []string{PC}
	}
	if len(name) > 0 {
		tmpl.Subject.CommonName = name
	}
	return x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, key.PublicKey(privKey), privKey)
}

func ConvertSelfCertBytesToBlock(certBytes []byte) *pem.Block {
	return &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	}
}

func GenerateSelfCertBlock(name, C, ST, L, O, OU, SA, PC string, altNames []string, ips []net.IP, key any, days int) (*pem.Block, error) {
	if certBytes, err := GenerateSelfCertBytes(name, C, ST, L, O, OU, SA, PC, altNames, ips, key, days); err != nil {
		return nil, err
	} else {
		return ConvertSelfCertBytesToBlock(certBytes), nil
	}
}

func GenerateSelfCertStream(name, C, ST, L, O, OU, SA, PC string, altNames []string, ips []net.IP, key any, days int, file *os.File) error {
	if certBlock, err := GenerateSelfCertBlock(name, C, ST, L, O, OU, SA, PC, altNames, ips, key, days); err != nil {
		return err
	} else {
		err = pem.Encode(file, certBlock)
		if err != nil {
			return errors.New("Error when encode private pem: " + err.Error())
		}
		return nil
	}
}

func GenerateSelfCertFile(name, C, ST, L, O, OU, SA, PC string, altNames []string, ips []net.IP, key any, days int, filename string) error {
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
	return GenerateSelfCertStream(name, C, ST, L, O, OU, SA, PC, altNames, ips, key, days, certPem)
}

func ConvertCertBytesToBlock(certBytes []byte) *pem.Block {
	return &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}
}

func ConvertCertBlockToBytes(certBlock *pem.Block) []byte {
	return certBlock.Bytes
}

func ConvertCertToBytes(cert *x509.Certificate) []byte {
	return cert.Raw
}

func ConvertCertToBlock(cert *x509.Certificate) *pem.Block {
	return &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}
}

func WriteCertStream(cert *x509.Certificate, file *os.File) error {
	certBlock := ConvertCertToBlock(cert)
	err := pem.Encode(file, certBlock)
	if err != nil {
		return errors.New("Error when encode private key to : " + err.Error())
	}
	return nil
}

func WriteCertFile(cert *x509.Certificate, filename string) error {
	var file *os.File
	var err error
	if filename == "-" {
		file = os.Stdout
	} else {
		file, err = os.Create(filename)
		if err != nil {
			return errors.New("Error when creating file")
		}
		defer file.Close()
	}
	return WriteCertStream(cert, file)
}
