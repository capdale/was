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

type Token struct {
	// this token is same as jwt token, write when token generated, delete when token blacklist, query when refresh request comes in
	UUID      uuid.UUID `gorm:"type:varchar(36);uniqueIndex;not null"`
	Token     string    `gorm:"type:varchar(100);not null"` // need type tuning
	ExpireAt  time.Time `gorm:"index;,sort:desc"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
