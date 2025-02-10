# CDN Go - Projet de Content Delivery Network

Ce projet implémente un Content Delivery Network (CDN) en Go, conçu pour optimiser la distribution de contenu web avec des fonctionnalités avancées de mise en cache, de répartition de charge et de monitoring.

## 🚀 Fonctionnalités

- **Proxy HTTP** : Redirection intelligente des requêtes
- **Système de Cache** : 
  - Cache LRU en mémoire
  - Support Redis pour le cache distribué
- **Load Balancing** : 
  - Round Robin
  - Weighted Round Robin
  - Least Connections
- **Sécurité** :
  - Rate Limiting
  - Protection DDoS
  - Headers de sécurité HTTP
- **Monitoring** :
  - Métriques Prometheus
  - Visualisation Grafana
  - Logging structuré avec Logrus

## 🛠 Prérequis

- Docker
- Docker Compose
- Go 1.23+ (pour le développement local)

## 🚦 Démarrage

1. **Mode Développement** :
```bash
docker compose -f docker-compose.dev.yml up -d
```
- Hot-reload activé
- Accessible sur http://localhost:8080
- Métriques sur http://localhost:8080/metrics

2. **Mode Production** :
```bash
docker compose -f docker-compose.prod.yml up -d
```
- Optimisé pour la production
- Accessible sur http://localhost:8081
- Métriques sur http://localhost:8081/metrics

3. **Services additionnels** :
- Grafana : http://localhost:3000 (admin/admin)
- Prometheus : http://localhost:9090
- Redis : localhost:6379

## 🏗 Structure du Projet

```
app/
├── internal/
│   ├── cache/          # Implémentation du cache (LRU, Redis)
│   ├── loadbalancer/   # Algorithmes de load balancing
│   └── middleware/     # Middlewares (sécurité, métriques)
├── pkg/
│   └── config/         # Configuration de l'application
└── main.go            # Point d'entrée de l'application
```

## 🔍 Fonctionnement Détaillé

### 1. Système de Cache
- **Cache LRU** (`internal/cache/cache.go`) :
  - Implémente l'interface `Cache`
  - Utilise `hashicorp/golang-lru` pour la gestion du cache en mémoire
  - Limite configurable de la taille du cache

### 2. Load Balancer
- **Implémentations** (`internal/loadbalancer/loadbalancer.go`) :
  - `RoundRobin` : Distribution cyclique des requêtes
  - `WeightedRoundRobin` : Distribution pondérée selon la capacité des serveurs
  - `LeastConnections` : Envoi vers le serveur le moins chargé

### 3. Middlewares
- **Sécurité** (`internal/middleware/middleware.go`) :
  - Rate limiting avec `golang.org/x/time/rate`
  - Headers de sécurité HTTP
  - Protection contre les attaques courantes

### 4. Monitoring
- **Métriques** :
  - Temps de réponse des requêtes
  - Nombre de requêtes par endpoint
  - Taux de succès/erreur
  - Utilisation du cache

### 5. Application Principale
Le fichier `main.go` orchestre tous ces composants :
1. Initialise le logger et le cache
2. Configure le load balancer
3. Met en place les middlewares de sécurité et monitoring
4. Démarre le serveur HTTP avec gestion gracieuse de l'arrêt

## 📊 Monitoring

### Métriques disponibles :
- `http_duration_seconds` : Temps de réponse des requêtes
- `http_requests_total` : Nombre total de requêtes par endpoint
- Visualisation dans Grafana via Prometheus

## 🔒 Sécurité

- Rate limiting : 100 requêtes/seconde par défaut
- Headers de sécurité :
  - `X-Frame-Options`
  - `X-Content-Type-Options`
  - `X-XSS-Protection`
  - `Content-Security-Policy`
  - `Strict-Transport-Security`