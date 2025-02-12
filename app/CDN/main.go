package main

import (
	"app/internal/cache"
	"app/internal/loadbalancer"
	"app/internal/metrics"
	"app/internal/middleware" // Ajout de l'import du package api
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

var log = logrus.New()

func init() {
	// Configuration de logrus
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)
}

func main() {
	// Configuration du cache avec une taille maximale de 1000 entrées
	memCache, err := cache.NewMemoryCache(1000)
	if err != nil {
		log.WithError(err).Fatal("Erreur initialisation cache")
	}

	// Configuration du Load Balancer en mode Weighted Round Robin
	// avec deux backends de même poids
	backends := []string{"http://backend:8080","http://backend:8080","http://backend:8080","http://backend:8080","http://backend:8080"}
	weights := []int{1,1,1,1,1}
	lb := loadbalancer.NewWeightedRoundRobin(backends, weights, loadbalancer.Config{
		HealthCheckInterval: time.Second,
		HealthCheckTimeout:  time.Second,
		MaxFailCount:       3,
		RetryTimeout:       time.Second,
	})

	// Configuration du Rate Limiter (100 requêtes par minute par IP)
	rateLimiter := middleware.NewRateLimiter(rate.Limit(100/60.0), 100)

	// Configuration du routeur HTTP
	mux := http.NewServeMux()

	// Endpoint de monitoring
	mux.Handle("/metrics", promhttp.Handler())

	// Health check endpoints
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("healthy"))
	})

	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ready"))
	})

	// Endpoint pour vider le cache
	mux.HandleFunc("/cache/purge", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		memCache.Clear()
		log.Info("Cache vidé avec succès")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Cache purgé"))
	})

	// Route principale avec middleware de sécurité
	mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := fmt.Sprintf("%d", time.Now().UnixNano())
		requestIDInt, _ := strconv.ParseInt(requestID, 10, 64)

		// Logger la requête entrante
		log.WithFields(logrus.Fields{
			"request_id": requestID,
			"method":     r.Method,
			"path":       r.URL.Path,
			"client_ip":  r.RemoteAddr,
		}).Info("Requête entrante reçue")

		// Vérification du cache uniquement pour les requêtes GET
		if r.Method == http.MethodGet {
			if cachedResponse, found, err := memCache.Get(r.Context(), r.URL.Path); err == nil && found {
				metrics.CacheHits.Inc()
				log.WithFields(logrus.Fields{
					"request_id": requestID,
					"path":      r.URL.Path,
					"source":    "cache",
				}).Info("Réponse servie depuis le cache")
				w.Write(cachedResponse.Value.([]byte))
				return
			}
			metrics.CacheMisses.Inc()
		}

		// Sélection du backend
		backend, err := lb.NextBackend(r.Context())
		if err != nil {
			metrics.RecordRequest(r.Method, r.URL.Path, http.StatusServiceUnavailable, time.Since(start).Seconds(), 0)
			log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":     err,
			}).Error("Aucun backend disponible")
			http.Error(w, "No backend available", http.StatusServiceUnavailable)
			return
		}

		// Créer une nouvelle requête pour le backend
		backendReq, err := http.NewRequestWithContext(r.Context(), r.Method, backend.URL+r.URL.Path, r.Body)
		if err != nil {
			metrics.RecordRequest(r.Method, r.URL.Path, http.StatusBadGateway, time.Since(start).Seconds(), 0)
			log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":     err,
			}).Error("Erreur création requête backend")
			http.Error(w, "Backend error", http.StatusBadGateway)
			return
		}

		// Copier les headers
		for k, v := range r.Header {
			backendReq.Header[k] = v
		}

		// Envoyer la requête au backend
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(backendReq)
		if err != nil {
			metrics.RecordBackendRequest(backend.URL, time.Since(start).Seconds(), err)
			log.WithFields(logrus.Fields{
				"request_id": requestID,
				"backend":    backend.URL,
				"error":     err,
			}).Error("Erreur requête backend")
			http.Error(w, "Backend error", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Copier la réponse
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			metrics.RecordRequest(r.Method, r.URL.Path, http.StatusInternalServerError, time.Since(start).Seconds(), 0)
			log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":     err,
			}).Error("Erreur lecture réponse")
			http.Error(w, "Error reading response", http.StatusInternalServerError)
			return
		}

		// Mettre en cache si c'est une requête GET
		if r.Method == http.MethodGet && resp.StatusCode == http.StatusOK {
			if err := memCache.Set(r.Context(), r.URL.Path, body, map[string]string{}, 1*time.Hour); err != nil {
				log.WithFields(logrus.Fields{
					"request_id": requestID,
					"error":     err,
				}).Error("Erreur mise en cache")
			}
		}

		// Copier les headers de réponse
		for k, v := range resp.Header {
			w.Header()[k] = v
		}
		w.WriteHeader(resp.StatusCode)
		w.Write(body)

		// Enregistrer les métriques finales
		metrics.RecordRequest(r.Method, r.URL.Path, resp.StatusCode, time.Since(start).Seconds(), int64(len(body)))
		metrics.RecordBackendRequest(backend.URL, time.Since(start).Seconds(), nil)

		log.WithFields(logrus.Fields{
			"request_id":   requestID,
			"status_code": resp.StatusCode,
			"backend":      backend.URL,
			"elapsed_time": time.Since(time.Unix(0, requestIDInt)).String(),
		}).Info("Requête terminée")
	})

	// Application des middlewares
	handler := middleware.SecurityHeaders(rateLimiter.RateLimit(mainHandler))
	mux.Handle("/", handler)

	// Configuration du serveur
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Canal pour les signaux d'arrêt
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Démarrage du serveur dans une goroutine
	go func() {
		log.WithFields(logrus.Fields{
			"address": srv.Addr,
			"pid":     os.Getpid(),
		}).Info("Démarrage du serveur CDN")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("Erreur démarrage serveur")
		}
	}()

	// Attente du signal d'arrêt
	<-stop
	log.Info("Arrêt du serveur en cours...")

	// Création d'un contexte avec timeout pour l'arrêt gracieux
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.WithError(err).Error("Erreur lors de l'arrêt du serveur")
	}

	// Fermeture du load balancer
	if err := lb.Close(); err != nil {
		log.WithError(err).Error("Erreur lors de la fermeture du load balancer")
	}

	log.Info("Serveur arrêté avec succès")
}
