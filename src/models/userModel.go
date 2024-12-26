package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id" json:"id"`
	Name         *string            `json:"name" validate:"required"`
	Email        *string            `json:"email" validate:"required"`
	Password     *string            `json:"password" validate:"required"`
	UserType     *string            `json:"userType" validate:"required"`
	Uid          string             `json:"uid"`
	PhoneNumber  *string            `json:"phoneNumber"`
	Role         *string            `json:"role"`
	EntryDate    *time.Time         `json:"entryDate"`
	BirthDate    *time.Time         `json:"birthDate"`
	LinkedinURL  *string            `json:"linkedinUrl"`
	FacebookURL  *string            `json:"facebookUrl"`
	InstagramURL *string            `json:"instagramUrl"`
	PTotal       *int               `json:"pTotal"`
	PSpent       *int               `json:"pSpent"`
	PCurrent     *int               `json:"pCurrent"`
}
