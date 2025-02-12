package api

import (
	"app/internal/metrics"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// generateLoad génère une charge CPU intensive
func generateLoad(duration time.Duration, intensity int) {
	start := time.Now()
	for time.Since(start) < duration {
		for i := 0; i < intensity; i++ {
			sha := sha256.New()
			data := make([]byte, 1024)
			rand.Read(data)
			sha.Sum(data)
		}
	}
}

// SetupTestRoutes configure les routes de test pour générer des métriques
func SetupTestRoutes(r *gin.Engine) {
	test := r.Group("/test")
	{
		// Test de cache avec contenu statique
		test.GET("/cache/static/:id", func(c *gin.Context) {
			id := c.Param("id")
			// Contenu statique qui sera mis en cache
			content := fmt.Sprintf("Contenu statique pour l'ID %s - Timestamp de création: %d", 
				id, time.Now().Unix())
			
			c.Header("Cache-Control", "public, max-age=60")
			c.String(http.StatusOK, content)
		})

		// Test de cache avec différentes tailles
		test.GET("/cache/size/:size", func(c *gin.Context) {
			size := c.Param("size")
			var content string
			switch size {
			case "small":
				content = "Petit contenu cacheable"
			case "medium":
				content = string(make([]byte, 1024*10)) // 10KB
			case "large":
				content = string(make([]byte, 1024*100)) // 100KB
			}
			
			c.Header("Cache-Control", "public, max-age=60")
			c.String(http.StatusOK, content)
		})

		// Test de cache intensif
		var hammerLimiter = time.NewTicker(5 * time.Second)
		test.GET("/cache/hammer", func(c *gin.Context) {
			select {
			case <-hammerLimiter.C:
				var wg sync.WaitGroup
				iterations := 100 // Réduit de 1000 à 100 pour limiter la charge

				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func(i int) {
						defer wg.Done()
						if rand.Float64() < 0.7 { // 70% de hits
							metrics.CacheHits.Inc()
							time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
						} else {
							metrics.CacheMisses.Inc()
							time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
						}
					}(i)
				}

				wg.Wait()
				c.JSON(http.StatusOK, gin.H{"status": "cache hammer complete", "iterations": iterations})
			default:
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": "Veuillez attendre 5 secondes entre chaque test de charge",
				})
				return
			}
		})

		// Test de latence variable
		test.GET("/latency/random", func(c *gin.Context) {
			// Distribution exponentielle pour simuler des pics de latence
			latency := time.Duration(rand.ExpFloat64() * 1000) * time.Millisecond
			if latency > 5*time.Second {
				latency = 5 * time.Second // Cap à 5 secondes
			}
			time.Sleep(latency)
			c.JSON(http.StatusOK, gin.H{"latency": latency.String()})
		})

		// Test de charge CPU intensive
		test.GET("/cpu/stress/:cores", func(c *gin.Context) {
			var cores int
			fmt.Sscanf(c.Param("cores"), "%d", &cores)
			if cores <= 0 || cores > runtime.NumCPU() {
				cores = runtime.NumCPU()
			}

			duration := 30 * time.Second // 30 secondes de charge
			intensity := 100000         // Intensité de la charge

			for i := 0; i < cores; i++ {
				go generateLoad(duration, intensity)
			}

			metrics.CpuUsage.Set(float64(cores) * 100 / float64(runtime.NumCPU()))
			c.JSON(http.StatusOK, gin.H{
				"status":    "CPU stress started",
				"cores":     cores,
				"duration":  duration.String(),
				"intensity": intensity,
			})
		})

		// Test de mémoire intensive
		test.GET("/memory/stress/:gb", func(c *gin.Context) {
			var gb int
			fmt.Sscanf(c.Param("gb"), "%d", &gb)
			if gb <= 0 {
				gb = 1
			}

			// Allouer de la mémoire en plusieurs chunks pour éviter OOM killer
			chunks := make([][]byte, gb)
			for i := 0; i < gb; i++ {
				chunks[i] = make([]byte, 1024*1024*1024) // 1GB par chunk
				rand.Read(chunks[i])
				metrics.MemoryUsage.Add(1024 * 1024 * 1024)
			}

			// Garder la mémoire allouée pendant 10 secondes
			time.Sleep(10 * time.Second)

			c.JSON(http.StatusOK, gin.H{"allocated": fmt.Sprintf("%dGB", gb)})
		})

		// Test de charge mixte
		test.GET("/mixed/chaos", func(c *gin.Context) {
			var wg sync.WaitGroup
			duration := 30 * time.Second
			start := time.Now()

			// Goroutine pour la charge CPU
			wg.Add(1)
			go func() {
				defer wg.Done()
				for time.Since(start) < duration {
					generateLoad(100*time.Millisecond, 10000)
					time.Sleep(50 * time.Millisecond)
				}
			}()

			// Goroutine pour la mémoire
			wg.Add(1)
			go func() {
				defer wg.Done()
				chunks := make([][]byte, 0)
				for time.Since(start) < duration {
					chunk := make([]byte, 100*1024*1024) // 100MB
					rand.Read(chunk)
					chunks = append(chunks, chunk)
					metrics.MemoryUsage.Add(100 * 1024 * 1024)
					time.Sleep(1 * time.Second)
				}
			}()

			// Goroutine pour les erreurs
			wg.Add(1)
			go func() {
				defer wg.Done()
				errors := []int{400, 401, 403, 404, 500, 502, 503}
				for time.Since(start) < duration {
					code := errors[rand.Intn(len(errors))]
					metrics.HttpRequestsTotal.WithLabelValues("GET", "/error", fmt.Sprintf("%d", code)).Inc()
					time.Sleep(100 * time.Millisecond)
				}
			}()

			// Goroutine pour les attaques DDoS simulées
			wg.Add(1)
			go func() {
				defer wg.Done()
				for time.Since(start) < duration {
					if rand.Float64() < 0.3 { // 30% de chance d'attaque
						metrics.DDoSAttempts.Inc()
					}
					time.Sleep(50 * time.Millisecond)
				}
			}()

			wg.Wait()
			c.JSON(http.StatusOK, gin.H{
				"status":   "chaos complete",
				"duration": duration.String(),
			})
		})

		// Test de cache
		test.GET("/cache/hit", func(c *gin.Context) {
			metrics.CacheHits.Inc()
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			c.JSON(http.StatusOK, gin.H{"status": "cache hit"})
		})

		test.GET("/cache/miss", func(c *gin.Context) {
			metrics.CacheMisses.Inc()
			time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
			c.JSON(http.StatusOK, gin.H{"status": "cache miss"})
		})

		// Test de latence
		test.GET("/latency/:ms", func(c *gin.Context) {
			ms := c.Param("ms")
			duration, err := time.ParseDuration(ms + "ms")
			if err != nil {
				duration = 100 * time.Millisecond
			}
			time.Sleep(duration)
			c.JSON(http.StatusOK, gin.H{"latency": ms + "ms"})
		})

		// Test d'erreurs
		test.GET("/error/:code", func(c *gin.Context) {
			code := c.Param("code")
			switch code {
			case "404":
				c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
			case "500":
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			default:
				c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
			}
		})

		// Test de charge CPU
		test.GET("/cpu/:seconds", func(c *gin.Context) {
			seconds, err := time.ParseDuration(c.Param("seconds") + "s")
			if err != nil {
				seconds = 5 * time.Second
			}
			
			// Simuler une charge CPU
			go func() {
				start := time.Now()
				for time.Since(start) < seconds {
					for i := 0; i < 1000000; i++ {
						_ = rand.Float64() * rand.Float64()
					}
				}
			}()

			c.JSON(http.StatusOK, gin.H{"status": "CPU load started", "duration": seconds.String()})
		})

		// Test de mémoire
		test.GET("/memory/:mb", func(c *gin.Context) {
			var mb int
			_, err := fmt.Sscanf(c.Param("mb"), "%d", &mb)
			if err != nil {
				mb = 100
			}

			// Allouer de la mémoire (temporairement)
			data := make([]byte, mb*1024*1024)
			rand.Read(data)

			metrics.MemoryUsage.Set(float64(mb * 1024 * 1024))
			
			c.JSON(http.StatusOK, gin.H{"allocated": fmt.Sprintf("%dMB", mb)})
		})

		// Test de DDoS
		test.GET("/security/ddos", func(c *gin.Context) {
			metrics.DDoSAttempts.Inc()
			c.JSON(http.StatusForbidden, gin.H{"error": "DDoS attempt detected"})
		})

		// Test de Rate Limit
		test.GET("/security/ratelimit", func(c *gin.Context) {
			metrics.RateLimitExceeded.Inc()
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
		})
	}
}
