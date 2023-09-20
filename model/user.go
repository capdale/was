package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Username    string
	AccountType int
	UUID        uuid.UUID `gorm:"type:varchar(36);uniqueIndex;not null"`
	Email       string    `gorm:"size:64;uniqueIndex;not null"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdateAt    time.Time `gorm:"autoUpdateTime"`
}
