package handlers

import (
	"app/internal/models"
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func setupFileTest(t *testing.T) (*FileHandler, *gin.Engine, string) {
	// Création d'un dossier temporaire pour les uploads
	tempDir, err := os.MkdirTemp("", "file_test")
	if err != nil {
		t.Fatalf("Erreur lors de la création du dossier temporaire: %v", err)
	}

	// Configuration du handler
	h := NewFileHandler(testDB, tempDir)
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	return h, r, tempDir
}

func createTestFolder(t *testing.T, h *FileHandler, userID primitive.ObjectID) primitive.ObjectID {
	folder := models.Folder{
		Name:      "Test Folder",
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err := h.folderCollection.InsertOne(context.Background(), folder)
	if err != nil {
		t.Fatalf("Erreur lors de la création du dossier de test: %v", err)
	}

	return result.InsertedID.(primitive.ObjectID)
}

func TestFileHandler_UploadFile(t *testing.T) {
	h, r, tempDir := setupFileTest(t)
	defer os.RemoveAll(tempDir)

	userID := primitive.NewObjectID()
	folderID := createTestFolder(t, h, userID)

	r.POST("/upload", func(c *gin.Context) {
		c.Set("user_id", userID)
		h.UploadFile(c)
	})

	tests := []struct {
		name       string
		setup      func() (*bytes.Buffer, *multipart.Writer)
		wantStatus int
	}{
		{
			name: "Valid upload",
			setup: func() (*bytes.Buffer, *multipart.Writer) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("file", "test.txt")
				part.Write([]byte("test content"))
				writer.WriteField("folder_id", folderID.Hex())
				writer.Close()
				return body, writer
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "Missing file",
			setup: func() (*bytes.Buffer, *multipart.Writer) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				writer.WriteField("folder_id", folderID.Hex())
				writer.Close()
				return body, writer
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid folder ID",
			setup: func() (*bytes.Buffer, *multipart.Writer) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("file", "test.txt")
				part.Write([]byte("test content"))
				writer.WriteField("folder_id", "invalid-id")
				writer.Close()
				return body, writer
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, writer := tt.setup()
			req := httptest.NewRequest("POST", "/upload", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestFileHandler_GetFile(t *testing.T) {
	h, r, tempDir := setupFileTest(t)
	defer os.RemoveAll(tempDir)

	userID := primitive.NewObjectID()
	folderID := createTestFolder(t, h, userID)

	// Création d'un fichier de test
	testFilePath := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFilePath, []byte("test content"), 0644)
	assert.NoError(t, err)

	file := models.File{
		Name:      "test.txt",
		Path:      testFilePath,
		Size:      12,
		MimeType:  "text/plain",
		FolderID:  folderID,
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err := h.fileCollection.InsertOne(context.Background(), file)
	assert.NoError(t, err)
	fileID := result.InsertedID.(primitive.ObjectID)

	r.GET("/files/:id", func(c *gin.Context) {
		c.Set("user_id", userID)
		h.GetFile(c)
	})

	tests := []struct {
		name       string
		fileID     string
		wantStatus int
	}{
		{
			name:       "Valid file",
			fileID:     fileID.Hex(),
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid file ID",
			fileID:     "invalid-id",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Non-existent file",
			fileID:     primitive.NewObjectID().Hex(),
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/files/"+tt.fileID, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestFileHandler_DeleteFile(t *testing.T) {
	h, r, tempDir := setupFileTest(t)
	defer os.RemoveAll(tempDir)

	userID := primitive.NewObjectID()
	folderID := createTestFolder(t, h, userID)

	// Création d'un fichier de test
	testFilePath := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFilePath, []byte("test content"), 0644)
	assert.NoError(t, err)

	file := models.File{
		Name:      "test.txt",
		Path:      testFilePath,
		Size:      12,
		MimeType:  "text/plain",
		FolderID:  folderID,
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err := h.fileCollection.InsertOne(context.Background(), file)
	assert.NoError(t, err)
	fileID := result.InsertedID.(primitive.ObjectID)

	r.DELETE("/files/:id", func(c *gin.Context) {
		c.Set("user_id", userID)
		h.DeleteFile(c)
	})

	tests := []struct {
		name       string
		fileID     string
		wantStatus int
	}{
		{
			name:       "Valid deletion",
			fileID:     fileID.Hex(),
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid file ID",
			fileID:     "invalid-id",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Non-existent file",
			fileID:     primitive.NewObjectID().Hex(),
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/files/"+tt.fileID, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				// Vérifier que le fichier a bien été supprimé de la base de données
				var count int64
				count, err = h.fileCollection.CountDocuments(context.Background(), bson.M{"_id": fileID})
				assert.NoError(t, err)
				assert.Equal(t, int64(0), count)

				// Vérifier que le fichier a bien été supprimé du système de fichiers
				_, err = os.Stat(testFilePath)
				assert.True(t, os.IsNotExist(err))
			}
		})
	}
}
