package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/pinokiochan/social-network/internal/logger"
	"github.com/sirupsen/logrus"
)

type visitor struct {
	lastSeen time.Time
	count    int
}

var (
	visitors = make(map[string]*visitor)
	mu       sync.Mutex
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		logger.Log.WithFields(logrus.Fields{
			"method":     r.Method,
			"path":       r.URL.Path,
			"ip":         r.RemoteAddr,
			"user_agent": r.UserAgent(),
		}).Info("Incoming request")

		next.ServeHTTP(w, r)

		duration := time.Since(start)
		logger.Log.WithFields(logrus.Fields{
			"method":      r.Method,
			"path":        r.URL.Path,
			"duration_ms": duration.Milliseconds(),
		}).Info("Request completed")
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
			
			logger.Log.WithFields(logrus.Fields{
				"ip": ip,
			}).Debug("New visitor")
			
			next.ServeHTTP(w, r)
			return
		}

		if time.Since(v.lastSeen) > time.Minute {
			v.count = 0
			logger.Log.WithFields(logrus.Fields{
				"ip": ip,
			}).Info("Rate limit counter reset")
		}

		if v.count > 60 {
			mu.Unlock()
			logger.Log.WithFields(logrus.Fields{
				"ip":    ip,
				"count": v.count,
			}).Warn("Rate limit exceeded")
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		v.count++
		v.lastSeen = time.Now()
		mu.Unlock()

		logger.Log.WithFields(logrus.Fields{
			"ip":    ip,
			"count": v.count,
		}).Debug("Request within rate limit")

		next.ServeHTTP(w, r)
	})
}

func ErrorHandlingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Log.WithFields(logrus.Fields{
					"error":  fmt.Sprintf("%v", err),
					"path":   r.URL.Path,
					"method": r.Method,
				}).Error("Panic recovered in request handler")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

