package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Payment struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID          primitive.ObjectID `bson:"user_id" json:"user_id"`
	RiderID         primitive.ObjectID `bson:"rider_id" json:"rider_id"`
	Amount          float64            `bson:"amount" json:"amount"`
	Method          string             `bson:"method" json:"method"`
	Date            time.Time          `bson:"date" json:"date"`

	ExpectedAmount  float64            `bson:"expected_amount" json:"expected_amount"`
	Balance         float64            `bson:"balance" json:"balance"` // +ve means debt

	Status          string             `bson:"status" json:"status"` // paid, partial, overpaid

	Notes           string             `bson:"notes" json:"notes"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
	MotorcycleID    string             `bson:"motorcycle_id" json:"motorcycle_id"`
}