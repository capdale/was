package model

import "github.com/google/uuid"

type Collection struct {
	UserUUID uuid.UUID `gorm:""`
}
