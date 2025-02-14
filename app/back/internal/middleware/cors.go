package middleware

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	allowedOrigins := getAllowedOrigins()

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Si l'origine est autorisée ou en mode développement
		if isAllowedOrigin(origin, allowedOrigins) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigins[0]) // Par défaut
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400") // 24 heures

		if c.Request.Method == "OPTIONS" {
			c.Writer.Header().Set("Access-Control-Max-Age", "86400")
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func getAllowedOrigins() []string {
	// En développement, autoriser localhost
	if gin.Mode() == gin.DebugMode {
		return []string{
			"http://localhost:5173",    // Vite dev server
			"http://localhost:5174",    // Vite dev server
			"http://localhost:5175",    // Vite dev server
			"http://localhost:3000",    // Autre port courant
			"http://localhost:3001",    // Autre port courant
			"http://localhost:8080",    // Backend
			"http://localhost",    // Backend
			"http://127.0.0.1:5173",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:8080",
			"http://127.0.0.1",
		}
	}

	// En production, utiliser les origines définies dans les variables d'environnement
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		return []string{"*"} // Fallback, à modifier en production
	}

	return strings.Split(allowedOrigins, ",")
}

func isAllowedOrigin(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	// Si "*" est dans la liste, tout est autorisé
	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			return true
		}
	}

	// Sinon, vérifier si l'origine est dans la liste
	for _, allowed := range allowedOrigins {
		if allowed == origin {
			return true
		}
	}

	return false
}
