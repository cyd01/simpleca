package acmeca

import (
	"bytes"
	"crypto/rand"
	"io"

	//	"crypto/ecdsa"
	//	"crypto/elliptic"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"path"
	"simpleca/flags"
	"simpleca/internal/acme"
	"simpleca/internal/cert"
	"simpleca/internal/key"
	"simpleca/tools"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	f = flags.NewFlag("simpleca")
)

func usage() {
	fmt.Print(`
Usage:  simpleca acme COMMAND

Start an ACME certificate authority web server

Commands:
  

`)
}

var (
	CaKey  *rsa.PrivateKey   = nil
	CaCert *x509.Certificate = nil
	days   int               = 90
)

////
// Variables & Constants
////

const (
	directoryPath  = "/directory"
	newNoncePath   = "/new-nonce"
	newAccountPath = "/new-account"
	newOrderPath   = "/new-order"
	revokeCertPath = "/revoke-cert"
	keyChangePath  = "/key-change"

	finalizePath    = "/finalize/"
	certificatePath = "/certificate/"
	orderPath       = "/order/"
)

func Main(args []string) {
	port := f.String("port", ":8080", "Port server")

	caKeyFile := f.String("ca-key", "ca.key", "Private key of the certificate authority")
	caPassphrase := f.String("ca-pass", "", "Private key passphrase of the certificate authority")
	caCertFile := f.String("ca-cert", "ca.crt", "Certificate of the certificate authority")

	ssl := f.Bool("ssl", false, "Enable SSL server mode")
	keyFile := f.String("key", "", "Private key of the ACME web server (if ssl enabled)")
	certFile := f.String("cert", "", "Certificate of the ACME web server (if ssl enabled)")

	nbDays := f.Int("days", 0, "Not valid after days")

	f.Parse(args[1:])

	if *nbDays != 0 {
		days = *nbDays
	}
	if b, _ := tools.Exists(*caKeyFile); !b {
		fmt.Fprintln(os.Stderr, "Certificate authority private key does not exist")
		os.Exit(1)
	}
	if b, _ := tools.Exists(*caCertFile); !b {
		fmt.Fprintln(os.Stderr, "Certificate authority certificate does not exist")
		os.Exit(1)
	}

	if *ssl {
		if len(*keyFile) == 0 {
			*keyFile = *caKeyFile
		} else if b, _ := tools.Exists(*keyFile); !b {
			fmt.Fprintln(os.Stderr, "Certificate authority web server private key does not exist")
			os.Exit(1)
		}
		if len(*certFile) == 0 {
			*certFile = *caCertFile
		} else if b, _ := tools.Exists(*certFile); !b {
			fmt.Fprintln(os.Stderr, "Certificate authority web server certificate does not exist")
			os.Exit(1)
		}
	}

	var err error
	CaKey, err = key.LoadRSAKeyFile(*caKeyFile, *caPassphrase)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	CaCert, err = cert.LoadCertFile(*caCertFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.Handle(directoryPath, jsonMiddleware(directoryHandler))
	mux.HandleFunc(newNoncePath, nonceHandler)
	mux.Handle(newAccountPath, jsonMiddleware(accountHandler))
	mux.Handle(newOrderPath, jwtMiddleware(jsonMiddleware(newOrderHandler)))
	mux.Handle(finalizePath, jwtMiddleware(jsonMiddleware(finalizeHandler)))
	mux.HandleFunc(certificatePath, certHandler)
	mux.Handle(orderPath, jsonMiddleware(orderHandler))

	if !strings.Contains(*port, ":") {
		*port = ":" + *port
	}
	fmt.Println("Starting ACME web server on port " + *port + " ...")

	if *ssl {
		log.Fatal(http.ListenAndServeTLS(*port, *certFile, *keyFile, mux))
	} else {
		log.Fatal(http.ListenAndServe(*port, mux))
	}

}

////
// Types
////

type acmeFn func(http.ResponseWriter, *http.Request) interface{}

type orderCtx struct {
	obj *acme.Order
	crt []byte
}

type jwsobj struct {
	Protected string `json:"protected"`
	Payload   string `json:"payload"`
	Signature string `json:"signature"`
}

////
// Variables & Constants
////

var orders []*orderCtx
var ordersMtx sync.Mutex

////
// Utility functions
////

func createCrt(csrMsg *acme.CSRMessage) ([]byte, error) {
	data, err := base64.RawURLEncoding.DecodeString(csrMsg.Csr)
	if err != nil {
		fmt.Println("Can not read ACME message")
		return nil, err
	}

	csr, err := x509.ParseCertificateRequest(data)
	if err != nil {
		fmt.Println("Can not parse CSR")
		return nil, err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	temp := x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      csr.Subject,
		//	SignatureAlgorithm: csr.SignatureAlgorithm,
		//	Signature:          csr.Signature,
		NotBefore: time.Now().AddDate(0, 0, -1),
		NotAfter:  time.Now().AddDate(0, 0, days),
		//	KeyUsage:           x509.KeyUsageDigitalSignature,
		//	ExtKeyUsage:        []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		//PublicKey:          csr.PublicKey.(*rsa.PublicKey),
		DNSNames:       csr.DNSNames,
		EmailAddresses: csr.EmailAddresses,
		IPAddresses:    csr.IPAddresses,
	}
	/*
			switch pub:=csr.PublicKey.(type) {
			case *rsa.PublicKey:
			    temp.PublicKey=csr.PublicKey.(*rsa.PublicKey)
			case *ecdsa.PublicKey:
			    temp.PublicKey=csr.PublicKey.(*ecdsa.PublicKey)
			    switch pub.Curve {
		 	    case elliptic.P224(), elliptic.P256():
				temp.SignatureAlgorithm = x509.ECDSAWithSHA256
			    case elliptic.P384():
				temp.SignatureAlgorithm = x509.ECDSAWithSHA384
			    case elliptic.P521():
				temp.SignatureAlgorithm = x509.ECDSAWithSHA512
			    default:
				err = errors.New("x509: unknown elliptic curve")
				return nil, err
			    }
			}*/

	//return x509.CreateCertificate(rand.Reader, &temp, &temp, &key.PublicKey, key)
	//	return x509.CreateCertificate(rand.Reader, &temp, CaCert, csr.PublicKey, CaKey)

	return x509.CreateCertificate(rand.Reader, &temp, CaCert, csr.PublicKey, CaKey)
}

func getOrder(r *http.Request) (*orderCtx, error) {
	id, err := strconv.Atoi(path.Base(r.URL.Path))
	if err != nil {
		fmt.Println("Can not get id")
		return nil, err
	}

	ordersMtx.Lock()
	defer ordersMtx.Unlock()

	if id < len(orders) {
		return orders[id], nil
	} else {
		return nil, errors.New("order not found")
	}
}

func createURL(r *http.Request, path string) string {
	r.URL.Host = r.Host
	r.URL.Scheme = "https"
	r.URL.Path = path

	return r.URL.String()
}

////
// Handlers
////

func directoryHandler(w http.ResponseWriter, r *http.Request) interface{} {
	return acme.Directory{
		NewNonceURL:   createURL(r, newNoncePath),
		NewAccountURL: createURL(r, newAccountPath),
		NewOrderURL:   createURL(r, newOrderPath),
		RevokeCertURL: createURL(r, revokeCertPath),
		KeyChangeURL:  createURL(r, keyChangePath),
	}
}

func nonceHandler(w http.ResponseWriter, r *http.Request) {
	// Hardcoded value copied from RFC 8555
	w.Header().Add("Replay-Nonce", "oFvnlFP1wIhRlYS2jTaXbA")

	w.Header().Add("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
}

func accountHandler(w http.ResponseWriter, r *http.Request) interface{} {
	return acme.Account{
		Status: acme.StatusValid,
		Orders: createURL(r, "orders"),
	}
}

func newOrderHandler(w http.ResponseWriter, r *http.Request) interface{} {
	var order acme.Order
	err := json.NewDecoder(r.Body).Decode(&order)
	if err != nil {
		fmt.Println("Bad Request")
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return nil
	}

	ordersMtx.Lock()
	orderId := strconv.Itoa(len(orders))
	orders = append(orders, &orderCtx{&order, nil})
	ordersMtx.Unlock()

	order.Finalize = createURL(r, path.Join(finalizePath, orderId))
	order.Authorizations = []string{}
	order.Status = "ready"

	orderURL := createURL(r, path.Join(orderPath, orderId))
	w.Header().Add("Location", orderURL)

	w.WriteHeader(http.StatusCreated)
	return order
}

func finalizeHandler(w http.ResponseWriter, r *http.Request) interface{} {
	id := path.Base(r.URL.Path)
	order, err := getOrder(r)
	if err != nil {
		fmt.Println("Not found")
		http.Error(w, "Not Found", http.StatusNotFound)
		return nil
	}

	var csrMsg acme.CSRMessage
	err = json.NewDecoder(r.Body).Decode(&csrMsg)
	if err != nil {
		fmt.Println("Invalid JSON")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return nil
	}

	order.crt, err = createCrt(&csrMsg)
	if err != nil {
		fmt.Println("CreateCrt failed: ", err)
		http.Error(w, "createCrt failed", http.StatusInternalServerError)
		return nil
	}

	order.obj.Status = acme.StatusValid
	order.obj.Certificate = createURL(r, path.Join(certificatePath, id))

	orderURL := createURL(r, path.Join(orderPath, id))
	w.Header().Add("Location", orderURL)

	return order.obj
}

func orderHandler(w http.ResponseWriter, r *http.Request) interface{} {
	order, err := getOrder(r)
	if err != nil {
		fmt.Println("Not found")
		http.Error(w, "Not Found", http.StatusNotFound)
		return nil
	}

	return order.obj
}

func certHandler(w http.ResponseWriter, r *http.Request) {
	order, err := getOrder(r)
	if err != nil {
		fmt.Println("Not found")
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	err = pem.Encode(w, &pem.Block{Type: "CERTIFICATE", Bytes: order.crt})
	if err != nil {
		fmt.Println("PEM encoding failed")
		http.Error(w, "PEM encoding failed", http.StatusInternalServerError)
		return
	}
}

////
// Middleware
////

func jsonMiddleware(fn acmeFn) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.Method, r.URL.String())
		w.Header().Add("Content-Type", "application/json")

		val := fn(w, r)

		if val == nil {
			return
		}
		//fmt.Printf("%v\n",val)

		err := json.NewEncoder(w).Encode(val)
		if err != nil {
			fmt.Println("JSON encoding failed")
			http.Error(w, "JSON encoding failed", http.StatusInternalServerError)
			return
		}
	})
}

func jwtMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var jws jwsobj
		err := json.NewDecoder(r.Body).Decode(&jws)
		if err != nil {
			fmt.Println("Invalid JSON")
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		payload, err := base64.RawURLEncoding.DecodeString(jws.Payload)
		if err != nil {
			fmt.Println("Invalid Base64")
			http.Error(w, "Invalid Base64", http.StatusBadRequest)
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(payload))
		h.ServeHTTP(w, r)
	})
}
