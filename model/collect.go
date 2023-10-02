package model

import (
	"time"

	"github.com/google/uuid"
)

type Collection struct {
	UserUUID        uuid.UUID `gorm:"type:varchar(36);not null"`
	CollectionIndex int64     `gorm:"not null"`
	Longtitude      float64   `gorm:"not null"`
	Latitude        float64   `gorm:"not null"`
	Altitude        float64   `gorm:"not null"`
	Accuracy        float64   `gorm:"not null"`
	OriginAt        time.Time `gorm:"not null"`
	CreatedAt       time.Time `gorm:"autoCreateTime"`
}
