package web

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net"
	"net/http"
	"simpleca/internal/ca"
	"simpleca/internal/cert"
	"simpleca/internal/csr"
	"simpleca/internal/key"
	"simpleca/tools"
	"strconv"
	"strings"
)

// Generate all (key+csr+crt+ca.crt) all in one

func Crt(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	var err error

	var size int = ConfigSize
	s := GetParam(r, "size", strconv.Itoa(ConfigSize))
	if len(s) > 0 {
		if size, err = strconv.Atoi(s); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Wrong key size"))
			return
		}
	}
	if size < 1024 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Key size not big enough"))
		return
	}
	if size > 16384 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Key size too much big"))
		return
	}
	var days int = 3650
	s = GetParam(r, "days", strconv.Itoa(ConfigDays))
	if len(s) > 0 {
		if days, err = strconv.Atoi(s); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Wrong number of days"))
			return
		}
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

	mkey, err := key.GenerateRSAKey(size)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error while generate private key"))
		return
	}
	keyBlock, err := key.ConvertRSAKeyToBlock(mkey, passphrase)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Can not convert key to block"))
		return
	}
	keyBytes := pem.EncodeToMemory(keyBlock)
	publicKey, err := x509.MarshalPKIXPublicKey(&mkey.PublicKey)
	publicKeyBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKey,
	})

	ccsr, err := csr.GenerateCSR(name, C, ST, L, O, OU, SA, PC, altNames, ips, mkey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error while generate certificate signing request"))
		return
	}
	csrBlock := csr.ConvertCSRToBlock(ccsr)
	csrBytes := pem.EncodeToMemory(csrBlock)

	ccrt, err := ca.CASign(ccsr, days, CaCert, CaKey, CaCertURL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Can not sign certificate signing request"))
		return
	}
	crtBlock := cert.ConvertCertToBlock(ccrt)
	crtBytes := pem.EncodeToMemory(crtBlock)

	CaCertBlock := cert.ConvertCertToBlock(CaCert)
	caCertBytes := pem.EncodeToMemory(CaCertBlock)

	if tools.Contains(r.Header["Accept"], "application/json") {
		w.Header().Add("Content-type", "application/json")
		r := Resp{Csr: string(csrBytes), Crt: string(crtBytes)}
		r.Key = new(TKey)
		r.Key.Priv = string(keyBytes)
		r.Key.Pub = string(publicKeyBytes)
		r.Ca = new(TCa)
		r.Ca.Crt = string(caCertBytes)
		resp, _ := json.Marshal(r)
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	} else {
		w.Header().Add("Content-type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write(keyBytes)
		//w.Write([]byte("\r\n"))
		w.Write(csrBytes)
		//w.Write([]byte("\r\n"))
		w.Write(crtBytes)
	}
}
