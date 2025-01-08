package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Validation struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID      string             `bson:"userId" validate:"required" json:"userId"`
	MissionID   string             `bson:"missionId" validate:"required" json:"missionId"`
	URL         string             `bson:"url" validate:"required,url" json:"url"`
	Status      string             `bson:"status" validate:"required" json:"status"`
	SubmittedAt time.Time          `bson:"submittedAt" json:"submittedAt"`
	ValidatedBy string             `bson:"validatedBy,omitempty" json:"validatedBy,omitempty"`
	ValidatedAt time.Time          `bson:"validatedAt,omitempty" json:"validatedAt,omitempty"`
}
