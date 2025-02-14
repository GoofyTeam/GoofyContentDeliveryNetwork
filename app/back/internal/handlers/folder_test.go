package handlers

import (
	"app/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func setupFolderTest(t *testing.T) (*FolderHandler, *gin.Engine) {
	h := NewFolderHandler(testDB)
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	return h, r
}

func TestFolderHandler_CreateFolder(t *testing.T) {
	h, r := setupFolderTest(t)
	clearCollection(t, h.folderCollection)

	userID := primitive.NewObjectID()
	r.POST("/folders", func(c *gin.Context) {
		c.Set("user_id", userID)
		h.CreateFolder(c)
	})

	tests := []struct {
		name       string
		input      models.Folder
		wantStatus int
	}{
		{
			name: "Valid root folder",
			input: models.Folder{
				Name: "Root Folder",
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "Empty name",
			input: models.Folder{
				Name: "",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.input)
			assert.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/folders", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)

			if w.Code == http.StatusCreated {
				var response models.Folder
				err = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)	
				assert.Equal(t, tt.input.Name, response.Name)
				assert.Equal(t, userID, response.UserID)
				assert.Equal(t, 0, response.Depth)
				assert.Equal(t, "/"+tt.input.Name, response.Path)
			}
		})
	}
}

func TestFolderHandler_CreateSubFolder(t *testing.T) {
	h, r := setupFolderTest(t)
	clearCollection(t, h.folderCollection)

	userID := primitive.NewObjectID()
	
	// Création d'un dossier parent
	parentFolder := models.Folder{
		Name:      "Parent",
		UserID:    userID,
		Depth:     0,
		Path:      "/Parent",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	result, err := h.folderCollection.InsertOne(context.Background(), parentFolder)
	assert.NoError(t, err)
	parentID := result.InsertedID.(primitive.ObjectID)

	r.POST("/folders", func(c *gin.Context) {
		c.Set("user_id", userID)
		h.CreateFolder(c)
	})

	tests := []struct {
		name       string
		input      models.Folder
		wantStatus int
	}{
		{
			name: "Valid subfolder",
			input: models.Folder{
				Name:     "Subfolder",
				ParentID: &parentID,
			},
			wantStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.input)
			assert.NoError(t, err)

			req := httptest.NewRequest("POST", "/folders", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusCreated {
				var response models.Folder
				err = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.input.Name, response.Name)
				assert.Equal(t, userID, response.UserID)
				assert.Equal(t, 1, response.Depth)
				assert.Equal(t, "/Parent/Subfolder", response.Path)
			}
		})
	}
}

func TestFolderHandler_ListFolderContents(t *testing.T) {
	h, r := setupFolderTest(t)
	clearCollection(t, h.folderCollection)
	clearCollection(t, h.fileCollection)

	userID := primitive.NewObjectID()
	
	// Création d'un dossier parent
	parentFolder := models.Folder{
		Name:      "Parent",
		UserID:    userID,
		Depth:     0,
		Path:      "/Parent",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	result, err := h.folderCollection.InsertOne(context.Background(), parentFolder)
	assert.NoError(t, err)
	parentID := result.InsertedID.(primitive.ObjectID)

	// Création d'un sous-dossier
	subFolder := models.Folder{
		Name:      "Subfolder",
		UserID:    userID,
		ParentID:  &parentID,
		Depth:     1,
		Path:      "/Parent/Subfolder",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err = h.folderCollection.InsertOne(context.Background(), subFolder)
	assert.NoError(t, err)

	r.GET("/folders/:id", func(c *gin.Context) {
		c.Set("user_id", userID)
		h.ListFolderContents(c)
	})

	tests := []struct {
		name       string
		folderID   string
		wantStatus int
		wantCount  int
	}{
		{
			name:       "Valid folder",
			folderID:   parentID.Hex(),
			wantStatus: http.StatusOK,
			wantCount:  1,
		},
		{
			name:       "Invalid folder ID",
			folderID:   "invalid-id",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Non-existent folder",
			folderID:   primitive.NewObjectID().Hex(),
			wantStatus: http.StatusOK,
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/folders/"+tt.folderID, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var response struct {
					Folders []models.Folder `json:"folders"`
					Files   []models.File   `json:"files"`
				}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(response.Folders))
			}
		})
	}
}

func TestFolderHandler_DeleteFolder(t *testing.T) {
	h, r := setupFolderTest(t)
	clearCollection(t, h.folderCollection)
	clearCollection(t, h.fileCollection)

	userID := primitive.NewObjectID()
	
	// Création d'un dossier parent
	parentFolder := models.Folder{
		Name:      "Parent",
		UserID:    userID,
		Depth:     0,
		Path:      "/Parent",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	result, err := h.folderCollection.InsertOne(context.Background(), parentFolder)
	assert.NoError(t, err)
	parentID := result.InsertedID.(primitive.ObjectID)

	// Création d'un sous-dossier
	subFolder := models.Folder{
		Name:      "Subfolder",
		UserID:    userID,
		ParentID:  &parentID,
		Depth:     1,
		Path:      "/Parent/Subfolder",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err = h.folderCollection.InsertOne(context.Background(), subFolder)
	assert.NoError(t, err)

	r.DELETE("/folders/:id", func(c *gin.Context) {
		c.Set("user_id", userID)
		h.DeleteFolder(c)
	})

	tests := []struct {
		name       string
		folderID   string
		wantStatus int
	}{
		{
			name:       "Valid deletion",
			folderID:   parentID.Hex(),
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid folder ID",
			folderID:   "invalid-id",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Non-existent folder",
			folderID:   primitive.NewObjectID().Hex(),
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/folders/"+tt.folderID, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				// Vérifier que le dossier et ses sous-dossiers ont été supprimés
				count, err := h.folderCollection.CountDocuments(context.Background(), bson.M{
					"$or": []bson.M{
						{"_id": parentID},
						{"path": bson.M{"$regex": "^/Parent/"}},
					},
				})
				assert.NoError(t, err)
				assert.Equal(t, int64(0), count)
			}
		})
	}
}
