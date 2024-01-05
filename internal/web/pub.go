package web

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"simpleca/internal/key"
	"simpleca/tools"
)

func Pub(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm()
	var err error

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable ro read request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	passphrase := GetParam(r, "passphrase", "")

	if kkey, err := key.LoadPrivateKey(body, passphrase); err != nil {
		http.Error(w, "Unable to convert to private key: "+err.Error(), http.StatusBadRequest)
	} else {
		switch kkey.(type) {
		case *rsa.PrivateKey:
			publickey := &kkey.(*rsa.PrivateKey).PublicKey
			publicKeyBytes, err := x509.MarshalPKIXPublicKey(publickey)
			if err != nil {
				http.Error(w, "Unable to convert to public key: "+err.Error(), http.StatusInternalServerError)
				return
			}
			publicKeyBlock := &pem.Block{
				Type:  "PUBLIC KEY",
				Bytes: publicKeyBytes,
			}
			bytes := pem.EncodeToMemory(publicKeyBlock)
			if tools.Contains(r.Header["Accept"], "application/json") {
				w.Header().Add("Content-type", "application/json")
				r := Resp{}
				r.Key = new(TKey)
				r.Key.Pub = string(bytes)
				resp, _ := json.Marshal(r)
				w.WriteHeader(http.StatusOK)
				w.Write(resp)
			} else {
				w.Header().Add("Content-type", "text/plain")
				w.WriteHeader(http.StatusOK)
				w.Write(bytes)
			}
		case *ecdsa.PrivateKey:
			publickey := &kkey.(*ecdsa.PrivateKey).PublicKey
			publicKeyBytes, err := x509.MarshalPKIXPublicKey(publickey)
			if err != nil {
				http.Error(w, "Unable to convert to public key: "+err.Error(), http.StatusInternalServerError)
				return
			}
			publicKeyBlock := &pem.Block{
				Type:  "PUBLIC KEY",
				Bytes: publicKeyBytes,
			}
			bytes := pem.EncodeToMemory(publicKeyBlock)
			if tools.Contains(r.Header["Accept"], "application/json") {
				w.Header().Add("Content-type", "application/json")
				r := Resp{}
				r.Key = new(TKey)
				r.Key.Pub = string(bytes)
				resp, _ := json.Marshal(r)
				w.WriteHeader(http.StatusOK)
				w.Write(resp)
			} else {
				w.Header().Add("Content-type", "text/plain")
				w.WriteHeader(http.StatusOK)
				w.Write(bytes)
			}
		default:
			http.Error(w, "Wrong key type", http.StatusBadRequest)
		}
	}
}
