package web

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"simpleca/flags"
	"simpleca/internal/ca"
	"simpleca/internal/cert"
	"simpleca/internal/key"
	"simpleca/tools"
	"strings"
	"syscall"
	"time"
)

var (
	f = flags.NewFlag("simpleca")
)

func usage() {
	fmt.Print(`
Usage:  simpleca web COMMAND

Start an automatic certificate authority web server

Options:
`)
	f.PrintDefaults()
	os.Exit(1)
}

var (
	CaKey     *rsa.PrivateKey   = nil
	CaCert    *x509.Certificate = nil
	CaCertURL string            = ""
)

type TKey struct {
	Priv string `json:"priv,omitempty"`
	Pub  string `json:"pub,omitempty"`
}

type TCa struct {
	Key string `json:"key,omitempty"`
	Crt string `json:"crt,omitempty"`
}

type Resp struct {
	Csr string `json:"csr,omitempty"`
	Crt string `json:"crt,omitempty"`
	Key *TKey  `json:"key,omitempty"`
	Ca  *TCa   `json:"ca,omitempty"`
}

func Main(args []string) {
	dir := f.String("dir", ".", "Root directory")
	port := f.String("port", ":80", "Port server")

	caKeyFile := f.String("ca-key", "ca.key", "Private key of the certificates authority")
	caPassphrase := f.String("ca-pass", "", "Private key passphrase of the certificates authority")
	caCertFile := f.String("ca-cert", "ca.crt", "Certificate of the certificates authority")
	caCertURL := f.String("issuer-cert-url", "", "URL of the certificates authority's certificate")

	ssl := f.Bool("ssl", false, "Enable SSL server mode")
	keyFile := f.String("key", "", "Private key of the certificates authority web server")
	certFile := f.String("cert", "", "Certificate of the certificates authority web server")

	c := f.String("C", ConfigC, "Default Country name")
	st := f.String("ST", ConfigST, "Default State")
	l := f.String("L", ConfigL, "Default Locality")
	o := f.String("O", ConfigO, "Default Organization")
	ou := f.String("OU", ConfigOU, "Default Unit")
	nbDays := f.Int("nbdays", ConfigDays, "Default certificate number of days expiration")
	size := f.Int("size", ConfigSize, "Default private key size")

	f.SetUsage(usage)
	f.Parse(args[1:])

	ConfigC = *c
	ConfigST = *st
	ConfigL = *l
	ConfigO = *o
	ConfigOU = *ou
	ConfigDays = *nbDays
	ConfigSize = *size

	var err error

	if b, _ := tools.Exists(*caKeyFile); !b {
		fmt.Fprintln(os.Stderr, "Certificate authority private key does not exist, creating", *caKeyFile)
		if err := ca.GenerateCAPrivateKeyFile(*caKeyFile, *caPassphrase, *size); err != nil {
			fmt.Fprintln(os.Stderr, "Can not create CA private key file")
			os.Exit(1)
		}
	}
	CaKey, err = key.LoadRSAKeyFile(*caKeyFile, *caPassphrase)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if b, _ := tools.Exists(*caCertFile); !b {
		fmt.Fprintln(os.Stderr, "Certificate authority certificate does not exist, creating", *caCertFile)
		if err := ca.GenerateCACertFile("EasyCA", *c, *st, *l, *o, *ou, "", "", CaKey, *nbDays, *caCertFile); err != nil {
			fmt.Fprintln(os.Stderr, "Can not create CA certificate file")
			os.Exit(1)
		}
	}
	CaCert, err = cert.LoadCertFile(*caCertFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if len(*caCertURL) > 0 {
		CaCertURL = *caCertURL
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

	mux := http.NewServeMux()
	mux.HandleFunc("/alive", alive)
	mux.Handle("/key/pub", Logs(http.HandlerFunc(Pub)))
	mux.Handle("/key", Logs(http.HandlerFunc(Key)))
	mux.Handle("/csr", Logs(http.HandlerFunc(Csr)))
	mux.Handle("/crt", Logs(http.HandlerFunc(Crt)))
	mux.Handle("/sign", Logs(http.HandlerFunc(Sign)))
	mux.Handle("/ca/ca.crt", Logs(http.HandlerFunc(CaCaCrt)))

	mux.Handle("/", Logs(http.FileServer(http.Dir(*dir))))

	if !strings.Contains(*port, ":") {
		*port = ":" + *port
	}
	fmt.Println("Starting web server on port " + *port + " ...")

	var server = &http.Server{
		Addr:    *port,
		Handler: mux,
	}
	go func() {
		if *ssl {
			err = server.ListenAndServeTLS(*certFile, *keyFile)
		} else {
			err = server.ListenAndServe()
		}
		if err != nil {
			if err != http.ErrServerClosed {
				fmt.Fprintln(os.Stderr, "Can not start CA web server", err)
			}
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if server != nil {
		if err := server.Shutdown(ctx); err != nil {
			log.Fatal("Server forced to shutdown:", err)
		}
	}
	log.Println("Server exiting")
}

func alive(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	fmt.Fprintf(w, "alive")
}

func GetParam(r *http.Request, name, def string) string {
	t := r.URL.Query().Get(name)
	if len(t) == 0 {
		t = def
	}
	return t
}
