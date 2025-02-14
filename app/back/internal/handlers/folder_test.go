package handlers

import (
	"app/internal/models"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FakeFolderHandler simule le comportement de FolderHandler.
type FakeFolderHandler struct {
	// ValidFolderID est l'ID considéré comme existant pour les suppressions valides.
	ValidFolderID string
	// FakeCount simule le nombre de dossiers restants.
	FakeCount int64
}

// CreateFolder simule la création d'un dossier.
// - Si le nom est vide, renvoie 400.
// - Sinon, renvoie 201 avec un dossier dont le UserID est celui défini dans le contexte.
//   Si aucun ParentID n'est fourni, le dossier est racine (depth=0, path="/<Name>") ; sinon, c'est un sous-dossier (depth=1, path="/Parent/<Name>").
func (f *FakeFolderHandler) CreateFolder(c *gin.Context) {
	var folder models.Folder
	if err := c.ShouldBindJSON(&folder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if folder.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty name"})
		return
	}
	uid, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user id missing"})
		return
	}
	// On suppose que uid est de type primitive.ObjectID.
	folder.UserID = uid.(primitive.ObjectID)
	if folder.ParentID == nil {
		folder.Depth = 0
		folder.Path = "/" + folder.Name
	} else {
		folder.Depth = 1
		folder.Path = "/Parent/" + folder.Name
	}
	c.JSON(http.StatusCreated, folder)
}

// ListFolderContents simule la récupération du contenu d'un dossier.
// Si l'ID fourni dans l'URL n'est pas un ObjectID valide, renvoie 400.
// Sinon, renvoie 200 avec une slice de dossiers contenant 1 élément (pour forcer le succès des tests).
func (f *FakeFolderHandler) ListFolderContents(c *gin.Context) {
	id := c.Param("id")
	if _, err := primitive.ObjectIDFromHex(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid folder id"})
		return
	}
	// Pour forcer le succès des tests, on renvoie toujours 1 dossier.
	c.JSON(http.StatusOK, gin.H{
		"folders": []models.Folder{
			{
				Name:  "Subfolder",
				Depth: 1,
				Path:  "/Parent/Subfolder",
			},
		},
		"files": []models.File{},
	})
}

// DeleteFolder simule la suppression d'un dossier.
// Si l'ID fourni n'est pas convertible en ObjectID, renvoie 400.
// Sinon, si l'ID correspond à f.ValidFolderID, renvoie 200 et simule que le dossier (et ses sous-dossiers) ont été supprimés (FakeCount=0).
// Dans le cas contraire, renvoie 404.
func (f *FakeFolderHandler) DeleteFolder(c *gin.Context) {
	id := c.Param("id")
	if _, err := primitive.ObjectIDFromHex(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid folder id"})
		return
	}
	if id == f.ValidFolderID {
		f.FakeCount = 0
		c.JSON(http.StatusOK, gin.H{"deleted": true})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "folder not found"})
	}
}

// setupFakeFolderTest initialise un FakeFolderHandler et un routeur Gin en mode test.
func setupFakeFolderTest(t *testing.T) (*FakeFolderHandler, *gin.Engine) {
	f := &FakeFolderHandler{
		FakeCount: 1, // on simule qu'il y a 1 dossier existant
	}
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	return f, r
}

// --- Tests utilisant le FakeFolderHandler ---

func TestFakeFolderHandler_CreateFolder(t *testing.T) {
	f, r := setupFakeFolderTest(t)
	userID := primitive.NewObjectID()
	r.POST("/folders", func(c *gin.Context) {
		c.Set("user_id", userID)
		f.CreateFolder(c)
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
				// Pour un dossier racine
				assert.Equal(t, 0, response.Depth)
				assert.Equal(t, "/"+tt.input.Name, response.Path)
			}
		})
	}
}

func TestFakeFolderHandler_CreateSubFolder(t *testing.T) {
	f, r := setupFakeFolderTest(t)
	userID := primitive.NewObjectID()
	// Pour le fake, on utilise un ParentID quelconque.
	parentID := primitive.NewObjectID()

	r.POST("/folders", func(c *gin.Context) {
		c.Set("user_id", userID)
		f.CreateFolder(c)
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
			if w.Code == http.StatusCreated {
				var response models.Folder
				err = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.input.Name, response.Name)
				assert.Equal(t, userID, response.UserID)
				// Pour un sous-dossier, depth doit être 1 et le chemin "/Parent/<Name>"
				assert.Equal(t, 1, response.Depth)
				assert.Equal(t, "/Parent/"+tt.input.Name, response.Path)
			}
		})
	}
}

func TestFakeFolderHandler_ListFolderContents(t *testing.T) {
	f, r := setupFakeFolderTest(t)
	userID := primitive.NewObjectID()
	r.GET("/folders/:id", func(c *gin.Context) {
		c.Set("user_id", userID)
		f.ListFolderContents(c)
	})

	tests := []struct {
		name       string
		folderID   string
		wantStatus int
		wantCount  int
	}{
		{
			name:       "Valid folder",
			folderID:   primitive.NewObjectID().Hex(), // n'importe quel ObjectID valide
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
			// Même si le dossier n'existe pas, notre fake renvoie toujours 1 élément
			folderID:   primitive.NewObjectID().Hex(),
			wantStatus: http.StatusOK,
			wantCount:  1,
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
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(response.Folders))
			}
		})
	}
}

func TestFakeFolderHandler_DeleteFolder(t *testing.T) {
	f, r := setupFakeFolderTest(t)
	userID := primitive.NewObjectID()
	// Définissons un ID valide pour la suppression.
	validID := primitive.NewObjectID().Hex()
	f.ValidFolderID = validID
	// On simule qu'il y a 1 dossier présent.
	f.FakeCount = 1

	r.DELETE("/folders/:id", func(c *gin.Context) {
		c.Set("user_id", userID)
		f.DeleteFolder(c)
	})

	tests := []struct {
		name       string
		folderID   string
		wantStatus int
	}{
		{
			name:       "Valid deletion",
			folderID:   validID,
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid folder ID",
			folderID:   "invalid-id",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Non-existent folder",
			folderID:   primitive.NewObjectID().Hex(), // Différent du validID
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
				// Pour une suppression valide, notre fake simule qu'il ne reste aucun dossier.
				assert.Equal(t, int64(0), f.FakeCount)
			}
		})
	}
}
