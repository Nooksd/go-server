package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Missions struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	OwnerId     string             `bson:"ownerId" json:"ownerId" validate:"required"`
	Text        string             `bson:"text" json:"text" validate:"required"`
	MissionType string             `bson:"missionType" json:"missionType" validate:"required"`
	Hashtag     string             `bson:"hashtag" json:"hashtag"`
	Duration    int                `bson:"duration" json:"duration" validate:"required"`
	EndDate     time.Time          `bson:"endDate" json:"endDate" validate:"required"`
	Value       int                `bson:"value" json:"value" validate:"required"`
	Completed   []string           `bson:"completed" json:"completed"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt" validate:"required"`
}
