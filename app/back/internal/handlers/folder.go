package handlers

import (
	"app/internal/models"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

	// Vérification que le nom du dossier n'existe pas déjà pour cet utilisateur
	var existingFolder models.Folder
	err := h.folderCollection.FindOne(c, bson.M{
		"name": folder.Name,
		"user_id": folder.UserID,
	}).Decode(&existingFolder)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Un dossier avec ce nom existe déjà"})
		return
	} else if err != mongo.ErrNoDocuments {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la vérification du nom du dossier"})
		return
	}

	// Vérification de la profondeur maximale
	if folder.ParentID != nil {
		var parentFolder models.Folder
		err := h.folderCollection.FindOne(c, bson.M{"_id": folder.ParentID}).Decode(&parentFolder)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Parent folder not found"})
			return
		}

		// Construction du chemin complet
		folder.Path = path.Join(parentFolder.Path, folder.Name)

		// Vérification de la profondeur maximale 10
		if len(strings.Split(folder.Path, "/")) > 10 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Profondeur maximale de dossiers atteinte"})
			return
		}
		folder.Depth = parentFolder.Depth + 1
	} else {
		folder.Depth = 0
		folder.Path = "/" + folder.Name
	}

	folder.CreatedAt = time.Now()
	folder.UpdatedAt = time.Now()

	// Insertion du dossier dans la base de données
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
	folderName := c.Param("name")
	userID, _ := c.Get("user_id")
	userIDObj := userID.(primitive.ObjectID)

	// Récupération du dossier par son nom
	var folder models.Folder
	err := h.folderCollection.FindOne(c, bson.M{
		"name": folderName,
		"user_id": userIDObj,
	}).Decode(&folder)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find folder"})
		}
		return
	}

	// Récupération des sous-dossiers
	var folders []models.Folder
	folderCursor, err := h.folderCollection.Find(c, bson.M{
		"parent_id": folder.ID,
		"user_id":   userIDObj,
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
		"folder_id": folder.ID,
		"user_id":   userIDObj,
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

// ListAllFolders liste tous les dossiers de l'utilisateur
func (h *FolderHandler) ListAllFolders(c *gin.Context) {
	// Récupérer l'ID de l'utilisateur depuis le contexte
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	// L'ID est déjà un ObjectID depuis le middleware
	objectID, ok := userID.(primitive.ObjectID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}

	// Trouver tous les dossiers de l'utilisateur
	cursor, err := h.folderCollection.Find(c, bson.M{"user_id": objectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch folders"})
		return
	}
	defer cursor.Close(c)

	var folders []models.Folder
	if err := cursor.All(c, &folders); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode folders"})
		return
	}

	c.JSON(http.StatusOK, folders)
}

// DeleteFolder supprime un dossier et son contenu
func (h *FolderHandler) DeleteFolder(c *gin.Context) {
	folderName := c.Param("name")
	userID, _ := c.Get("user_id")
	userIDObj := userID.(primitive.ObjectID)

	// Récupération du dossier par son nom
	var folder models.Folder
	err := h.folderCollection.FindOne(c, bson.M{
		"name": folderName,
		"user_id": userIDObj,
	}).Decode(&folder)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find folder"})
		}
		return
	}

	// Vérification que ce n'est pas le dossier racine
	if folder.Name == "root" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete root folder"})
		return
	}

	// Suppression récursive des sous-dossiers et fichiers
	if err := h.deleteSubFoldersAndFiles(c, folder.ID, userIDObj); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete folder contents"})
		return
	}

	// Suppression du dossier lui-même
	_, err = h.folderCollection.DeleteOne(c, bson.M{
		"_id": folder.ID,
		"user_id": userIDObj,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete folder"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Folder deleted successfully"})
}

func (h *FolderHandler) deleteSubFoldersAndFiles(c *gin.Context, folderID primitive.ObjectID, userID primitive.ObjectID) error {
	// Suppression récursive des sous-dossiers et fichiers
	_, err := h.folderCollection.DeleteMany(c, bson.M{
		"path": bson.M{"$regex": "^" + folderID.Hex() + "/"},
		"user_id": userID,
	})
	if err != nil {
		return err
	}

	_, err = h.fileCollection.DeleteMany(c, bson.M{
		"folder_id": folderID,
		"user_id":   userID,
	})
	if err != nil {
		return err
	}

	return nil
}
