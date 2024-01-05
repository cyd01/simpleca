package web

import (
	"encoding/json"
	"encoding/pem"
	"net/http"
	"simpleca/internal/key"
	"simpleca/tools"
	"strconv"
	"strings"
)

func Key(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	var err error
	var size int = ConfigSize

	q := r.URL.Query()

	s := q.Get("size")
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

	passphrase := GetParam(r, "passphrase", "")
	t := strings.ToLower(GetParam(r, "type", ConfigKeyType))

	w.Header().Add("Cache-control", "no-cache, no-store, must-revalidate")
	w.Header().Add("Expires", "0")
	switch t {
	case "rsa":
		if key, err := key.GenerateRSAKeyBlock(size, passphrase); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error while generate private key"))
			return
		} else {
			bytes := pem.EncodeToMemory(key)
			if tools.Contains(r.Header["Accept"], "application/json") {
				w.Header().Add("Content-type", "application/json")
				r := Resp{}
				r.Key = new(TKey)
				r.Key.Priv = string(bytes)
				resp, _ := json.Marshal(r)
				w.WriteHeader(http.StatusOK)
				w.Write(resp)
			} else {
				w.Header().Add("Content-type", "text/plain")
				w.WriteHeader(http.StatusOK)
				w.Write(bytes)
			}
		}
		return
	case "ecdsa":
		if key, err := key.GenerateECDSAKeyBlock(passphrase); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error while generate private key"))
			return
		} else {
			bytes := pem.EncodeToMemory(key)
			if tools.Contains(r.Header["Accept"], "application/json") {
				w.Header().Add("Content-type", "application/json")
				r := Resp{}
				r.Key = new(TKey)
				r.Key.Priv = string(bytes)
				resp, _ := json.Marshal(r)
				w.WriteHeader(http.StatusOK)
				w.Write(resp)
			} else {
				w.Header().Add("Content-type", "text/plain")
				w.WriteHeader(http.StatusOK)
				w.Write(bytes)
			}
		}
		return
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Wrong key type"))
		return
	}
}
