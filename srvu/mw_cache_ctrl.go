package srvu

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func WithCacheCtrlHeader(ttl time.Duration) Middleware {
	if ttl <= 0 {
		panic("ttl must be greater than 0")
	}
	cacheControl := fmt.Sprintf("public, max-age=%d", int64(ttl.Seconds()))
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("serving %s", r.URL.Path)
			w.Header().Set("Cache-Control", cacheControl)
			h.ServeHTTP(w, r)
		})
	}
}
