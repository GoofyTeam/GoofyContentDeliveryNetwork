package main

import (
	"app/internal/api"
	"app/internal/handlers"
	"app/internal/middleware"
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Chargement des variables d'environnement
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Configuration MongoDB
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://mongodb:27017"
	}

	// Connexion à MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	// Ping de la base de données
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB!")

	db := client.Database("goofy_cdn")

	// Initialisation des handlers
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatal(err)
	}

	authHandler := handlers.NewAuthHandler(db)
	fileHandler := handlers.NewFileHandler(db, uploadDir)
	folderHandler := handlers.NewFolderHandler(db)
	healthHandler := handlers.NewHealthHandler()

	// Configuration de Gin
	r := gin.Default()

	// Ajouter le middleware CORS
	r.Use(middleware.CORSMiddleware())

	// Configuration des routes de test
	api.SetupTestRoutes(r)

	// Routes publiques
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)

	// Routes protégées
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		// Gestion des dossiers
		protected.GET("/folders", folderHandler.ListAllFolders)
		protected.POST("/folders", folderHandler.CreateFolder)
		protected.GET("/folders/:name", folderHandler.ListFolderContents)
		protected.DELETE("/folders/:name", folderHandler.DeleteFolder)

		// Gestion des fichiers
		protected.POST("/files", fileHandler.UploadFile)
		protected.GET("/files/:id", fileHandler.GetFile)
		protected.DELETE("/files/:id", fileHandler.DeleteFile)
	}

	// Configuration du serveur HTTP
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.GET("/health", healthHandler.Health)

	server := &http.Server{
		Addr:         "0.0.0.0:" + port,
		Handler:      r,
		ReadTimeout:  30 * time.Second,  // Augmente le timeout de lecture
		WriteTimeout: 30 * time.Second,  // Augmente le timeout d'écriture
		IdleTimeout:  120 * time.Second, // Augmente le timeout d'inactivité
	}

	log.Printf("Serveur démarré sur le port %s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
