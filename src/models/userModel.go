package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID                primitive.ObjectID `bson:"_id" json:"id"`
	Name              *string            `bson:"name" json:"name" validate:"required"`
	Email             *string            `bson:"email" json:"email" validate:"required"`
	Password          *string            `bson:"password" json:"password" validate:"required"`
	UserType          *string            `bson:"userType" json:"userType" validate:"required"`
	Uid               string             `bson:"uid" json:"uid"`
	ProfilePictureUrl string             `bson:"profilePictureUrl" json:"profilePictureUrl"`
	PhoneNumber       *string            `bson:"phoneNumber" json:"phoneNumber"`
	Role              *string            `bson:"role" json:"role"`
	EntryDate         time.Time          `bson:"entryDate" json:"entryDate"`
	Birthday          time.Time          `bson:"birthday" json:"birthday"`
	LinkedinURL       *string            `bson:"linkedinUrl" json:"linkedinUrl"`
	FacebookURL       *string            `bson:"facebookUrl" json:"facebookUrl"`
	InstagramURL      *string            `bson:"instagramUrl" json:"instagramUrl"`
	PTotal            *int               `bson:"pTotal" json:"pTotal"`
	PSpent            *int               `bson:"pSpent" json:"pSpent"`
	PCurrent          *int               `bson:"pCurrent" json:"pCurrent"`
}
