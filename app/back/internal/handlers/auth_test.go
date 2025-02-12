package handlers

import (
	"app/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var testDB *mongo.Database

func setupTestDB() *mongo.Database {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("Erreur de connexion à MongoDB:", err)
	}

	db := client.Database("goofycdn_test")
	return db
}

func cleanupTestDB(db *mongo.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Suppression de toutes les collections de test
	if err := db.Collection("users").Drop(ctx); err != nil {
		log.Printf("Erreur lors du nettoyage de la collection users: %v", err)
	}
}

func TestMain(m *testing.M) {
	// Initialisation de la base de données de test
	testDB = setupTestDB()

	// Exécution des tests
	code := m.Run()

	// Nettoyage après les tests
	cleanupTestDB(testDB)

	os.Exit(code)
}

func clearCollection(t *testing.T, collection *mongo.Collection) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	_, err := collection.DeleteMany(ctx, bson.M{})
	if err != nil {
		t.Fatalf("Erreur lors du nettoyage de la collection: %v", err)
	}
}

// createTestUser crée un utilisateur de test dans la base de données
func createTestUser(t *testing.T, h *AuthHandler, email, password string) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Erreur lors du hashage du mot de passe: %v", err)
	}

	user := models.User{
		Email:     email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = h.userCollection.InsertOne(context.Background(), user)
	if err != nil {
		t.Fatalf("Erreur lors de la création de l'utilisateur de test: %v", err)
	}
}

func TestAuthHandler_Register(t *testing.T) {
	// Nettoyage de la collection avant le test
	clearCollection(t, testDB.Collection("users"))

	// Configuration du mode test pour Gin
	gin.SetMode(gin.TestMode)

	// Création d'un router de test
	r := gin.Default()
	h := NewAuthHandler(testDB)
	r.POST("/register", h.Register)

	tests := []struct {
		name       string
		input      models.User
		wantStatus int
	}{
		{
			name: "Valid registration",
			input: models.User{
				// Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "Invalid email",
			input: models.User{
				// Username: "testuser2",
				Email:    "invalid-email",
				Password: "password123",
			},
			wantStatus: http.StatusBadRequest,
		},
		// Ajoutez d'autres cas de test selon vos besoins
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Conversion de l'input en JSON
			jsonInput, err := json.Marshal(tt.input)
			assert.NoError(t, err)

			// Création de la requête
			req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonInput))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Exécution de la requête
			r.ServeHTTP(w, req)

			// Vérification du statut
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	// Configuration du mode test pour Gin
	gin.SetMode(gin.TestMode)

	// Création d'un router de test
	r := gin.Default()
	h := NewAuthHandler(testDB)
	r.POST("/login", h.Login)

	tests := []struct {
		name       string
		input      models.LoginRequest
		setupUser  bool
		wantStatus int
	}{
		{
			name: "Valid login",
			input: models.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupUser:  true,
			wantStatus: http.StatusOK,
		},
		{
			name: "Invalid credentials",
			input: models.LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			setupUser:  true,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "Invalid email format",
			input: models.LoginRequest{
				Email:    "invalid-email",
				Password: "password123",
			},
			setupUser:  false,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Nettoyage de la collection avant chaque test
			clearCollection(t, testDB.Collection("users"))

			if tt.setupUser {
				createTestUser(t, h, tt.input.Email, "password123")
			}

			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Enregistrement de la réponse
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
