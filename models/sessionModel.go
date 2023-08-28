package models

import (
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Session struct {
	ID            primitive.ObjectID `bson:"_id"`
	User_id       uuid.UUID          `json:"user_id" validate:"required"`
	Refresh_token *string            `json:"refresh_token"`
	Is_active     bool               `json:"is_active"`
}
