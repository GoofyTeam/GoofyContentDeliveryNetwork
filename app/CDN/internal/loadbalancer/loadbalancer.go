// Package loadbalancer fournit différentes stratégies de répartition de charge
// pour distribuer le trafic entre plusieurs serveurs backend
package loadbalancer

import (
	"sync"
	"sync/atomic"
)

// Backend représente un serveur backend avec ses propriétés
type Backend struct {
	URL           string // URL du serveur backend
	Weight        int    // Poids pour l'algorithme weighted round-robin
	CurrentWeight int    // Poids actuel (utilisé dans l'algorithme weighted round-robin)
	Connections   int32  // Nombre de connexions actives (utilisé pour least connections)
}

// LoadBalancer définit l'interface commune pour toutes les stratégies de load balancing
type LoadBalancer interface {
	// NextBackend retourne le prochain backend à utiliser selon la stratégie choisie
	NextBackend() *Backend
}

// RoundRobin implémente la stratégie de répartition cyclique simple
type RoundRobin struct {
	backends []*Backend     // Liste des backends disponibles
	current  uint32         // Index du backend courant (accès atomique)
}

// NewRoundRobin crée une nouvelle instance de RoundRobin
// urls: liste des URLs des serveurs backend
func NewRoundRobin(urls []string) *RoundRobin {
	backends := make([]*Backend, len(urls))
	for i, url := range urls {
		backends[i] = &Backend{URL: url}
	}
	return &RoundRobin{backends: backends}
}

// NextBackend retourne le prochain backend dans l'ordre cyclique
// Utilise des opérations atomiques pour être thread-safe
func (r *RoundRobin) NextBackend() *Backend {
	next := atomic.AddUint32(&r.current, 1) % uint32(len(r.backends))
	return r.backends[next]
}

// WeightedRoundRobin implémente la stratégie de répartition pondérée
type WeightedRoundRobin struct {
	backends []*Backend  // Liste des backends avec leurs poids
	mu      sync.Mutex  // Mutex pour la synchronisation
}

// NewWeightedRoundRobin crée une nouvelle instance de WeightedRoundRobin
// urls: liste des URLs des serveurs backend
// weights: poids correspondants pour chaque serveur
func NewWeightedRoundRobin(urls []string, weights []int) *WeightedRoundRobin {
	backends := make([]*Backend, len(urls))
	for i, url := range urls {
		backends[i] = &Backend{
			URL:           url,
			Weight:        weights[i],
			CurrentWeight: weights[i],
		}
	}
	return &WeightedRoundRobin{backends: backends}
}

// NextBackend implémente l'algorithme de weighted round-robin
// Sélectionne le backend avec le plus grand poids actuel
func (w *WeightedRoundRobin) NextBackend() *Backend {
	w.mu.Lock()
	defer w.mu.Unlock()

	var best *Backend
	var totalWeight int

	for _, b := range w.backends {
		b.CurrentWeight += b.Weight
		totalWeight += b.Weight
		if best == nil || b.CurrentWeight > best.CurrentWeight {
			best = b
		}
	}

	best.CurrentWeight -= totalWeight
	return best
}

// LeastConnections implémente la stratégie du nombre minimum de connexions
type LeastConnections struct {
	backends []*Backend // Liste des backends
}

// NewLeastConnections crée une nouvelle instance de LeastConnections
// urls: liste des URLs des serveurs backend
func NewLeastConnections(urls []string) *LeastConnections {
	backends := make([]*Backend, len(urls))
	for i, url := range urls {
		backends[i] = &Backend{URL: url}
	}
	return &LeastConnections{backends: backends}
}

// NextBackend sélectionne le backend ayant le moins de connexions actives
// Utilise des opérations atomiques pour le comptage des connexions
func (l *LeastConnections) NextBackend() *Backend {
	var best *Backend
	var minConn int32 = -1

	for _, b := range l.backends {
		conn := atomic.LoadInt32(&b.Connections)
		if minConn == -1 || conn < minConn {
			minConn = conn
			best = b
		}
	}

	atomic.AddInt32(&best.Connections, 1)
	return best
}
