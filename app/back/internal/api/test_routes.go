package api

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

// SetupTestRoutes configure les routes de test pour le backend
func SetupTestRoutes(r *gin.Engine) {
	test := r.Group("/test")
	{
		// Test de cache avec contenu statique
		test.GET("/cache/static/:id", func(c *gin.Context) {
			id := c.Param("id")
			content := fmt.Sprintf("Contenu statique pour l'ID %s - Timestamp: %d", id, time.Now().Unix())
			c.Header("Cache-Control", "public, max-age=60")
			c.String(http.StatusOK, content)
		})

		// Test de latence avec différents patterns
		test.GET("/latency/:pattern", func(c *gin.Context) {
			pattern := c.Param("pattern")
			var delay time.Duration

			switch pattern {
			case "random":
				delay = time.Duration(float64(time.Second) * (float64(time.Now().UnixNano()%1000) / 1000.0))
			case "spike":
				if time.Now().UnixNano()%10 == 0 { // 10% de chance d'avoir un pic
					delay = 2 * time.Second
				}
			case "wave":
				t := float64(time.Now().Unix())
				// Génère une latence sinusoïdale entre 100ms et 1s
				factor := (math.Sin(t/10) + 1) / 2
				delay = time.Duration(100+900*factor) * time.Millisecond
			default:
				delay = 100 * time.Millisecond
			}

			time.Sleep(delay)
			c.JSON(http.StatusOK, gin.H{
				"pattern": pattern,
				"delay":   delay.String(),
			})
		})

		// Test de téléchargement
		test.GET("/download/:size", func(c *gin.Context) {
			size := c.Param("size")
			var fileSize int64

			switch size {
			case "small":
				fileSize = 1 * 1024 * 1024 // 1MB
			case "medium":
				fileSize = 10 * 1024 * 1024 // 10MB
			case "large":
				fileSize = 100 * 1024 * 1024 // 100MB
			default:
				fileSize = 1 * 1024 * 1024
			}

			// Générer un nom de fichier aléatoire
			randomBytes := make([]byte, 16)
			rand.Read(randomBytes)
			fileName := base64.URLEncoding.EncodeToString(randomBytes)

			c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.bin", fileName))
			c.Header("Content-Type", "application/octet-stream")
			c.Header("Content-Length", fmt.Sprintf("%d", fileSize))
			c.Header("Cache-Control", "no-cache")

			// Envoyer les données par chunks
			chunkSize := int64(64 * 1024) // 64KB chunks
			remaining := fileSize

			for remaining > 0 {
				if remaining < chunkSize {
					chunkSize = remaining
				}
				chunk := make([]byte, chunkSize)
				rand.Read(chunk)
				c.Writer.Write(chunk)
				c.Writer.Flush()
				remaining -= chunkSize
				time.Sleep(time.Millisecond * 10) // Simuler une latence réseau
			}
		})

		// Test de compression
		test.GET("/compression", func(c *gin.Context) {
			// Générer un texte très compressible
			var buffer bytes.Buffer
			for i := 0; i < 1024*1024; i++ { // 1MB de données
				buffer.WriteByte('a' + byte(i%26))
			}
			c.String(http.StatusOK, buffer.String())
		})

		// Test d'upload
		test.POST("/upload", func(c *gin.Context) {
			file, err := c.FormFile("file")
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Fichier manquant"})
				return
			}

			// Créer le dossier uploads s'il n'existe pas
			uploadDir := "uploads"
			if err := os.MkdirAll(uploadDir, 0755); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la création du dossier"})
				return
			}

			// Sauvegarder le fichier
			dst := filepath.Join(uploadDir, file.Filename)
			if err := c.SaveUploadedFile(file, dst); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la sauvegarde"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status": "success",
				"file":   file.Filename,
				"size":   file.Size,
			})
		})

		// Test de streaming
		test.GET("/stream/:seconds", func(c *gin.Context) {
			c.Header("Content-Type", "text/event-stream")
			c.Header("Cache-Control", "no-cache")
			c.Header("Connection", "keep-alive")
			c.Header("Transfer-Encoding", "chunked")

			duration := 10 // Durée par défaut en secondes
			fmt.Sscanf(c.Param("seconds"), "%d", &duration)

			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()

			for i := 0; i < duration; i++ {
				select {
				case <-ticker.C:
					c.SSEvent("message", fmt.Sprintf("Événement %d/%d", i+1, duration))
					c.Writer.Flush()
				case <-c.Request.Context().Done():
					return
				}
			}
		})

		// Endpoint de santé pour le CDN
		test.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "healthy",
				"time":   time.Now().Unix(),
			})
		})
	}
}
