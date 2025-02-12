package cache

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestMemoryCache(t *testing.T) {
	t.Run("Test création du cache", func(t *testing.T) {
		cache, err := NewMemoryCache(100)
		if err != nil {
			t.Fatalf("Erreur lors de la création du cache: %v", err)
		}
		if cache == nil {
			t.Fatal("Le cache ne devrait pas être nil")
		}
	})

	t.Run("Test Set et Get basique", func(t *testing.T) {
		cache, _ := NewMemoryCache(100)
		ctx := context.Background()
		key := "test-key"
		value := "test-value"
		headers := map[string]string{"Content-Type": "text/plain"}

		err := cache.Set(ctx, key, value, headers, time.Minute)
		if err != nil {
			t.Fatalf("Erreur lors du Set: %v", err)
		}

		entry, exists, err := cache.Get(ctx, key)
		if err != nil {
			t.Fatalf("Erreur lors du Get: %v", err)
		}
		if !exists {
			t.Fatal("La valeur devrait exister dans le cache")
		}
		if entry.Value != value {
			t.Errorf("Valeur attendue %v, obtenue %v", value, entry.Value)
		}
		if entry.Headers["Content-Type"] != "text/plain" {
			t.Errorf("Header Content-Type attendu 'text/plain', obtenu '%v'", entry.Headers["Content-Type"])
		}
	})

	t.Run("Test expiration", func(t *testing.T) {
		cache, _ := NewMemoryCache(100)
		ctx := context.Background()
		key := "test-expiration"
		value := "test-value"

		err := cache.Set(ctx, key, value, nil, time.Millisecond)
		if err != nil {
			t.Fatalf("Erreur lors du Set: %v", err)
		}

		time.Sleep(time.Millisecond * 2)

		_, exists, err := cache.Get(ctx, key)
		if err != nil {
			t.Fatalf("Erreur lors du Get: %v", err)
		}
		if exists {
			t.Error("La valeur devrait être expirée")
		}
	})

	t.Run("Test Delete", func(t *testing.T) {
		cache, _ := NewMemoryCache(100)
		ctx := context.Background()
		key := "test-delete"
		value := "test-value"

		cache.Set(ctx, key, value, nil, time.Minute)
		err := cache.Delete(ctx, key)
		if err != nil {
			t.Fatalf("Erreur lors du Delete: %v", err)
		}

		_, exists, err := cache.Get(ctx, key)
		if err != nil {
			t.Fatalf("Erreur lors du Get après Delete: %v", err)
		}
		if exists {
			t.Error("La valeur devrait être supprimée")
		}
	})

	t.Run("Test métriques", func(t *testing.T) {
		cache, _ := NewMemoryCache(100)
		ctx := context.Background()
		key := "test-metrics"
		value := "test-value"

		// Test des misses
		cache.Get(ctx, key)
		metrics := cache.GetMetrics()
		if metrics.Misses != 1 {
			t.Errorf("Attendu 1 miss, obtenu %d", metrics.Misses)
		}

		// Test des hits
		cache.Set(ctx, key, value, nil, time.Minute)
		cache.Get(ctx, key)
		metrics = cache.GetMetrics()
		if metrics.Hits != 1 {
			t.Errorf("Attendu 1 hit, obtenu %d", metrics.Hits)
		}

		// Test du nombre d'items
		if metrics.Items != 1 {
			t.Errorf("Attendu 1 item, obtenu %d", metrics.Items)
		}
	})

	t.Run("Test limite de taille", func(t *testing.T) {
		size := 2
		cache, _ := NewMemoryCache(size)
		ctx := context.Background()

		// Ajouter plus d'éléments que la taille maximale
		for i := 0; i < size+1; i++ {
			key := fmt.Sprintf("key-%d", i)
			cache.Set(ctx, key, i, nil, time.Minute)
		}

		// Vérifier que le cache respecte sa taille maximale
		metrics := cache.GetMetrics()
		if metrics.Items > uint64(size) {
			t.Errorf("Le cache contient %d items, devrait en contenir maximum %d", metrics.Items, size)
		}
	})
}

func TestRedisCache(t *testing.T) {
	// Skip si Redis n'est pas disponible en local
	redisURL := "redis://localhost:6379/0"
	cache, err := NewRedisCache(redisURL, 0)
	if err != nil {
		t.Skip("Redis n'est pas disponible, tests ignorés")
	}

	t.Run("Test Set et Get basique", func(t *testing.T) {
		ctx := context.Background()
		key := "test-redis-key"
		value := "test-redis-value"
		headers := map[string]string{"Content-Type": "text/plain"}

		err := cache.Set(ctx, key, value, headers, time.Minute)
		if err != nil {
			t.Fatalf("Erreur lors du Set Redis: %v", err)
		}

		entry, exists, err := cache.Get(ctx, key)
		if err != nil {
			t.Fatalf("Erreur lors du Get Redis: %v", err)
		}
		if !exists {
			t.Fatal("La valeur devrait exister dans Redis")
		}
		if entry.Value != value {
			t.Errorf("Valeur Redis attendue %v, obtenue %v", value, entry.Value)
		}
		if entry.Headers["Content-Type"] != "text/plain" {
			t.Errorf("Header Redis Content-Type attendu 'text/plain', obtenu '%v'", entry.Headers["Content-Type"])
		}

		// Nettoyage
		cache.Delete(ctx, key)
	})

	// Les autres tests Redis suivent le même modèle que MemoryCache...
}

type MockHTTPResponse struct {
	Body        []byte
	StatusCode  int
	Headers     map[string]string
	RequestTime time.Time
}

