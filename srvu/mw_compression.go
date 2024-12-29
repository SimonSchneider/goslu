package srvu

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type GzipRW interface {
	http.ResponseWriter
	Close() error
}

type GzipResponseWriter struct {
	http.ResponseWriter
	gz *gzip.Writer
}

func NewGzipResponseWriter(w http.ResponseWriter) *GzipResponseWriter {
	return &GzipResponseWriter{ResponseWriter: w, gz: gzip.NewWriter(w)}
}

func (w *GzipResponseWriter) Write(b []byte) (int, error) {
	return w.gz.Write(b)
}

func (w *GzipResponseWriter) Close() error {
	return w.gz.Close()
}

type GzipFlushingResponseWriter struct {
	*GzipResponseWriter
	http.Flusher
}

func NewGzipFlushingResponseWriter(w *GzipResponseWriter, flusher http.Flusher) *GzipFlushingResponseWriter {
	return &GzipFlushingResponseWriter{GzipResponseWriter: w, Flusher: flusher}
}

func (w *GzipFlushingResponseWriter) Flush() {
	w.GzipResponseWriter.gz.Flush()
	w.Flusher.Flush()
}

func NewGzipResponseWriterWithAutoDetection(w http.ResponseWriter) GzipRW {
	gw := NewGzipResponseWriter(w)
	if f, ok := w.(http.Flusher); ok {
		return NewGzipFlushingResponseWriter(gw, f)
	}
	return gw
}

func WithCompression() Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				w.Header().Set("Content-Encoding", "gzip")
				gz := NewGzipResponseWriterWithAutoDetection(w)
				defer gz.Close()
				h.ServeHTTP(gz, r)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}
