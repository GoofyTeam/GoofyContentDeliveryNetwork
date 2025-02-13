package handlers

import (
	"app/internal/models"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userCollection *mongo.Collection
}

func NewAuthHandler(db *mongo.Database) *AuthHandler {
	return &AuthHandler{
		userCollection: db.Collection("users"),
	}
}

// isValidEmail vérifie si l'email est valide
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// Register crée un nouveau compte utilisateur
func (h *AuthHandler) Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validation de l'email
	if !isValidEmail(user.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	// Vérification si l'email existe déjà
	var existingUser models.User
	err := h.userCollection.FindOne(c, bson.M{"email": user.Email}).Decode(&existingUser)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	}

	// Hashage du mot de passe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Log pour le débogage
	fmt.Printf("Register - Password length: %d, Hash length: %d\n", len(user.Password), len(hashedPassword))

	// Création de l'utilisateur
	user.Password = string(hashedPassword)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	result, err := h.userCollection.InsertOne(c, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	user.Password = "" // Ne pas renvoyer le mot de passe

	c.JSON(http.StatusCreated, user)
}

// Login authentifie un utilisateur et renvoie un token JWT
func (h *AuthHandler) Login(c *gin.Context) {
	var loginReq models.LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validation de l'email
	if !isValidEmail(loginReq.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	var user models.User
	err := h.userCollection.FindOne(c, bson.M{"email": loginReq.Email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non trouvé"})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Erreur lors de la recherche de l'utilisateur: " + err.Error()})
		}
		return
	}

	// Log pour le débogage
	fmt.Printf("Login - Stored hash length: %d, Input password length: %d\n", len(user.Password), len(loginReq.Password))
	fmt.Printf("Login - Input password: %s\n", loginReq.Password)

	// Vérification du mot de passe
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Mot de passe incorrect"})
		return
	}

	// Création du token JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.Hex(),
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
	})
}
