package model

import (
	"time"

	"github.com/google/uuid"
)

type ImageQueue struct {
	UUID       uuid.UUID `gorm:"type:varchar(36);not null"`
	UserUUID   uuid.UUID `gorm:"type:varchar(36);not null"`
	Longtitude float64   `gorm:"not null"`
	Latitude   float64   `gorm:"not null"`
	Altitude   float64   `gorm:"not null"`
	Accuracy   float64   `gorm:"not null"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}
