package api

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// SetupTestRoutes configure les routes de test pour le backend
func SetupTestRoutes(r *gin.Engine) {
	test := r.Group("/test")
	{
		// Test de téléchargement de fichiers de différentes tailles
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
				fileSize = 1 * 1024 * 1024 // 1MB par défaut
			}

			// Générer un nom de fichier aléatoire
			randomBytes := make([]byte, 16)
			rand.Read(randomBytes)
			fileName := base64.URLEncoding.EncodeToString(randomBytes)

			c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.bin", fileName))
			c.Header("Content-Type", "application/octet-stream")
			c.Header("Content-Length", fmt.Sprintf("%d", fileSize))

			// Envoyer les données par chunks pour simuler un téléchargement
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

		// Test d'upload de fichiers
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

		// Test de latence variable
		test.GET("/latency/:pattern", func(c *gin.Context) {
			pattern := c.Param("pattern")
			var delay time.Duration

			switch pattern {
			case "spike":
				// Pics de latence aléatoires
				if rand.Float64() < 0.2 { // 20% de chance d'avoir un pic
					delay = time.Duration(rand.Float64()*5000) * time.Millisecond
				} else {
					delay = time.Duration(rand.Float64()*100) * time.Millisecond
				}
			case "wave":
				// Latence sinusoïdale
				t := float64(time.Now().UnixNano()) / float64(time.Second)
				wave := math.Sin(t*math.Pi/30) + 1 // Période de 60 secondes
				delay = time.Duration(wave*500) * time.Millisecond
			case "random":
				// Distribution exponentielle
				delay = time.Duration(rand.ExpFloat64()*1000) * time.Millisecond
			default:
				delay = time.Duration(rand.Float64()*1000) * time.Millisecond
			}

			time.Sleep(delay)
			c.JSON(http.StatusOK, gin.H{
				"pattern": pattern,
				"delay":   delay.String(),
			})
		})

		// Test de streaming de données
		test.GET("/stream/:duration", func(c *gin.Context) {
			duration, err := time.ParseDuration(c.Param("duration") + "s")
			if err != nil {
				duration = 30 * time.Second
			}

			c.Header("Content-Type", "text/event-stream")
			c.Header("Cache-Control", "no-cache")
			c.Header("Connection", "keep-alive")

			start := time.Now()
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()

			c.Stream(func(w io.Writer) bool {
				<-ticker.C
				if time.Since(start) > duration {
					return false
				}

				// Générer des données aléatoires
				data := make([]byte, 1024)
				rand.Read(data)
				msg := base64.StdEncoding.EncodeToString(data)

				c.SSEvent("data", gin.H{
					"timestamp": time.Now().Unix(),
					"data":     msg[:30] + "...", // Tronquer pour la lisibilité
				})
				return true
			})
		})

		// Test de compression
		test.GET("/compression", func(c *gin.Context) {
			// Générer un grand texte répétitif
			var buffer bytes.Buffer
			for i := 0; i < 1000; i++ {
				buffer.WriteString("Ceci est un test de compression. Les données répétitives se compressent bien. ")
			}

			c.Header("Content-Type", "text/plain")
			c.String(http.StatusOK, buffer.String())
		})

		// Test d'erreurs
		test.GET("/error/:type", func(c *gin.Context) {
			errorType := c.Param("type")
			switch errorType {
			case "timeout":
				time.Sleep(30 * time.Second)
			case "memory":
				data := make([]byte, 1024*1024*1024) // Allouer 1GB
				rand.Read(data)
				c.String(http.StatusOK, string(data[:100]))
			case "cpu":
				start := time.Now()
				for time.Since(start) < 10*time.Second {
					// Boucle intensive
					for i := 0; i < 1000000; i++ {
						math.Sqrt(float64(i))
					}
				}
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur interne"})
			}
		})
	}
}
