package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Comment struct {
	Name      string `json:"name" validate:"required"`
	AvatarURL string `json:"avatarUrl" validate:"required"`
	Text      string `json:"text" validate:"required"`
}

type Post struct {
	ID        primitive.ObjectID `bson:"_id" json:"id"`
	OwnerId   string             `json: "ownerId" validate:"required"`
	Name      string             `json:"name" validate:"required"`
	AvatarURL string             `json:"avatarUrl" validate:"required"`
	Text      string             `json:"text" validate:"required"`
	Hashtags  []string           `json:"hashtags" validate:"max=3"`
	Likes     int                `json:"likes" validate:"gte=0"`
	Comments  []Comment          `json:"comments"`
}
