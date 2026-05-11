package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Expense struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID `bson:"user_id" json:"user_id"`
	MotorcycleID primitive.ObjectID `bson:"motorcycle_id" json:"motorcycle_id"`

	Category     string             `bson:"category" json:"category"`
	Description  string             `bson:"description" json:"description"`
	Amount       float64            `bson:"amount" json:"amount"`

	Date         time.Time          `bson:"date" json:"date"`

	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}