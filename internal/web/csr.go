package web

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"encoding/json"
	"encoding/pem"
	"io"
	"net"
	"net/http"
	"simpleca/internal/csr"
	"simpleca/internal/key"
	"simpleca/tools"
	"strings"
)

func Csr(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	var err error

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable ro read private key: "+err.Error(), http.StatusInternalServerError)
		return
	}

	passphrase := GetParam(r, "passphrase", "")
	C := GetParam(r, "C", ConfigC)
	ST := GetParam(r, "ST", ConfigST)
	L := GetParam(r, "L", ConfigL)
	O := GetParam(r, "O", ConfigO)
	OU := GetParam(r, "OU", ConfigOU)
	name := GetParam(r, "CN", "")
	SA := ""
	PC := ""
	if len(name) == 0 {
		http.Error(w, "Common name can not be empty", http.StatusBadRequest)
		return
	}
	altNames := strings.Split(GetParam(r, "altnames", name), ",")
	ipss := GetParam(r, "ips", "")
	var ips []net.IP = []net.IP{}
	if len(ipss) > 0 {
		for _, i := range strings.Split(ipss, ",") {
			ips = append(ips, net.ParseIP(strings.TrimSpace(i)))
		}
	} else {
		ips = []net.IP{net.ParseIP(strings.TrimSpace("127.0.0.1"))}
	}

	if kkey, err := key.LoadPrivateKey(body, passphrase); err != nil {
		http.Error(w, "Unable to convert to private key: "+err.Error(), http.StatusBadRequest)
	} else {
		switch kkey.(type) {
		case *rsa.PrivateKey:
			privateKey := kkey.(*rsa.PrivateKey)
			ccsr, err := csr.GenerateCSR(name, C, ST, L, O, OU, SA, PC, altNames, ips, privateKey)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Error while generate certificate signing request: " + err.Error()))
				return
			}
			csrBlock := csr.ConvertCSRToBlock(ccsr)
			csrBytes := pem.EncodeToMemory(csrBlock)
			if tools.Contains(r.Header["Accept"], "application/json") {
				w.Header().Add("Content-type", "application/json")
				r := Resp{Csr: string(csrBytes)}
				resp, _ := json.Marshal(r)
				w.WriteHeader(http.StatusOK)
				w.Write(resp)
			} else {
				w.Header().Add("Content-type", "text/plain")
				w.WriteHeader(http.StatusOK)
				w.Write(csrBytes)
			}
		case *ecdsa.PrivateKey:
			privateKey := kkey.(*ecdsa.PrivateKey)
			ccsr, err := csr.GenerateCSR(name, C, ST, L, O, OU, SA, PC, altNames, ips, privateKey)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Error while generate certificate signing request"))
				return
			}
			csrBlock := csr.ConvertCSRToBlock(ccsr)
			csrBytes := pem.EncodeToMemory(csrBlock)
			if tools.Contains(r.Header["Accept"], "application/json") {
				w.Header().Add("Content-type", "application/json")
				r := Resp{Csr: string(csrBytes)}
				resp, _ := json.Marshal(r)
				w.WriteHeader(http.StatusOK)
				w.Write(resp)
			} else {
				w.Header().Add("Content-type", "text/plain")
				w.WriteHeader(http.StatusOK)
				w.Write(csrBytes)
			}
		default:
			http.Error(w, "Wrong key type", http.StatusBadRequest)
		}
	}
}
