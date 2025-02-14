// Package config définit la structure de configuration du CDN
// Il gère les paramètres du serveur, du cache, du load balancer et du monitoring
package config

import "time"

// Config représente la configuration globale de l'application
type Config struct {
	// Server contient la configuration du serveur HTTP
	Server struct {
		Port            int           `yaml:"port"`            // Port d'écoute du serveur
		ReadTimeout     time.Duration `yaml:"readTimeout"`     // Timeout pour la lecture des requêtes
		WriteTimeout    time.Duration `yaml:"writeTimeout"`    // Timeout pour l'écriture des réponses
		MaxHeaderBytes  int           `yaml:"maxHeaderBytes"`  // Taille maximale des headers HTTP
		TLSCertFile    string        `yaml:"tlsCertFile"`     // Chemin vers le certificat TLS
		TLSKeyFile     string        `yaml:"tlsKeyFile"`      // Chemin vers la clé privée TLS
	}

	// Cache définit la configuration du système de cache
	Cache struct {
		Type       string `yaml:"type"`      // Type de cache : "memory" ou "redis"
		Size       int    `yaml:"size"`      // Taille du cache en mémoire (nombre d'entrées)
		RedisURL   string `yaml:"redisUrl"`  // URL du serveur Redis (si type="redis")
		RedisDB    int    `yaml:"redisDb"`   // Index de la base de données Redis
	}

	// LoadBalancer configure la répartition de charge
	LoadBalancer struct {
		Type     string   `yaml:"type"`      // Stratégie : "round-robin", "weighted", "least-conn"
		Backends []string `yaml:"backends"`   // Liste des URLs des serveurs backend
		Weights  []int    `yaml:"weights"`   // Poids pour la stratégie weighted round-robin
	}

	// RateLimit définit les limites de taux de requêtes
	RateLimit struct {
		RequestsPerSecond float64 `yaml:"requestsPerSecond"` // Nombre max de requêtes par seconde
		Burst            int     `yaml:"burst"`              // Nombre de requêtes autorisées en rafale
	}

	// Monitoring configure les outils de surveillance
	Monitoring struct {
		PrometheusPort int    `yaml:"prometheusPort"` // Port pour les métriques Prometheus
		LogLevel       string `yaml:"logLevel"`        // Niveau de log (debug, info, warn, error)
	}
}
