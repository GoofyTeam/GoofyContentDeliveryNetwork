package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Folder struct {
	ID        primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string              `bson:"name" json:"name" binding:"required"`
	Path      string              `bson:"path" json:"path"` // Chemin complet du dossier (ex: /user1/docs/images)
	ParentID  *primitive.ObjectID `bson:"parent_id,omitempty" json:"parent_id,omitempty"`
	UserID    primitive.ObjectID  `bson:"user_id" json:"user_id"`
	Depth     int                 `bson:"depth" json:"depth"` // Profondeur dans l'arborescence (max 10)
	CreatedAt time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time           `bson:"updated_at" json:"updated_at"`
}
