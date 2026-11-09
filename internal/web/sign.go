package web

import (
	"encoding/json"
	"encoding/pem"
	"net/http"
	"simpleca/internal/ca"
	"simpleca/internal/cert"
	"simpleca/internal/csr"
	"simpleca/tools"
)

func Sign(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	days, ok := requestedDays(w, r)
	if !ok {
		return
	}
	body, ok := readBody(w, r)
	if !ok {
		return
	}

	if ccsr, err := csr.LoadCSR(body); err != nil {
		http.Error(w, "Unable to convert to certificate signing request: "+err.Error(), http.StatusBadRequest)
	} else {
		crt, err := ca.CASign(ccsr, days, CaCert, CaKey, CaCertURL)
		if err != nil {
			http.Error(w, "Can not sign certificate signing request: "+err.Error(), http.StatusBadRequest)
			return
		}
		bytes := pem.EncodeToMemory(cert.ConvertCertToBlock(crt))
		if tools.Contains(r.Header["Accept"], "application/json") {
			w.Header().Add("Content-type", "application/json")
			resp, _ := json.Marshal(Resp{Crt: string(bytes)})
			w.WriteHeader(http.StatusOK)
			w.Write(resp)
		} else {
			w.Header().Add("Content-type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write(bytes)
		}
	}
}

func CaCaCrt(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	CaCertBlock := cert.ConvertCertToBlock(CaCert)
	bytes := pem.EncodeToMemory(CaCertBlock)
	if tools.Contains(r.Header["Accept"], "application/json") {
		w.Header().Add("Content-type", "application/json")
		var r Resp
		r.Ca = new(TCa)
		r.Ca.Crt = string(bytes)
		resp, _ := json.Marshal(r)
		w.Write(resp)
	} else {
		w.Header().Add("Content-Type", "text/plain")
		w.Write(bytes)
	}
}
