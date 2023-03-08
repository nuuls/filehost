package api

import (
	"net/http"
	"sync"
	"time"
)

// RatelimitMiddleware limits new requests when encountering too many 404 errors to prevent enumeration attacks
func RatelimitMiddleware(maxFails int, resetInterval time.Duration) func(next http.Handler) http.Handler {
	var mu sync.Mutex
	requests := map[string]int{}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr

			mu.Lock()
			limited := requests[ip] >= maxFails
			mu.Unlock()

			if limited {
				http.Error(w, "Ratelimited", 429)
				return
			}

			wrapped := &wrapWriter{ResponseWriter: w}
			next.ServeHTTP(wrapped, r)
			if wrapped.statusCode != 404 {
				return
			}

			mu.Lock()
			requests[ip]++
			mu.Unlock()

			time.AfterFunc(resetInterval, func() {
				mu.Lock()
				requests[ip]--
				if requests[ip] == 0 {
					delete(requests, ip)
				}
				mu.Unlock()
			})
		})
	}
}

type wrapWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrapWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}
