package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Métriques du cache
	CacheHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cdn_cache_hits_total",
		Help: "Nombre total de hits du cache",
	})

	CacheMisses = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cdn_cache_misses_total",
		Help: "Nombre total de misses du cache",
	})

	CacheSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "cdn_cache_size_bytes",
		Help: "Taille totale du cache en bytes",
	})

	// Métriques du load balancer
	BackendRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cdn_backend_requests_total",
		Help: "Nombre total de requêtes par backend",
	}, []string{"backend_url"})

	BackendLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "cdn_backend_latency_seconds",
		Help:    "Latence des requêtes par backend",
		Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5},
	}, []string{"backend_url"})

	BackendErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cdn_backend_errors_total",
		Help: "Nombre total d'erreurs par backend",
	}, []string{"backend_url", "error_type"})

	ActiveBackends = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "cdn_active_backends",
		Help: "Nombre de backends actifs",
	})

	// Métriques de sécurité
	RateLimitExceeded = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cdn_rate_limit_exceeded_total",
		Help: "Nombre total de requêtes ayant dépassé la limite de taux",
	})

	DDoSAttempts = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cdn_ddos_attempts_total",
		Help: "Nombre total de tentatives de DDoS détectées",
	})

	// Métriques HTTP
	HttpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cdn_http_requests_total",
		Help: "Nombre total de requêtes HTTP",
	}, []string{"method", "path", "status"})

	HttpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "cdn_http_request_duration_seconds",
		Help:    "Durée des requêtes HTTP",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	HttpResponseSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "cdn_http_response_size_bytes",
		Help:    "Taille des réponses HTTP en bytes",
		Buckets: []float64{100, 1000, 10000, 100000, 1000000},
	}, []string{"method", "path"})

	// Métriques système
	CpuUsage = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "cdn_cpu_usage_percent",
		Help: "Utilisation CPU en pourcentage",
	})

	MemoryUsage = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "cdn_memory_usage_bytes",
		Help: "Utilisation mémoire en bytes",
	})

	OpenConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "cdn_open_connections",
		Help: "Nombre de connexions ouvertes",
	})
)

// RecordRequest enregistre les métriques d'une requête HTTP
func RecordRequest(method, path string, status int, duration float64, size int64) {
	HttpRequestsTotal.WithLabelValues(method, path, string(status)).Inc()
	HttpRequestDuration.WithLabelValues(method, path).Observe(duration)
	HttpResponseSize.WithLabelValues(method, path).Observe(float64(size))
}

// RecordBackendRequest enregistre les métriques d'une requête backend
func RecordBackendRequest(backendURL string, duration float64, err error) {
	BackendRequests.WithLabelValues(backendURL).Inc()
	BackendLatency.WithLabelValues(backendURL).Observe(duration)
	
	if err != nil {
		BackendErrors.WithLabelValues(backendURL, err.Error()).Inc()
	}
}

// UpdateCacheMetrics met à jour les métriques du cache
func UpdateCacheMetrics(hits, misses uint64, size int64) {
	CacheHits.Add(float64(hits))
	CacheMisses.Add(float64(misses))
	CacheSize.Set(float64(size))
}

// UpdateSystemMetrics met à jour les métriques système
func UpdateSystemMetrics(cpuPercent float64, memoryBytes int64, connections int) {
	CpuUsage.Set(cpuPercent)
	MemoryUsage.Set(float64(memoryBytes))
	OpenConnections.Set(float64(connections))
}

// UpdateActiveBackends met à jour le nombre de backends actifs
func UpdateActiveBackends(count int32) {
	ActiveBackends.Set(float64(count))
}

// RecordSecurityEvent enregistre les événements de sécurité
func RecordSecurityEvent(eventType string) {
	switch eventType {
	case "rate_limit":
		RateLimitExceeded.Inc()
	case "ddos":
		DDoSAttempts.Inc()
	}
}
