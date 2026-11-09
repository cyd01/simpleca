package web

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"simpleca/internal/ca"
	"simpleca/internal/key"
	"strconv"
	"strings"
)

const maxBodyBytes = 1 << 20

var (
	// maxKeySize is raised at startup if the operator picks a larger -size.
	maxKeySize = 4096
	// keyWorkSlots bounds concurrent key generation and passphrase handling.
	keyWorkSlots = make(chan struct{}, runtime.NumCPU())
)

// acquireKeyWork reserves a slot for expensive key work, or answers 503.
// On success the caller must call releaseKeyWork.
func acquireKeyWork(w http.ResponseWriter) bool {
	select {
	case keyWorkSlots <- struct{}{}:
		return true
	default:
		w.Header().Set("Retry-After", "5")
		http.Error(w, "Server busy", http.StatusServiceUnavailable)
		return false
	}
}

func releaseKeyWork() {
	<-keyWorkSlots
}

func checkKeySize(w http.ResponseWriter, size int) bool {
	if size < key.MinRSABits {
		http.Error(w, "Key size not big enough", http.StatusBadRequest)
		return false
	}
	if size > maxKeySize {
		http.Error(w, "Key size too much big", http.StatusBadRequest)
		return false
	}
	return true
}

// requestedDays reads the validity requested for a certificate, or answers 400.
func requestedDays(w http.ResponseWriter, r *http.Request) (int, bool) {
	days, err := strconv.Atoi(GetParam(r, "days", strconv.Itoa(ConfigDays)))
	if err != nil {
		http.Error(w, "Wrong number of days", http.StatusBadRequest)
		return 0, false
	}
	if days < 1 {
		http.Error(w, "Number of days too small", http.StatusBadRequest)
		return 0, false
	}
	if days > ca.MaxCertDays {
		http.Error(w, "Number of days too big, maximum is "+strconv.Itoa(ca.MaxCertDays), http.StatusBadRequest)
		return 0, false
	}
	return days, true
}

// readBody reads a request body limited by the server, or answers with an error.
func readBody(w http.ResponseWriter, r *http.Request) ([]byte, bool) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		if err.Error() == "http: request body too large" {
			http.Error(w, "Request too large", http.StatusRequestEntityTooLarge)
		} else {
			http.Error(w, "Unable to read request: "+err.Error(), http.StatusBadRequest)
		}
		return nil, false
	}
	return body, true
}

// resolvePath returns an absolute path with symbolic links resolved. The file
// itself may not exist yet.
func resolvePath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		return resolved, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(abs))
	if err != nil {
		return "", err
	}
	return filepath.Join(parent, filepath.Base(abs)), nil
}

// insideDir tells whether file is, or would be, inside dir.
func insideDir(file, dir string) (bool, error) {
	resolvedFile, err := resolvePath(file)
	if err != nil {
		return false, err
	}
	resolvedDir, err := resolvePath(dir)
	if err != nil {
		return false, err
	}
	rel, err := filepath.Rel(resolvedDir, resolvedFile)
	if err != nil {
		return false, nil // different volumes
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)), nil
}
