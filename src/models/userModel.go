package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID                primitive.ObjectID `bson:"_id" json:"id"`
	Name              *string            `json:"name" validate:"required"`
	Email             *string            `json:"email" validate:"required"`
	Password          *string            `json:"password" validate:"required"`
	UserType          *string            `json:"userType" validate:"required"`
	Uid               string             `json:"uid"`
	ProfilePictureUrl string             `json:"profilePictureUrl"`
	PhoneNumber       *string            `json:"phoneNumber"`
	Role              *string            `json:"role"`
	EntryDate         string             `json:"entryDate"`
	Birthday          string             `json:"birthday"`
	LinkedinURL       *string            `json:"linkedinUrl"`
	FacebookURL       *string            `json:"facebookUrl"`
	InstagramURL      *string            `json:"instagramUrl"`
	PTotal            *int               `json:"pTotal"`
	PSpent            *int               `json:"pSpent"`
	PCurrent          *int               `json:"pCurrent"`
}
