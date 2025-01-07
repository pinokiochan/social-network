package middleware

import (
	"net/http"
	"sync"
	"time"
	"log"
)

type visitor struct {
	lastSeen time.Time
	count    int
}

var (
	visitors = make(map[string]*visitor)
	mu       sync.Mutex
)



// LoggingMiddleware logs details of each incoming request
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received %s request for %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}


func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		mu.Lock()
		v, exists := visitors[ip]
		if !exists {
			visitors[ip] = &visitor{time.Now(), 1}
			mu.Unlock()
			next.ServeHTTP(w, r)
			return
		}

		// Reset count if last request was more than a minute ago
		if time.Since(v.lastSeen) > time.Minute {
			v.count = 0
		}

		if v.count > 60 { // 60 requests per minute limit
			mu.Unlock()
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		v.count++
		v.lastSeen = time.Now()
		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
