package handlers

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"path"
	"time"
	"app/internal/models"
)

type FolderHandler struct {
	folderCollection *mongo.Collection
	fileCollection   *mongo.Collection
}

func NewFolderHandler(db *mongo.Database) *FolderHandler {
	return &FolderHandler{
		folderCollection: db.Collection("folders"),
		fileCollection:   db.Collection("files"),
	}
}

// CreateFolder crée un nouveau dossier
func (h *FolderHandler) CreateFolder(c *gin.Context) {
	var folder models.Folder
	if err := c.ShouldBindJSON(&folder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validation du nom du dossier
	if folder.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Folder name cannot be empty"})
		return
	}

	// Récupération de l'ID utilisateur depuis le token JWT
	userID, _ := c.Get("user_id")
	folder.UserID = userID.(primitive.ObjectID)

	// Vérification de la profondeur maximale
	if folder.ParentID != nil {
		var parentFolder models.Folder
		err := h.folderCollection.FindOne(c, bson.M{"_id": folder.ParentID}).Decode(&parentFolder)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Parent folder not found"})
			return
		}

		if parentFolder.Depth >= 9 { // Max depth = 10
			c.JSON(http.StatusBadRequest, gin.H{"error": "Maximum folder depth reached"})
			return
		}
		folder.Depth = parentFolder.Depth + 1
		folder.Path = path.Join(parentFolder.Path, folder.Name)
	} else {
		folder.Depth = 0
		folder.Path = "/" + folder.Name
	}

	folder.CreatedAt = time.Now()
	folder.UpdatedAt = time.Now()

	result, err := h.folderCollection.InsertOne(c, folder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create folder"})
		return
	}

	folder.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusCreated, folder)
}

// ListFolderContents liste le contenu d'un dossier
func (h *FolderHandler) ListFolderContents(c *gin.Context) {
	folderID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid folder ID"})
		return
	}

	userID, _ := c.Get("user_id")

	// Récupération des sous-dossiers
	var folders []models.Folder
	folderCursor, err := h.folderCollection.Find(c, bson.M{
		"parent_id": folderID,
		"user_id":   userID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list folders"})
		return
	}
	if err = folderCursor.All(c, &folders); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode folders"})
		return
	}

	// Récupération des fichiers
	var files []models.File
	fileCursor, err := h.fileCollection.Find(c, bson.M{
		"folder_id": folderID,
		"user_id":   userID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list files"})
		return
	}
	if err = fileCursor.All(c, &files); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode files"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"folders": folders,
		"files":   files,
	})
}

// DeleteFolder supprime un dossier et son contenu
func (h *FolderHandler) DeleteFolder(c *gin.Context) {
	folderID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid folder ID"})
		return
	}

	userID, _ := c.Get("user_id")

	// Vérification que le dossier appartient à l'utilisateur
	var folder models.Folder
	err = h.folderCollection.FindOne(c, bson.M{
		"_id":     folderID,
		"user_id": userID,
	}).Decode(&folder)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}

	// Suppression récursive des sous-dossiers et fichiers
	_, err = h.folderCollection.DeleteMany(c, bson.M{
		"path": bson.M{"$regex": "^" + folder.Path + "/"},
		"user_id": userID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete subfolders"})
		return
	}

	_, err = h.fileCollection.DeleteMany(c, bson.M{
		"folder_id": folderID,
		"user_id":   userID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete files"})
		return
	}

	// Suppression du dossier lui-même
	_, err = h.folderCollection.DeleteOne(c, bson.M{
		"_id":     folderID,
		"user_id": userID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete folder"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Folder deleted successfully"})
}
