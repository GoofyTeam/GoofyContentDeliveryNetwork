package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type File struct {
	ID        primitive.ObjectID  `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string             `bson:"name" json:"name"`
	Path      string             `bson:"path" json:"path"` // Chemin physique du fichier sur le disque
	Size      int64              `bson:"size" json:"size"`
	MimeType  string             `bson:"mime_type" json:"mime_type"`
	FolderID  primitive.ObjectID `bson:"folder_id" json:"folder_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}
