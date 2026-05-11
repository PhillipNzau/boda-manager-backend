package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Hub model - stores bookmarks, IPs, notes, etc.
type Hub struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Title     string             `bson:"title" json:"title"`
	Type      string             `bson:"type" json:"type"` 
	Value     string             `bson:"value" json:"value"` 
	Notes     string             `bson:"notes,omitempty" json:"notes,omitempty"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}
