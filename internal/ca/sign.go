package ca

import (
	"crypto/rand"
	"crypto/x509"
	"fmt"
	"math/big"
	"os"
	"simpleca/internal/cert"
	"simpleca/internal/csr"
	"simpleca/internal/key"
	"simpleca/tools"
	"time"
)

func SignUsage() {
	fmt.Println(`
Usage:  simpleca ca sign [OPTIONS] FILENAME

Sign a certificate signing request

Options:`)
	f.PrintDefaults()
	os.Exit(0)
}

func Sign(args []string) {

	caKeyFile := f.String("ca-key", "ca.key", "Private key of the certificate authority")
	caPassphrase := f.String("ca-pass", "", "Private key passphrase of the certificate authority")
	caCertFile := f.String("ca-cert", "ca.crt", "Certificate of the certificate authority")

	days := f.Int("days", 3650, "Not valid after days")
	out := f.StringP("out", "c", "-", "Output file (- for standard output)")

	f.SetUsage(SignUsage)
	f.Parse(args[1:])
	if f.NArg() != 1 {
		SignUsage()
	} else {
		filename := f.Arg(0)

		if filename != "-" {
			if b, _ := tools.Exists(filename); !b {
				fmt.Fprintln(os.Stderr, "Certificate signing request file does not exist")
				os.Exit(1)
			}
		}
		if b, _ := tools.Exists(*caKeyFile); !b {
			fmt.Fprintln(os.Stderr, "Certificate authority private key does not exist")
			os.Exit(1)
		}
		if b, _ := tools.Exists(*caCertFile); !b {
			fmt.Fprintln(os.Stderr, "Certificate authority certificate does not exist")
			os.Exit(1)
		}

		caKey, err := key.LoadPrivateKeyFile(*caKeyFile, *caPassphrase)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		caCert, err := cert.LoadCertFile(*caCertFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		csr, err := csr.LoadCSRFile(filename)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		crt, err := CASign(csr, *days, caCert, caKey)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Unable to sign certificate "+err.Error())
			os.Exit(1)
		}

		fmt.Fprintln(os.Stderr, "Generating certificate")
		cert.WriteCertFile(crt, *out)
	}
}

// Sign CSR with CA
func CASign(csr *x509.CertificateRequest, days int, ca *x509.Certificate, caPrivKey any) (*x509.Certificate, error) {
	var err error
	if err = csr.CheckSignature(); err != nil {
		return nil, err
	}
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}
	certTemplate := x509.Certificate{
		SerialNumber:       serialNumber,
		Subject:            csr.Subject,
		SignatureAlgorithm: csr.SignatureAlgorithm,
		Signature:          csr.Signature,
		NotBefore:          time.Now(),
		NotAfter:           time.Now().AddDate(0, 0, days),
		KeyUsage:           x509.KeyUsageDigitalSignature,
		ExtKeyUsage:        []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		PublicKey:          csr.PublicKey,
		DNSNames:           csr.DNSNames,
		EmailAddresses:     csr.EmailAddresses,
		IPAddresses:        csr.IPAddresses,
		URIs:               csr.URIs,
	}
	//certTemplate.PublicKeyAlgorithm = csr.PublicKeyAlgorithm
	crtBytes, err := x509.CreateCertificate(rand.Reader, &certTemplate, ca, csr.PublicKey, caPrivKey)
	if err != nil {
		return nil, err
	}
	return x509.ParseCertificate(crtBytes)
}
