package models

import "time"

type Admin struct {
    ID           string    `bson:"_id,omitempty" json:"id,omitempty"`
    Name         string    `bson:"name" json:"name"`
    Email        string    `bson:"email" json:"email"`
    PasswordHash string    `bson:"password_hash" json:"-"`
    CreatedAt    time.Time `bson:"created_at" json:"created_at"`
}
