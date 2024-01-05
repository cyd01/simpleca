package web

import (
	"html"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Middleware logguer (https://blog.questionable.services/article/guide-logging-middleware-go/)
// responseWriter is a minimal wrapper for http.ResponseWriter that allows the
// written HTTP status code to be captured for logging.
type responseWriter struct {
	http.ResponseWriter
	status      int
	contentType string
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, status: 200, contentType: "application/octet-stream", wroteHeader: false}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
	return
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.contentType = rw.Header().Get("Content-Type")
	rw.wroteHeader = true
	return rw.ResponseWriter.Write(b)
}

// Logger is a middleware handler that does request logging
type Logger struct {
	handler http.Handler
}

// MetricsAndLogs is a middleware handler that does Prometheus metrics counts, and write logs
func Logs(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := wrapResponseWriter(w)
		next.ServeHTTP(wrapped, r)

		PrintLogs(r, wrapped.status, time.Now().Sub(start), wrapped.contentType)
		wrapped = nil
	})
}

func PrintLogs(r *http.Request, status int, dur time.Duration, s ...string) {
	ra := r.RemoteAddr
	forward := r.Header.Get("X-Forwarded-For")
	if len(forward) > 0 {
		ra = ra + "," + forward
	}
	ua := r.Header.Get("User-Agent")
	if len(ua) == 0 {
		ua = "No User-Agent"
	}
	if len(s) > 0 {
		log.Println(r.Method, ra, html.EscapeString(r.URL.String()), "["+ua+"]", strconv.Itoa(status), dur.String(), s)
	} else {
		log.Println(r.Method, ra, html.EscapeString(r.URL.String()), "["+ua+"]", strconv.Itoa(status), dur.String())
	}
}

//ServeHTTP handles the request by passing it to the real
