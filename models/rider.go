package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Rider struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID          primitive.ObjectID `bson:"user_id" json:"user_id"`

	FullName        string             `bson:"full_name" json:"full_name"`
	PhoneNumber     string             `bson:"phone_number" json:"phone_number"`
	LicenseNo       string             `bson:"license_no" json:"license_no"`

	DailyTarget     float64            `bson:"daily_target" json:"daily_target"`

	MotorcycleID    primitive.ObjectID `bson:"motorcycle_id" json:"motorcycle_id"`

	OutstandingDebt float64            `bson:"outstanding_debt" json:"outstanding_debt"`
	LastPaymentDate *time.Time         `bson:"last_payment_date,omitempty" json:"last_payment_date,omitempty"`

	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}