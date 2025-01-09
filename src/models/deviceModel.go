package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DeviceToken struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserId            string             `bson:"userId" json:"userId" validate:"required"`
	DeviceToken       string             `bson:"deviceToken" json:"deviceToken" validate:"required"`
	NotificationTypes []string           `bson:"notificationTypes" json:"notificationTypes"`
}
