package csr

import (
	"bufio"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
	"simpleca/tools"
	"strconv"
	"strings"
	"unicode"
)

func ReadUsage() {
	fmt.Println(`
Usage:  simpleca csr read [OPTIONS] FILENAME

Read a certificate signing request

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

		if filename != "-" {
			if b, _ := tools.Exists(filename); !b {
				fmt.Fprintln(os.Stderr, "Certificate signing request file does not exist")
				os.Exit(1)
			}
		}

		if csr, err := LoadCSRFile(filename); err == nil {
			PrintCSR(csr)
		} else {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func LoadCSRFile(filename string) (*x509.CertificateRequest, error) {
	if filename == "-" {
		return LoadCSRStream(os.Stdin)
	} else {
		if file, err := os.Open(filename); err == nil {
			defer file.Close()
			return LoadCSRStream(file)
		} else {
			return nil, errors.New("Can not open filename " + filename)
		}
	}
}

func LoadCSRStream(file *os.File) (*x509.CertificateRequest, error) {
	reader := bufio.NewReader(file)
	if bytes, err := io.ReadAll(reader); err != nil {
		return nil, errors.New("Can not read certificate signin request file")
	} else {
		return LoadCSR(bytes)
	}
}

func LoadCSR(bytes []byte) (*x509.CertificateRequest, error) {
	if csrBlock, _ := pem.Decode(bytes); csrBlock == nil {
		return nil, errors.New("Unable to decode certificate signin request")
	} else {
		return ConvertCSRBytes(csrBlock.Bytes)
	}
}

func PrintCSR(csr *x509.CertificateRequest) {
	fmt.Println("Certificate Request:")
	fmt.Println("    Data:")
	fmt.Println("        Version:", csr.Version)
	fmt.Println("        Subject: C =", csr.Subject.Country, ", ST =", csr.Subject.Province, " L =", csr.Subject.Locality, " O =", csr.Subject.Organization, " OU =", csr.Subject.OrganizationalUnit, " CN =", csr.Subject.CommonName)
	fmt.Println("        Subject Public Key Info:")
	fmt.Println("            Public Key Algorithm:", csr.PublicKeyAlgorithm)
	publicKey := csr.PublicKey
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
	fmt.Println("        Attributes:")
	fmt.Println("        Requested Extensions:")
	fmt.Println("            " + csr.Extensions[0].Id.String() + ":")
	fmt.Print("                ")
	for i, v := range csr.Extensions {
		if i > 0 {
			fmt.Println(", ")
		}
		fmt.Println(printable(string(v.Value)))
	}
	if len(csr.DNSNames) > 0 {
		fmt.Println("        DSNNames:")
		fmt.Print("            ")
		for i, v := range csr.DNSNames {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(v)
		}
		fmt.Println("")
	}
	if len(csr.EmailAddresses) > 0 {
		fmt.Println("        EmailAddresses:")
		fmt.Print("            ")
		for i, v := range csr.EmailAddresses {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(v)
		}
		fmt.Println("")
	}
	if len(csr.IPAddresses) > 0 {
		fmt.Println("        IPAddresses:")
		fmt.Print("            ")
		for i, v := range csr.IPAddresses {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(v.String())
		}
		fmt.Println("")
	}
	if len(csr.URIs) > 0 {
		fmt.Println("        URIs:")
		fmt.Print("            ")
		for i, v := range csr.URIs {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(v.String())
		}
		fmt.Println("")
	}
	fmt.Println("    Signature Algorithm:", csr.SignatureAlgorithm)
	fmt.Print("         ")
	for i, v := range csr.Signature {
		fmt.Printf("%02x:", v)
		if ((i + 1) % 18) == 0 {
			fmt.Printf("\n         ")
		}
	}
	fmt.Println("")
}

func printable(text string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, text)
}
