// Package cache fournit des implémentations de cache pour le CDN
// Il propose deux types de cache : en mémoire (LRU) et Redis
package cache

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/redis/go-redis/v9"
)

// CacheMetrics contient les métriques de performance du cache
type CacheMetrics struct {
	Hits   uint64
	Misses uint64
	Items  uint64
}




// CacheEntry représente une entrée dans le cache avec TTL
type CacheEntry struct {
	Value      interface{}
	Expiration time.Time
	Headers    map[string]string
}

// Cache définit l'interface commune pour toutes les implémentations de cache
type Cache interface {
	// Get récupère une valeur du cache à partir de sa clé
	Get(ctx context.Context, key string) (*CacheEntry, bool, error)
	
	// Set stocke une valeur dans le cache avec la clé spécifiée
	Set(ctx context.Context, key string, value interface{}, headers map[string]string, ttl time.Duration) error
	
	// Delete supprime une valeur du cache
	Delete(ctx context.Context, key string) error

	// GetMetrics retourne les métriques du cache
	GetMetrics() *CacheMetrics

	// Clear vide complètement le cache
	Clear()
}

// MemoryCache implémente un cache en mémoire utilisant l'algorithme LRU
type MemoryCache struct {
	lru     *lru.Cache
	metrics CacheMetrics
	maxSize int
}

// NewMemoryCache crée une nouvelle instance de MemoryCache
func NewMemoryCache(size int) (*MemoryCache, error) {
	l, err := lru.New(size)
	if err != nil {
		return nil, err
	}
	return &MemoryCache{
		lru:     l,
		maxSize: size,
	}, nil
}

// Get récupère une valeur du cache mémoire
func (m *MemoryCache) Get(ctx context.Context, key string) (*CacheEntry, bool, error) {
	value, exists := m.lru.Get(key)
	if !exists {
		atomic.AddUint64(&m.metrics.Misses, 1)
		return nil, false, nil
	}

	entry := value.(*CacheEntry)
	if time.Now().After(entry.Expiration) {
		m.lru.Remove(key)
		atomic.AddUint64(&m.metrics.Misses, 1)
		return nil, false, nil
	}

	atomic.AddUint64(&m.metrics.Hits, 1)
	return entry, true, nil
}

// Set ajoute ou met à jour une valeur dans le cache mémoire
func (m *MemoryCache) Set(ctx context.Context, key string, value interface{}, headers map[string]string, ttl time.Duration) error {
	entry := &CacheEntry{
		Value:      value,
		Headers:    headers,
		Expiration: time.Now().Add(ttl),
	}

	// Si la clé existe déjà, ne pas incrémenter le compteur
	if _, exists := m.lru.Get(key); !exists {
		// Si le cache est plein, le LRU va automatiquement évincer un élément
		if m.lru.Len() >= m.maxSize {
			atomic.AddUint64(&m.metrics.Items, ^uint64(0)) // Décrémente le compteur pour l'élément qui sera évincé
		}
		atomic.AddUint64(&m.metrics.Items, 1)
	}

	m.lru.Add(key, entry)
	return nil
}

// Delete supprime une valeur du cache mémoire
func (m *MemoryCache) Delete(ctx context.Context, key string) error {
	if m.lru.Remove(key) {
		atomic.AddUint64(&m.metrics.Items, ^uint64(0))
	}
	return nil
}

// GetMetrics retourne les métriques du cache mémoire
func (m *MemoryCache) GetMetrics() *CacheMetrics {
	return &CacheMetrics{
		Hits:   atomic.LoadUint64(&m.metrics.Hits),
		Misses: atomic.LoadUint64(&m.metrics.Misses),
		Items:  atomic.LoadUint64(&m.metrics.Items),
	}
}

// Clear vide complètement le cache mémoire
func (m *MemoryCache) Clear() {
	m.lru.Purge()
	atomic.StoreUint64(&m.metrics.Items, 0)
}

// RedisCache implémente un cache distribué utilisant Redis
type RedisCache struct {
	client  *redis.Client
	metrics CacheMetrics
}

// crée une nouvelle instance de RedisCache
func NewRedisCache(url string, db int) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr: url,
		DB:   db,
	})

	// Test de connexion
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisCache{client: client}, nil
}

// Get récupère une valeur du cache Redis
func (r *RedisCache) Get(ctx context.Context, key string) (*CacheEntry, bool, error) {
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		atomic.AddUint64(&r.metrics.Misses, 1)
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false, err
	}

	if time.Now().After(entry.Expiration) {
		r.Delete(ctx, key)
		atomic.AddUint64(&r.metrics.Misses, 1)
		return nil, false, nil
	}

	atomic.AddUint64(&r.metrics.Hits, 1)
	return &entry, true, nil
}

// Set stocke une valeur dans Redis
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, headers map[string]string, ttl time.Duration) error {
	entry := &CacheEntry{
		Value:      value,
		Headers:    headers,
		Expiration: time.Now().Add(ttl),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return err
	}

	atomic.AddUint64(&r.metrics.Items, 1)
	return nil
}

// Delete supprime une valeur du cache Redis
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return err
	}
	atomic.AddUint64(&r.metrics.Items, ^uint64(0))
	return nil
}

// GetMetrics retourne les métriques du cache Redis
func (r *RedisCache) GetMetrics() *CacheMetrics {
	return &CacheMetrics{
		Hits:   atomic.LoadUint64(&r.metrics.Hits),
		Misses: atomic.LoadUint64(&r.metrics.Misses),
		Items:  atomic.LoadUint64(&r.metrics.Items),
	}
}

// Clear vide complètement le cache Redis
func (r *RedisCache) Clear() {
	r.client.FlushDB(context.Background())
	atomic.StoreUint64(&r.metrics.Items, 0)
}
