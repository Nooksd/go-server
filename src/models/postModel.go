package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Comment struct {
	ID        primitive.ObjectID `bson:"_id" json:"id"`
	OwnerId   string             `bson:"ownerId" json:"ownerId"`
	Name      string             `bson:"name" json:"name" validate:"required"`
	AvatarURL string             `bson:"avatarUrl" json:"avatarUrl" validate:"required"`
	Text      string             `bson:"text" json:"text" validate:"required"`
}

type Post struct {
	ID        primitive.ObjectID `bson:"_id" json:"id"`
	OwnerId   string             `bson:"ownerId" json:"ownerId"`
	Name      string             `bson:"name" json:"name" validate:"required"`
	AvatarURL string             `bson:"avatarUrl" json:"avatarUrl" validate:"required"`
	Role      string             `bson:"role" json:"role" validate:"required"`
	Text      string             `bson:"text" json:"text" validate:"required"`
	Hashtags  []string           `bson:"hashtags" json:"hashtags" validate:"max=3"`
	ImageUrl  string             `bson:"imageUrl" json:"imageUrl"`
	Likes     []string           `bson:"likes" json:"likes" validate:"gte=0"`
	Comments  []Comment          `bson:"comments" json:"comments"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}
