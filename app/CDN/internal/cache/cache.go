// Package cache fournit des implémentations de cache pour le CDN
// Il propose deux types de cache : en mémoire (LRU) et Redis
package cache

import (
	"context"
	"github.com/hashicorp/golang-lru"
	"github.com/redis/go-redis/v9"
	"time"
)

// Cache définit l'interface commune pour toutes les implémentations de cache
type Cache interface {
	// Get récupère une valeur du cache à partir de sa clé
	// Retourne la valeur et un booléen indiquant si la clé existe
	Get(key string) (interface{}, bool)
	
	// Set stocke une valeur dans le cache avec la clé spécifiée
	// Retourne une erreur si l'opération échoue
	Set(key string, value interface{}) error
	
	// Delete supprime une valeur du cache à partir de sa clé
	// Retourne une erreur si l'opération échoue
	Delete(key string) error
}

// MemoryCache implémente un cache en mémoire utilisant l'algorithme LRU
type MemoryCache struct {
	lru *lru.Cache // Cache LRU sous-jacent
}

// NewMemoryCache crée une nouvelle instance de MemoryCache avec une taille maximale spécifiée
// Retourne une erreur si la création du cache LRU échoue
func NewMemoryCache(size int) (*MemoryCache, error) {
	l, err := lru.New(size)
	if err != nil {
		return nil, err
	}
	return &MemoryCache{lru: l}, nil
}

// Get récupère une valeur du cache mémoire
func (m *MemoryCache) Get(key string) (interface{}, bool) {
	return m.lru.Get(key)
}

// Set ajoute ou met à jour une valeur dans le cache mémoire
func (m *MemoryCache) Set(key string, value interface{}) error {
	m.lru.Add(key, value)
	return nil
}

// Delete supprime une valeur du cache mémoire
func (m *MemoryCache) Delete(key string) error {
	m.lru.Remove(key)
	return nil
}

// RedisCache implémente un cache distribué utilisant Redis
type RedisCache struct {
	client *redis.Client // Client Redis
}

// NewRedisCache crée une nouvelle instance de RedisCache
// url: l'adresse du serveur Redis
// db: l'index de la base de données Redis à utiliser
func NewRedisCache(url string, db int) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr: url,
		DB:   db,
	})
	return &RedisCache{client: client}
}

// Get récupère une valeur du cache Redis
// Retourne nil, false si la clé n'existe pas ou en cas d'erreur
func (r *RedisCache) Get(key string) (interface{}, bool) {
	val, err := r.client.Get(context.Background(), key).Result()
	if err != nil {
		return nil, false
	}
	return val, true
}

// Set stocke une valeur dans Redis avec une expiration de 24 heures
func (r *RedisCache) Set(key string, value interface{}) error {
	return r.client.Set(context.Background(), key, value, 24*time.Hour).Err()
}

// Delete supprime une valeur du cache Redis
func (r *RedisCache) Delete(key string) error {
	return r.client.Del(context.Background(), key).Err()
}
