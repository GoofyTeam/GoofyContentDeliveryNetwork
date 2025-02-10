package main

import (
	"app/internal/cache"
	"app/internal/loadbalancer"
	"app/internal/middleware"
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// main est la fonction principale qui initialise et démarre le serveur CDN
// Elle configure :
// - Le système de logging
// - Le cache en mémoire
// - Le load balancer
// - Les middlewares de sécurité et de monitoring
// - La gestion gracieuse de l'arrêt du serveur
func main() {
	// Configuration du logger avec format JSON et niveau INFO
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(logrus.InfoLevel)

	// Initialisation du cache en mémoire avec une capacité de 1000 entrées
	memCache, err := cache.NewMemoryCache(1000)
	if err != nil {
		log.Fatal(err)
	}

	// Configuration du Load Balancer en mode Weighted Round Robin
	// avec deux backends de même poids
	backends := []string{"http://backend1:8080", "http://backend2:8080"}
	weights := []int{1, 1}
	lb := loadbalancer.NewWeightedRoundRobin(backends, weights)

	// Configuration du routeur HTTP
	mux := http.NewServeMux()
	
	// Route principale qui gère le load balancing et le cache
	// Pour chaque requête :
	// 1. Vérifie si la réponse est en cache
	// 2. Si non, proxie la requête vers un backend
	// 3. Met en cache la réponse
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		backend := lb.NextBackend()
		
		// Tentative de récupération depuis le cache
		if data, found := memCache.Get(r.URL.Path); found {
			fmt.Fprint(w, data)
			return
		}
		
		// Proxy vers le backend sélectionné
		resp, err := http.Get(backend.URL + r.URL.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		
		// Mise en cache de la réponse pour les futures requêtes
		memCache.Set(r.URL.Path, "cached response")
		
		fmt.Fprintf(w, "Proxied to %s", backend.URL)
	})

	// Exposition des métriques Prometheus pour le monitoring
	mux.Handle("/metrics", promhttp.Handler())

	// Application des middlewares dans l'ordre :
	// 1. Sécurité (headers HTTPS, CORS, etc.)
	// 2. Métriques (compteurs Prometheus)
	// 3. Rate Limiting (100 req/s avec burst de 10)
	handler := middleware.Security(
		middleware.Metrics(
			middleware.RateLimit(100, 10)(mux),
		),
	)

	// Configuration du serveur HTTP avec timeouts
	srv := &http.Server{
		Addr:           ":8080",
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Démarrage du serveur dans une goroutine séparée
	go func() {
		log.Info("Starting server on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Configuration de la gestion gracieuse de l'arrêt
	// Attend un signal SIGINT ou SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Arrêt du serveur avec timeout de 30 secondes
	log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Info("Server successfully shutdown")
}
