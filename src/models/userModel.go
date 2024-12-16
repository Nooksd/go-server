package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id" json:"id"`
	Name         *string            `json:"name" validate:"required, min=5, max=100"`
	Email        *string            `json:"email" validate:"required, email"`
	Password     *string            `json:"password" validate:"required, min=5"`
	UserType     *string            `json:"user_type" validate:"required, eq=ADMIN|eq=USER"`
	UserId       string             `json:"user_id" validate:"required"`
	PhoneNumber  *string            `json:"phone_number"`
	Role         *string            `json:"role"`
	EntryDate    *time.Time         `json:"entry_date"`
	BirthDate    *time.Time         `json:"birth_date"`
	LinkedinURL  *string            `json:"linkedin_url"`
	FacebookURL  *string            `json:"facebook_url"`
	InstagramURL *string            `json:"instagram_url"`
	PTotal       *int               `json:"p_total"`
	PSpent       *int               `json:"p_spent"`
	PCurrent     *int               `json:"p_current"`
}
