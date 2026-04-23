package models

import "time"

type User struct {
	ID           string    `bson:"_id,omitempty" json:"id,omitempty"`
	Name         string    `bson:"name" json:"name"`
	Email        string    `bson:"email" json:"email"`
	PasswordHash string    `bson:"password_hash" json:"-"`
	Role         string    `bson:"role" json:"role"` // Add this line
	CreatedAt    time.Time `bson:"created_at" json:"created_at"`
}
