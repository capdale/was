package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ImageQueue struct {
	ID         uint      `gorm:"primaryKey"`
	UUID       uuid.UUID `gorm:"type:varchar(36);not null"`
	UserUUID   uuid.UUID `gorm:"type:varchar(36);not null"`
	Longtitude float64   `gorm:"not null"`
	Latitude   float64   `gorm:"not null"`
	Altitude   float64   `gorm:"not null"`
	Accuracy   float64   `gorm:"not null"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	DeletedAt  gorm.DeletedAt
}