func TestCacheHTTPResponses(t *testing.T) {
	cache, err := NewMemoryCache(100)
	if err != nil {
		t.Fatalf("Erreur création cache: %v", err)
	}
	ctx := context.Background()

	t.Run("Test mise en cache réponse HTTP", func(t *testing.T) {
		// Simuler une réponse HTTP
		originalResponse := MockHTTPResponse{
			Body:       []byte("Contenu de test"),
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"Content-Type":  "text/plain",
				"Cache-Control": "max-age=3600",
			},
			RequestTime: time.Now(),
		}

		// Mettre en cache
		err := cache.Set(ctx, "/test-url", originalResponse, originalResponse.Headers, time.Hour)
		if err != nil {
			t.Fatalf("Erreur mise en cache: %v", err)
		}

		// Récupérer du cache
		entry, exists, err := cache.Get(ctx, "/test-url")
		if err != nil {
			t.Fatalf("Erreur récupération cache: %v", err)
		}
		if !exists {
			t.Fatal("La réponse devrait être en cache")
		}

		// Vérifier le contenu
		cachedResponse := entry.Value.(MockHTTPResponse)
		if !bytes.Equal(cachedResponse.Body, originalResponse.Body) {
			t.Error("Le contenu en cache ne correspond pas à l'original")
		}
		if cachedResponse.StatusCode != originalResponse.StatusCode {
			t.Error("Le status code en cache ne correspond pas à l'original")
		}
		if entry.Headers["Content-Type"] != originalResponse.Headers["Content-Type"] {
			t.Error("Les headers en cache ne correspondent pas à l'original")
		}
	})

	t.Run("Test expiration réponse HTTP", func(t *testing.T) {
		shortTTL := 100 * time.Millisecond
		response := MockHTTPResponse{
			Body:       []byte("Contenu qui expire vite"),
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"Content-Type": "text/plain",
			},
		}

		// Mettre en cache avec TTL court
		err := cache.Set(ctx, "/expire-test", response, response.Headers, shortTTL)
		if err != nil {
			t.Fatalf("Erreur mise en cache: %v", err)
		}

		// Vérifier immédiatement
		_, exists, _ := cache.Get(ctx, "/expire-test")
		if !exists {
			t.Fatal("La réponse devrait être en cache immédiatement")
		}

		// Attendre l'expiration
		time.Sleep(shortTTL + 50*time.Millisecond)

		// Vérifier après expiration
		_, exists, _ = cache.Get(ctx, "/expire-test")
		if exists {
			t.Error("La réponse devrait être expirée et non disponible")
		}
	})

	t.Run("Test mise à jour réponse HTTP", func(t *testing.T) {
		key := "/update-test"
		original := MockHTTPResponse{
			Body:       []byte("Contenu original"),
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"Content-Type": "text/plain",
			},
		}

		// Première mise en cache
		err := cache.Set(ctx, key, original, original.Headers, time.Hour)
		if err != nil {
			t.Fatalf("Erreur première mise en cache: %v", err)
		}

		// Mise à jour avec nouveau contenu
		updated := MockHTTPResponse{
			Body:       []byte("Contenu mis à jour"),
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"Content-Type": "text/plain",
				"Updated":      "true",
			},
		}

		err = cache.Set(ctx, key, updated, updated.Headers, time.Hour)
		if err != nil {
			t.Fatalf("Erreur mise à jour cache: %v", err)
		}

		// Vérifier le contenu mis à jour
		entry, exists, _ := cache.Get(ctx, key)
		if !exists {
			t.Fatal("La réponse mise à jour devrait être en cache")
		}

		cachedResponse := entry.Value.(MockHTTPResponse)
		if !bytes.Equal(cachedResponse.Body, updated.Body) {
			t.Error("Le contenu en cache ne correspond pas à la mise à jour")
		}
		if entry.Headers["Updated"] != "true" {
			t.Error("Les headers mis à jour ne sont pas présents")
		}
	})

	t.Run("Test suppression réponse HTTP", func(t *testing.T) {
		key := "/delete-test"
		response := MockHTTPResponse{
			Body:       []byte("Contenu à supprimer"),
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"Content-Type": "text/plain",
			},
		}

		// Mettre en cache
		err := cache.Set(ctx, key, response, response.Headers, time.Hour)
		if err != nil {
			t.Fatalf("Erreur mise en cache: %v", err)
		}

		// Supprimer
		err = cache.Delete(ctx, key)
		if err != nil {
			t.Fatalf("Erreur suppression cache: %v", err)
		}

		// Vérifier la suppression
		_, exists, _ := cache.Get(ctx, key)
		if exists {
			t.Error("La réponse devrait être supprimée du cache")
		}
	})
}

func BenchmarkHTTPCache(b *testing.B) {
	cache, _ := NewMemoryCache(1000)
	ctx := context.Background()
	response := MockHTTPResponse{
		Body:       []byte("Contenu de benchmark"),
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "text/plain",
		},
	}

	b.Run("Mise en cache", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("/bench-key-%d", i)
			cache.Set(ctx, key, response, response.Headers, time.Hour)
		}
	})

	b.Run("Lecture cache", func(b *testing.B) {
		key := "/bench-read"
		cache.Set(ctx, key, response, response.Headers, time.Hour)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cache.Get(ctx, key)
		}
	})
}

func BenchmarkCache(b *testing.B) {
	cache, _ := NewMemoryCache(1000)
	ctx := context.Background()
	key := "bench-key"
	value := "bench-value"

	b.Run("Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cache.Set(ctx, fmt.Sprintf("%s-%d", key, i), value, nil, time.Minute)
		}
	})

	b.Run("Get", func(b *testing.B) {
		cache.Set(ctx, key, value, nil, time.Minute)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cache.Get(ctx, key)
		}
	})
}
