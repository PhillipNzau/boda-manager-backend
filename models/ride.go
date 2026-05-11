package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Ride struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID `bson:"user_id" json:"user_id"`
	RiderID      primitive.ObjectID `bson:"rider_id" json:"rider_id"`
	MotorcycleID primitive.ObjectID `bson:"motorcycle_id" json:"motorcycle_id"`

	Date         string             `bson:"date" json:"date"`
	Trips        int                `bson:"trips" json:"trips"`
	Income       float64            `bson:"income" json:"income"`
	FuelCost     float64            `bson:"fuel_cost" json:"fuel_cost"`
	Expenses     float64            `bson:"expenses" json:"expenses"`
	Notes        string             `bson:"notes" json:"notes"`

	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}