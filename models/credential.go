package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Credential struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID            primitive.ObjectID `bson:"user_id" json:"user_id"`
	SiteName          string             `bson:"site_name" json:"site_name"`
	Username          string             `bson:"username" json:"username"`
	PasswordEncrypted string             `bson:"password_encrypted" json:"-"`
	LoginURL          string             `bson:"login_url" json:"login_url"`
	Notes             string             `bson:"notes,omitempty" json:"notes,omitempty"`
	Category          string             `bson:"category,omitempty" json:"category,omitempty"`
	CreatedAt         time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time          `bson:"updated_at" json:"updated_at"`
}
