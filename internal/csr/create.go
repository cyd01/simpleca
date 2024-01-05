package csr

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"simpleca/internal/key"
	"simpleca/tools"
	"strings"
)

func CreateUsage() {
	fmt.Println(`
Usage:  simpleca csr create [OPTIONS]

Create a new certificate signing request

Options:`)
	f.PrintDefaults()
	os.Exit(0)
}

func Create(args []string) {

	privKey := f.StringP("key", "k", "-", "Server private key file")
	size := f.IntP("size", "s", 2048, "Private key size")
	passphrase := f.String("passphrase", "", "Server private key passphrase")
	AltNames := f.StringP("alt-names", "a", "", "Coma separated alternate names list")
	IPs := f.StringP("ips", "i", "", "Coma separated IP addresses list")
	Country := f.String("C", "", "Country name")
	State := f.String("ST", "", "State")
	Locality := f.String("L", "", "Locality")
	Organization := f.String("O", "", "Organization")
	OrganizationalUnit := f.String("OU", "", "Unit")
	CommonName := f.String("CN", "", "Common name")
	out := f.String("out", "-", "Output file (- for standard output)")

	f.SetUsage(CreateUsage)
	f.Parse(args[1:])
	if f.NArg() != 0 {
		CreateUsage()
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

			fmt.Fprintln(os.Stderr, "Generating certificate sign request")
			err = GenerateCSRFile(CN, C, ST, L, O, OU, SA, PC, altNames, ips, privateKey, *out)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Can not generate CSR: "+err.Error())
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

			fmt.Fprintln(os.Stderr, "Generating certificate sign request")
			err = GenerateCSRFile(CN, C, ST, L, O, OU, SA, PC, altNames, ips, privateKey, *out)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Can not generate CSR: "+err.Error())
				os.Exit(1)
			}
		}
	}
}

func GenerateCSR(name, C, ST, L, O, OU, SA, PC string, altNames []string, ips []net.IP, key any) (*x509.CertificateRequest, error) {
	csrBytes, err := GenerateCSRBytes(name, C, ST, L, O, OU, SA, PC, altNames, ips, key)
	if err != nil {
		return nil, err
	}
	return x509.ParseCertificateRequest(csrBytes)
}

func GenerateCSRFile(name, C, ST, L, O, OU, SA, PC string, altNames []string, ips []net.IP, key any, filename string) error {
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
	return GenerateCSRStream(name, C, ST, L, O, OU, SA, PC, altNames, ips, key, file)
}

func GenerateCSRStream(name, C, ST, L, O, OU, SA, PC string, altNames []string, ips []net.IP, key any, file *os.File) error {
	if csrBlock, err := GenerateCSRBlock(name, C, ST, L, O, OU, SA, PC, altNames, ips, key); err != nil {
		return err
	} else {
		err = pem.Encode(file, csrBlock)
		if err != nil {
			return errors.New("Error when encode certificate signing request pem: " + err.Error())
		}
		return nil
	}
}

func GenerateCSRBlock(name, C, ST, L, O, OU, SA, PC string, altNames []string, ips []net.IP, key any) (*pem.Block, error) {
	csr, err := GenerateCSRBytes(name, C, ST, L, O, OU, SA, PC, altNames, ips, key)
	if err != nil {
		return nil, err
	}
	return ConvertCSRBytesToBlock(csr), nil
}

func GenerateCSRBytes(name, C, ST, L, O, OU, SA, PC string, altNames []string, ips []net.IP, key any) ([]byte, error) {
	csrTemplate, err := GenerateCSRTemplate(name, C, ST, L, O, OU, SA, PC, altNames, ips, key)
	if err != nil {
		return nil, err
	}
	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, csrTemplate, key)
	if err != nil {
		return nil, err
	}
	return csrBytes, nil
}

func ConvertCSRBytesToBlock(csrBytes []byte) *pem.Block {
	return &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes}
}

func ConvertCSRBlockToBytes(csrBlock *pem.Block) []byte {
	return csrBlock.Bytes
}

func ConvertCSRBytes(csrBytes []byte) (*x509.CertificateRequest, error) {
	return x509.ParseCertificateRequest(csrBytes)
}

func ConvertCSRToBytes(csr *x509.CertificateRequest) []byte {
	return csr.Raw
}

func ConvertCSRToBlock(csr *x509.CertificateRequest) *pem.Block {
	return &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csr.Raw}
}

func WriteCSRFile(csr *x509.CertificateRequest, filename string) error {
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
	return WriteCSRStream(csr, file)
}

func WriteCSRStream(csr *x509.CertificateRequest, file *os.File) error {
	csrBlock := ConvertCSRBytesToBlock(csr.Raw)
	err := pem.Encode(file, csrBlock)
	if err != nil {
		return errors.New("Error when encode csr to pem: " + err.Error())
	}
	return nil
}

func GenerateCSRTemplate(name, C, ST, L, O, OU, SA, PC string, altNames []string, ips []net.IP, key any) (*x509.CertificateRequest, error) {
	tmpl := &x509.CertificateRequest{
		Subject: pkix.Name{
			Country:            []string{},
			Province:           []string{},
			Locality:           []string{},
			Organization:       []string{},
			OrganizationalUnit: []string{},
			StreetAddress:      []string{},
			PostalCode:         []string{},
			CommonName:         name,
		},
		SignatureAlgorithm: x509.SHA256WithRSA,
		ExtraExtensions:    []pkix.Extension{},
		DNSNames:           altNames,
		EmailAddresses:     []string{},
		IPAddresses:        ips,
		URIs:               []*url.URL{},
	}
	tmpl.PublicKey = key.(crypto.Signer).Public()
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
	return tmpl, nil
}
