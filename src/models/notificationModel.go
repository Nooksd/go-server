package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Notification struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Text       string             `bson:"text" json:"text" validate:"required"`
	Type       string             `bson:"type" json:"type" validate:"required"`
	Visualized []string           `bson:"visualized" json:"visualized"`
	CreatedAt  time.Time          `bson:"createdAt" json:"createdAt"`
}
