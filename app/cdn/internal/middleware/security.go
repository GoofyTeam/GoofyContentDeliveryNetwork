package middleware

import (
	"golang.org/x/time/rate"
	"net/http"
	"sync"
)

// RateLimiter implémente la protection contre les attaques DDoS
// en limitant le nombre de requêtes par IP
type RateLimiter struct {
	visitors map[string]*rate.Limiter
	mu      sync.RWMutex
	rate    rate.Limit
	burst   int
}

// NewRateLimiter crée un nouveau limiteur de taux avec un taux et un burst spécifiés
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]*rate.Limiter),
		rate:    r,
		burst:   b,
	}
}

// getLimiter retourne le rate limiter pour une IP donnée
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.visitors[ip] = limiter
	}

	return limiter
}

// RateLimit est un middleware qui limite le taux de requêtes par IP
func (rl *RateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		limiter := rl.getLimiter(ip)

		if !limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// SecurityHeaders ajoute des en-têtes de sécurité à la réponse
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Protection XSS
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		// Protection contre le clickjacking
		w.Header().Set("X-Frame-Options", "DENY")
		// Protection contre le MIME sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// HSTS (forcer HTTPS)
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		// CSP
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		next.ServeHTTP(w, r)
	})
}
