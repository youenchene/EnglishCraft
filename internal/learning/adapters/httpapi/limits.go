package httpapi

import (
	"net/http"
	"sync"
	"time"
)

// Global request admission is bounded in memory and independent of spoofable
// proxy headers. Per-student consumption admission is persistent in PostgreSQL.
func SensitiveRateLimit(next http.Handler) http.Handler {
	var mu sync.Mutex
	start := time.Now()
	count := 0
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" || r.Method == "HEAD" {
			next.ServeHTTP(w, r)
			return
		}
		mu.Lock()
		if time.Since(start) >= time.Minute {
			start = time.Now()
			count = 0
		}
		count++
		allowed := count <= 120
		mu.Unlock()
		if !allowed {
			w.Header().Set("Retry-After", "60")
			writeError(w, 429, "rate_limit", "Trop de demandes. Réessaie dans une minute.")
			return
		}
		next.ServeHTTP(w, r)
	})
}
