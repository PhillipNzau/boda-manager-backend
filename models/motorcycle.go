package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Motorcycle struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID `bson:"user_id" json:"user_id"`

	Name         string             `bson:"name" json:"name"`               // e.g Boxer 150
	PlateNumber  string             `bson:"plate_number" json:"plate_number"`
	Model        string             `bson:"model" json:"model"`
	EngineNumber string             `bson:"engine_number" json:"engine_number"`

	Status       string             `bson:"status" json:"status"` // active, repair, inactive

	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}