package handlers

import (
	"app/internal/models"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FileHandler struct {
	fileCollection   *mongo.Collection
	folderCollection *mongo.Collection
	uploadDir        string
}

func NewFileHandler(db *mongo.Database, uploadDir string) *FileHandler {
	return &FileHandler{
		fileCollection:   db.Collection("files"),
		folderCollection: db.Collection("folders"),
		uploadDir:        uploadDir,
	}
}

// UploadFile gère l'upload d'un fichier
func (h *FileHandler) UploadFile(c *gin.Context) {
	// Récupération du fichier depuis la requête multipart
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}
	defer file.Close()

	// Récupération de l'ID du dossier parent
	folderID, err := primitive.ObjectIDFromHex(c.PostForm("folder_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid folder ID"})
		return
	}

	userID, _ := c.Get("user_id")

	// Vérification que le dossier existe et appartient à l'utilisateur
	var folder models.Folder
	err = h.folderCollection.FindOne(c, bson.M{
		"_id":     folderID,
		"user_id": userID,
	}).Decode(&folder)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}

	// Création du chemin de stockage
	userDir := filepath.Join(h.uploadDir, userID.(primitive.ObjectID).Hex())
	if err := os.MkdirAll(userDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Génération d'un nom de fichier unique
	filename := primitive.NewObjectID().Hex() + filepath.Ext(header.Filename)
	filePath := filepath.Join(userDir, filename)

	// Sauvegarde du fichier sur le disque
	out, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file"})
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Création de l'entrée dans la base de données
	fileDoc := models.File{
		Name:      header.Filename,
		Path:      filePath,
		Size:      header.Size,
		MimeType:  header.Header.Get("Content-Type"),
		FolderID:  folderID,
		UserID:    userID.(primitive.ObjectID),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err := h.fileCollection.InsertOne(c, fileDoc)
	if err != nil {
		// Suppression du fichier en cas d'erreur
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file metadata"})
		return
	}

	fileDoc.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusCreated, fileDoc)
}

// GetFile récupère un fichier
func (h *FileHandler) GetFile(c *gin.Context) {
	fileID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	userID, _ := c.Get("user_id")

	var file models.File
	err = h.fileCollection.FindOne(c, bson.M{
		"_id":     fileID,
		"user_id": userID,
	}).Decode(&file)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.File(file.Path)
}

// DeleteFile supprime un fichier
func (h *FileHandler) DeleteFile(c *gin.Context) {
	fileID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	userID, _ := c.Get("user_id")

	var file models.File
	err = h.fileCollection.FindOne(c, bson.M{
		"_id":     fileID,
		"user_id": userID,
	}).Decode(&file)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Suppression du fichier physique
	if err := os.Remove(file.Path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
		return
	}

	// Suppression de l'entrée dans la base de données
	_, err = h.fileCollection.DeleteOne(c, bson.M{
		"_id":     fileID,
		"user_id": userID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file metadata"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}
