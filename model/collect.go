package model

import (
	"time"

	"github.com/google/uuid"
)

type Collection struct {
	UserUUID        uuid.UUID `gorm:"type:varchar(36);not null"`
	UUID            uuid.UUID `gorm:"type:varchar(36);uniqueIndex;not null"`
	CollectionIndex int64     `gorm:"not null"`
	Longtitude      float64   `gorm:"not null"`
	Latitude        float64   `gorm:"not null"`
	Altitude        float64   `gorm:"not null"`
	Accuracy        float64   `gorm:"not null"`
	OriginAt        time.Time `gorm:"autoCreateTime"`
}

type CollectionAPI struct {
	UUID            *uuid.UUID `json:"uuid,omitempty"`
	CollectionIndex int64      `json:"index" binding:"required"`
	Longtitude      float64    `json:"long" binding:"required"`
	Latitude        float64    `json:"lat" binding:"required"`
	Altitude        float64    `json:"alt" binding:"required"`
	Accuracy        float64    `json:"acc" binding:"required"`
	OriginAt        time.Time  `json:"origin_at,omitempty"`
}
