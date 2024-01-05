package cert

import (
	"bufio"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"simpleca/tools"
	"strconv"
	"strings"
	"time"
)

func ReadUsage() {
	fmt.Println(`
Usage:  simpleca cert read [OPTIONS] FILENAME

Read a certificate

Options:`)
	f.PrintDefaults()
	os.Exit(0)
}

func Read(args []string) {

	f.SetUsage(ReadUsage)
	f.Parse(args[1:])
	if f.NArg() != 1 {
		ReadUsage()
	} else {
		filename := f.Arg(0)

		if strings.HasPrefix(filename, "https://") {
			if certs, err := LoadCertsServer(filename); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			} else {
				for _, c := range certs {
					fmt.Print(c.Subject.CommonName + " - ")
					PrintCert(c)
				}
			}
		} else {
			if filename != "-" {
				if b, _ := tools.Exists(filename); !b {
					fmt.Fprintln(os.Stderr, "Certificate file does not exist")
					os.Exit(1)
				}
			}
			if crt, err := LoadCertFile(filename); err == nil {
				PrintCert(crt)
			} else {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
	}
}

func PrintCert(cert *x509.Certificate) {
	fmt.Println("Certificate:")
	fmt.Println("    Data:")
	fmt.Println("        Version:", cert.Version)
	fmt.Println("        Serial Number:", cert.SerialNumber, "("+fmt.Sprintf("0x%x", cert.SerialNumber)+")")
	fmt.Println("        Signature Algorithm:", cert.SignatureAlgorithm)
	fmt.Println("        Issuer: C =", cert.Issuer.Country, ", ST =", cert.Issuer.Province, " L =", cert.Issuer.Locality, " O =", cert.Issuer.Organization, " OU =", cert.Issuer.OrganizationalUnit, " CN=", cert.Subject.CommonName)
	fmt.Println("        Validity:")
	fmt.Println("            Not Before: ", cert.NotBefore)
	fmt.Println("            Not After : ", cert.NotAfter)
	fmt.Println("        Subject: C =", cert.Subject.Country, ", ST =", cert.Subject.Province, " L =", cert.Subject.Locality, " O =", cert.Subject.Organization, " OU =", cert.Subject.OrganizationalUnit, " CN=", cert.Subject.CommonName)
	fmt.Println("        Subject Public Key Info:")
	fmt.Println("            Public Key Algorithm:", cert.PublicKeyAlgorithm)
	publicKey := cert.PublicKey
	var size int = 0
	switch publicKey.(type) {
	case *rsa.PublicKey:
		pub := publicKey.(*rsa.PublicKey)
		size = pub.Size() * 8
		fmt.Println("                RSA Public-Key: (" + strconv.Itoa(size) + " bit)")
		fmt.Println("                Modulus:")
		asn1Bytes, _ := asn1.Marshal(pub.N)
		fmt.Printf("                    ")
		for i, v := range asn1Bytes {
			if i > 3 {
				fmt.Printf("%02x:", v)
				if ((i - 3) % 15) == 0 {
					fmt.Printf("\n                    ")
				}
			}
		}
		fmt.Printf("\n")
		fmt.Println("                Exponent:", strconv.Itoa(pub.E), "("+fmt.Sprintf("0x%x", pub.E)+")")
	default:
		fmt.Println("                Unkonwn public key type")
	}
	fmt.Println("        X509v3 extensions:")
	for _, v := range cert.Extensions {
		fmt.Print("            " + v.Id.String() + ": ")
		if v.Critical {
			fmt.Println("critical")
		} else {
			fmt.Println("")
		}
		//		fmt.Println("                " + string(v.Value))
	}
	fmt.Println("            KeyUsage:")
	fmt.Print("                ")
	if cert.KeyUsage&x509.KeyUsageDigitalSignature != 0 {
		fmt.Print("Digital Signature, ")
	}
	if cert.KeyUsage&x509.KeyUsageContentCommitment != 0 {
		fmt.Print("Content Commitment, ")
	}
	if cert.KeyUsage&x509.KeyUsageKeyEncipherment != 0 {
		fmt.Print("Key Encipherment, ")
	}
	if cert.KeyUsage&x509.KeyUsageDataEncipherment != 0 {
		fmt.Print("Data Encipherment, ")
	}
	if cert.KeyUsage&x509.KeyUsageKeyAgreement != 0 {
		fmt.Print("Key Agreement, ")
	}

	if cert.KeyUsage&x509.KeyUsageCertSign != 0 {
		fmt.Print("Cert Sign, ")
	}
	if cert.KeyUsage&x509.KeyUsageCRLSign != 0 {
		fmt.Print("CRL Sign, ")
	}
	if cert.KeyUsage&x509.KeyUsageEncipherOnly != 0 {
		fmt.Print("Encipher Only, ")
	}
	if cert.KeyUsage&x509.KeyUsageDecipherOnly != 0 {
		fmt.Print("Decipher Only, ")
	}
	fmt.Println("")
	if len(cert.DNSNames) > 0 {
		fmt.Println("        DSNNames:")
		fmt.Print("            ")
		for i, v := range cert.DNSNames {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(v)
		}
		fmt.Println("")
	}
	if len(cert.EmailAddresses) > 0 {
		fmt.Println("        EmailAddresses:")
		fmt.Print("            ")
		for i, v := range cert.EmailAddresses {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(v)
		}
		fmt.Println("")
	}
	if len(cert.IPAddresses) > 0 {
		fmt.Println("        IPAddresses:")
		fmt.Print("            ")
		for i, v := range cert.IPAddresses {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(v.String())
		}
		fmt.Println("")
	}
	if len(cert.URIs) > 0 {
		fmt.Println("        URIs:")
		fmt.Print("            ")
		for i, v := range cert.URIs {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(v.String())
		}
		fmt.Println("")
	}
	fmt.Println("    Signature Algorithm:", cert.SignatureAlgorithm)
	fmt.Print("         ")
	for i, v := range cert.Signature {
		fmt.Printf("%02x:", v)
		if ((i + 1) % 18) == 0 {
			fmt.Printf("\n         ")
		}
	}
	fmt.Println("")
}

func LoadCertFile(filename string) (*x509.Certificate, error) {
	if filename == "-" {
		return LoadCertStream(os.Stdin)
	} else if strings.HasPrefix(filename, "https://") {
		certs, err := LoadCertsServer(filename)
		return certs[0], err
	} else {
		if file, err := os.Open(filename); err == nil {
			defer file.Close()
			return LoadCertStream(file)
		} else {
			return nil, errors.New("Can not open filename " + filename)
		}
	}
}

func LoadCertStream(file *os.File) (*x509.Certificate, error) {
	reader := bufio.NewReader(file)
	if bytes, err := io.ReadAll(reader); err != nil {
		return nil, errors.New("Can not read certificate file")
	} else {
		return LoadCert(bytes)
	}
}

func LoadCert(bytes []byte) (*x509.Certificate, error) {
	if certBlock, _ := pem.Decode(bytes); certBlock == nil {
		return nil, errors.New("Unable to decode certificate file")
	} else {
		return ConvertCertBytes(certBlock.Bytes)
	}
}

func ConvertCertBytes(caBytes []byte) (*x509.Certificate, error) {
	return x509.ParseCertificate(caBytes)
}

func LoadCertsServer(filename string) ([]*x509.Certificate, error) {
	host, port, err := SplitHostPort(strings.TrimPrefix(filename, "https://"))
	if err != nil {
		return nil, err
	}
	d := &net.Dialer{
		Timeout: time.Duration(5) * time.Second,
	}
	conn, err := tls.DialWithDialer(d, "tcp", host+":"+port, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	cert := conn.ConnectionState().PeerCertificates
	return cert, err
}

func SplitHostPort(hostport string) (string, string, error) {
	if !strings.Contains(hostport, ":") {
		return hostport, "443", nil
	}
	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		return "", "", err
	}
	if port == "" {
		port = "443"
	}
	return host, port, nil
}

/*
// https://yourbasic.org/golang/bitmask-flag-set-clear/

type Bits uint8

const (
    F0 Bits = 1 << iota
    F1
    F2
)

func Set(b, flag Bits) Bits    { return b | flag }
func Clear(b, flag Bits) Bits  { return b &^ flag }
func Toggle(b, flag Bits) Bits { return b ^ flag }
func Has(b, flag Bits) bool    { return b & flag != 0 }
*/
