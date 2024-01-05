package web

import (
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"simpleca/internal/ca"
	"simpleca/internal/cert"
	"simpleca/internal/csr"
	"simpleca/tools"
	"strconv"
)

func Sign(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	var err error
	var days int = ConfigDays
	q := r.URL.Query()
	s := q.Get("days")
	if len(s) > 0 {
		if days, err = strconv.Atoi(s); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Wrong number of days"))
			return
		}
	}
	if days < 1 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Number of days too small"))
		return
	}
	// Reading the request body
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable ro read request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if ccsr, err := csr.LoadCSR(body); err != nil {
		http.Error(w, "Unable to convert to certificate signing request: "+err.Error(), http.StatusBadRequest)
	} else {
		crt, err := ca.CASign(ccsr, days, CaCert, CaKey)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Can not sign certificate signing request"))
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
