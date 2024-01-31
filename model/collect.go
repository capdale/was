package model

import (
	"time"

	"github.com/google/uuid"
)

type Collection struct {
	UserUUID        uuid.UUID   `gorm:"type:varchar(36);index:user_uuid;not null"`
	UUID            uuid.UUID   `gorm:"type:varchar(36);uniqueIndex:uuid;not null"`
	CollectionIndex int64       `gorm:"not null"`
	Geolocation     Geolocation `gorm:"embedded;not null"`
	Accuracy        float64     `gorm:"not null"`
	OriginAt        time.Time   `gorm:"autoCreateTime"`
}

type Geolocation struct {
	Longtitude float64 `json:"longtitude" binding:"required"`
	Latitude   float64 `json:"latitude" binding:"required"`
	Altitude   float64 `json:"altitude" binding:"required"`
	Accuracy   float64 `json:"acc" binding:"required"`
}

type CollectionAPI struct {
	CollectionIndex int64       `json:"index" binding:"required"`
	Geolocation     Geolocation `json:"geolocation" gorm:"embedded;not null"`
	OriginAt        *time.Time  `json:"datetime,omitempty"`
}
