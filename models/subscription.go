package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Subscription struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	ServiceName string             `bson:"service_name" json:"service_name"`
	PlanName    string             `bson:"plan_name" json:"plan_name"`
	StartDate   *time.Time         `bson:"start_date,omitempty" json:"start_date,omitempty"`
	RenewalDate *time.Time         `bson:"renewal_date,omitempty" json:"renewal_date,omitempty"`
	Price       float64            `bson:"price" json:"price"`
	Currency    string             `bson:"currency" json:"currency"`
	Status      string             `bson:"status" json:"status"`
	Notes       string             `bson:"notes,omitempty" json:"notes,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}