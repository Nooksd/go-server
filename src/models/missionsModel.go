package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Missions struct {
	ID        primitive.ObjectID `bson:"_id" json:"id"`
	OwnerId   string             `bson:"ownerId" json:"ownerId"`
	Text      string             `json:"text" validate:"required"`
	Duration  int                `json:"duration" validate:"required"`
	EndDate   time.Time          `json:"endDate" validate:"required"`
	Value     int                `json:"value" validate:"required"`
	Completed []string           `json:"completed"`
	CreatedAt time.Time          `json:"date" validate:"required"`
}
