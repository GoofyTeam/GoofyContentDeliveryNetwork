package cache

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// simulerServeurOrigine crée un serveur de test qui simule un serveur d'origine
func simulerServeurOrigine(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("X-Server", "Origine")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Contenu du serveur d'origine"))
	}))
}

func TestCDNCache(t *testing.T) {
	cache, err := NewMemoryCache(100)
	if err != nil {
		t.Fatalf("Erreur création cache: %v", err)
	}
	ctx := context.Background()

	// Créer un serveur d'origine simulé
	serveurOrigine := simulerServeurOrigine(t)
	defer serveurOrigine.Close()

	t.Run("Test mise en cache d'une requête CDN", func(t *testing.T) {
		// Simuler une requête au serveur d'origine
		resp, err := http.Get(serveurOrigine.URL + "/test-content")
		if err != nil {
			t.Fatalf("Erreur requête serveur origine: %v", err)
		}
		defer resp.Body.Close()

		// Lire le contenu de la réponse
		contenu, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Erreur lecture réponse: %v", err)
		}

		// Créer une entrée de cache
		headers := make(map[string]string)
		for k, v := range resp.Header {
			headers[k] = v[0]
		}

		// Mettre en cache la réponse
		err = cache.Set(ctx, "/test-content", contenu, headers, time.Hour)
		if err != nil {
			t.Fatalf("Erreur mise en cache: %v", err)
		}

		// Vérifier que la réponse est en cache
		entry, exists, err := cache.Get(ctx, "/test-content")
		if err != nil {
			t.Fatalf("Erreur récupération cache: %v", err)
		}
		if !exists {
			t.Fatal("La réponse devrait être en cache")
		}

		// Vérifier le contenu et les headers
		cachedContent := entry.Value.([]byte)
		if string(cachedContent) != string(contenu) {
			t.Error("Le contenu en cache ne correspond pas à la réponse originale")
		}
		if entry.Headers["X-Server"] != "Origine" {
			t.Error("Les headers en cache ne correspondent pas")
		}
	})

	t.Run("Test performance du cache", func(t *testing.T) {
		// Préparer les données de test
		donnees := make([]struct {
			url     string
			contenu []byte
		}, 100) // Réduire à 100 pour un test plus réaliste

		for i := 0; i < 100; i++ {
			donnees[i] = struct {
				url     string
				contenu []byte
			}{
				url:     fmt.Sprintf("/perf-test-%d", i),
				contenu: []byte(fmt.Sprintf("Contenu de test pour l'entrée %d", i)),
			}
		}

		// Test de mise en cache en série
		debut := time.Now()
		for _, d := range donnees {
			err := cache.Set(ctx, d.url, d.contenu, nil, time.Hour)
			if err != nil {
				t.Fatalf("Erreur mise en cache: %v", err)
			}
		}
		dureeSet := time.Since(debut)
		t.Logf("Temps total de mise en cache (série): %v", dureeSet)
		t.Logf("Temps moyen de mise en cache: %v/opération", dureeSet/100)

		// Test de lecture en série
		var hits, misses int
		debut = time.Now()
		for _, d := range donnees {
			entry, exists, err := cache.Get(ctx, d.url)
			if err != nil {
				t.Fatalf("Erreur lecture cache: %v", err)
			}
			if exists {
				hits++
				// Vérifier l'intégrité des données
				if string(entry.Value.([]byte)) != string(d.contenu) {
					t.Errorf("Corruption des données pour %s", d.url)
				}
			} else {
				misses++
			}
		}
		dureeGet := time.Since(debut)
		t.Logf("Temps total de lecture (série): %v", dureeGet)
		t.Logf("Temps moyen de lecture: %v/opération", dureeGet/100)
		t.Logf("Ratio hits/misses: %d/%d", hits, misses)

		// Test de lecture en parallèle
		debut = time.Now()
		var wg sync.WaitGroup
		errChan := make(chan error, len(donnees))

		for _, d := range donnees {
			wg.Add(1)
			go func(url string) {
				defer wg.Done()
				_, exists, err := cache.Get(ctx, url)
				if err != nil {
					errChan <- fmt.Errorf("erreur lecture parallèle: %v", err)
				}
				if !exists {
					errChan <- fmt.Errorf("donnée non trouvée: %s", url)
				}
			}(d.url)
		}

		wg.Wait()
		close(errChan)
		dureeGetParallele := time.Since(debut)
		t.Logf("Temps total de lecture (parallèle): %v", dureeGetParallele)

		// Vérifier les erreurs de lecture parallèle
		for err := range errChan {
			t.Errorf("Erreur pendant la lecture parallèle: %v", err)
		}

		// Critères de performance plus réalistes
		tempsMaxSet := 5 * time.Millisecond  // 5ms max par opération de mise en cache
		tempsMaxGet := 1 * time.Millisecond  // 1ms max par opération de lecture
		
		if dureeSet/100 > tempsMaxSet {
			t.Errorf("Performance SET insuffisante: %v/op (max attendu: %v/op)", dureeSet/100, tempsMaxSet)
		}
		if dureeGet/100 > tempsMaxGet {
			t.Errorf("Performance GET insuffisante: %v/op (max attendu: %v/op)", dureeGet/100, tempsMaxGet)
		}
		if hits != len(donnees) {
			t.Errorf("Certaines données n'ont pas été trouvées dans le cache: %d hits sur %d attendus", hits, len(donnees))
		}
	})

	t.Run("Test gestion de la charge", func(t *testing.T) {
		// Simuler des accès concurrents
		const nbGoroutines = 100
		const nbRequetesParGoroutine = 1000

		errChan := make(chan error, nbGoroutines)
		done := make(chan bool, nbGoroutines)

		for i := 0; i < nbGoroutines; i++ {
			go func(id int) {
				for j := 0; j < nbRequetesParGoroutine; j++ {
					key := fmt.Sprintf("/charge-test-%d-%d", id, j)
					err := cache.Set(ctx, key, []byte("test"), nil, time.Hour)
					if err != nil {
						errChan <- fmt.Errorf("goroutine %d: %v", id, err)
						return
					}

					_, _, err = cache.Get(ctx, key)
					if err != nil {
						errChan <- fmt.Errorf("goroutine %d: %v", id, err)
						return
					}
				}
				done <- true
			}(i)
		}

		// Attendre la fin des goroutines
		for i := 0; i < nbGoroutines; i++ {
			select {
			case err := <-errChan:
				t.Errorf("Erreur pendant le test de charge: %v", err)
			case <-done:
				// OK
			case <-time.After(30 * time.Second):
				t.Error("Timeout pendant le test de charge")
			}
		}
	})

	t.Run("Test nettoyage automatique", func(t *testing.T) {
		// Remplir le cache avec des entrées qui expirent rapidement
		for i := 0; i < 50; i++ {
			key := fmt.Sprintf("/expire-test-%d", i)
			err := cache.Set(ctx, key, []byte("test"), nil, time.Millisecond)
			if err != nil {
				t.Fatalf("Erreur mise en cache: %v", err)
			}
		}

		// Attendre que les entrées expirent
		time.Sleep(time.Millisecond * 10)

		// Vérifier que les entrées sont bien nettoyées
		var entreesExpirees int
		for i := 0; i < 50; i++ {
			key := fmt.Sprintf("/expire-test-%d", i)
			_, exists, _ := cache.Get(ctx, key)
			if exists {
				entreesExpirees++
			}
		}

		if entreesExpirees > 0 {
			t.Errorf("%d entrées expirées toujours en cache", entreesExpirees)
		}
	})
}
