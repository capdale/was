package model

import (
	"time"

	"github.com/capdale/was/types/binaryuuid"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Collection struct {
	Id              uint64          `gorm:"primaryKey"`
	UserId          int64           `gorm:"index:id"`
	UUID            binaryuuid.UUID `gorm:"uniqueIndex:uuid;not null"`
	CollectionIndex int64           `gorm:"not null"`
	Geolocation     Geolocation     `gorm:"embedded;not null"`
	Accuracy        float64         `gorm:"not null"`
	OriginAt        time.Time       `gorm:"autoCreateTime"`
	DeletedAt       gorm.DeletedAt
}

func (c *Collection) BeforeCreate(tx *gorm.DB) error {
	uid, err := uuid.NewRandom()
	c.UUID = binaryuuid.UUID(uid)
	return err
}

type Geolocation struct {
	Longtitude float64 `json:"longtitude" binding:"required"`
	Latitude   float64 `json:"latitude" binding:"required"`
	Altitude   float64 `json:"altitude" binding:"required"`
	Accuracy   float64 `json:"acc" binding:"required"`
}

type CollectionUID struct {
	UUID binaryuuid.UUID `json:"uuid"`
}

type CollectionAPI struct {
	CollectionIndex int64       `json:"index" binding:"required"`
	Geolocation     Geolocation `json:"geolocation" gorm:"embedded"`
	OriginAt        *time.Time  `json:"datetime,omitempty"`
}
