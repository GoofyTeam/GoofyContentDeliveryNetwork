package loadbalancer

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func setupTestServers(t *testing.T) ([]*httptest.Server, []string) {
	servers := make([]*httptest.Server, 3)
	urls := make([]string, 3)

	for i := 0; i < 3; i++ {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		servers[i] = server
		urls[i] = server.URL
	}

	return servers, urls
}

func TestRoundRobin(t *testing.T) {
	servers, urls := setupTestServers(t)
	defer func() {
		for _, server := range servers {
			server.Close()
		}
	}()

	config := Config{
		HealthCheckInterval: time.Second,
		HealthCheckTimeout:  time.Second,
		MaxFailCount:       3,
		RetryTimeout:       time.Second,
	}

	lb := NewRoundRobin(urls, config)
	ctx := context.Background()

	// Test distribution équitable
	counts := make(map[string]int)
	for i := 0; i < 300; i++ {
		backend, err := lb.NextBackend(ctx)
		if err != nil {
			t.Fatalf("Erreur inattendue: %v", err)
		}
		counts[backend.URL]++
	}

	// Vérification de la distribution
	for _, count := range counts {
		if count < 95 || count > 105 {
			t.Errorf("Distribution non équitable: %v", counts)
		}
	}

	// Test des métriques
	metrics := lb.GetMetrics()
	if metrics.TotalRequests != 300 {
		t.Errorf("Nombre total de requêtes incorrect: %d", metrics.TotalRequests)
	}
}

func TestWeightedRoundRobin(t *testing.T) {
	servers, urls := setupTestServers(t)
	defer func() {
		for _, server := range servers {
			server.Close()
		}
	}()

	weights := []int{1, 2, 3} // Le dernier serveur devrait recevoir 3x plus de requêtes
	config := Config{
		HealthCheckInterval: time.Second,
		HealthCheckTimeout:  time.Second,
		MaxFailCount:       3,
		RetryTimeout:       time.Second,
	}

	lb := NewWeightedRoundRobin(urls, weights, config)
	ctx := context.Background()

	counts := make(map[string]int)
	for i := 0; i < 600; i++ {
		backend, err := lb.NextBackend(ctx)
		if err != nil {
			t.Fatalf("Erreur inattendue: %v", err)
		}
		counts[backend.URL]++
	}

	// Vérification des ratios
	total := float64(counts[urls[0]] + counts[urls[1]] + counts[urls[2]])
	ratios := make(map[string]float64)
	for url, count := range counts {
		ratios[url] = float64(count) / total
	}

	expectedRatios := map[int]float64{
		0: 1.0 / 6.0,  // Weight 1
		1: 2.0 / 6.0,  // Weight 2
		2: 3.0 / 6.0,  // Weight 3
	}

	for i, url := range urls {
		expected := expectedRatios[i]
		actual := ratios[url]
		if actual < expected-0.05 || actual > expected+0.05 {
			t.Errorf("Ratio incorrect pour %s: attendu %.2f, obtenu %.2f", url, expected, actual)
		}
	}
}

func TestLeastConnections(t *testing.T) {
	servers, urls := setupTestServers(t)
	defer func() {
		for _, server := range servers {
			server.Close()
		}
	}()

	config := Config{
		HealthCheckInterval: time.Second,
		HealthCheckTimeout:  time.Second,
		MaxFailCount:       3,
		RetryTimeout:       time.Second,
	}

	lb := NewLeastConnections(urls, config)
	ctx := context.Background()

	// Simulation de connexions actives
	backend1, _ := lb.NextBackend(ctx)
	backend1.Connections = 5
	backend2, _ := lb.NextBackend(ctx)
	backend2.Connections = 2
	backend3, _ := lb.NextBackend(ctx)
	backend3.Connections = 8

	// Le backend avec le moins de connexions devrait être choisi
	chosen, err := lb.NextBackend(ctx)
	if err != nil {
		t.Fatalf("Erreur inattendue: %v", err)
	}

	if chosen.URL != backend2.URL {
		t.Errorf("Le mauvais backend a été choisi: attendu %s, obtenu %s", backend2.URL, chosen.URL)
	}
}

func TestHealthCheck(t *testing.T) {
	servers, urls := setupTestServers(t)
	defer func() {
		for _, server := range servers {
			server.Close()
		}
	}()

	config := Config{
		HealthCheckInterval: 100 * time.Millisecond,
		HealthCheckTimeout:  time.Second,
		MaxFailCount:       2,
		RetryTimeout:       time.Second,
	}

	lb := NewRoundRobin(urls, config)
	ctx := context.Background()

	// Arrêt d'un serveur
	servers[1].Close()

	// Attente que le health check détecte le serveur mort
	time.Sleep(300 * time.Millisecond)

	// Vérification que le serveur mort n'est pas sélectionné
	for i := 0; i < 100; i++ {
		backend, err := lb.NextBackend(ctx)
		if err != nil {
			t.Fatalf("Erreur inattendue: %v", err)
		}
		if backend.URL == urls[1] {
			t.Error("Un serveur mort a été sélectionné")
		}
	}

	// Vérification des métriques
	metrics := lb.GetMetrics()
	//print metrics
	fmt.Println(metrics)
	if metrics.ActiveBackends != 2 {
		t.Errorf("Nombre incorrect de backends actifs: %d", metrics.ActiveBackends)
	}
}
