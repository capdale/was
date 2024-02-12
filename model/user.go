package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id          int64 `gorm:"primaryKey"`
	Username    string
	AccountType int
	UUID        uuid.UUID `gorm:"type:varchar(36);uniqueIndex:uuid;not null"`
	Email       string    `gorm:"size:64;uniqueIndex;not null"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdateAt    time.Time `gorm:"autoUpdateTime"`
}

type Token struct {
	// this token is same as jwt token, write when token generated, delete when token blacklist, query when refresh request comes in
	ID           []byte    `gorm:"type:binary(16);primaryKey"`
	UserUUID     uuid.UUID `gorm:"type:varchar(36);index:user_uuid;not null"`
	Token        string    `gorm:"type:varchar(225);not null"` // need type tuning
	RefreshToken []byte    `gorm:"type:binary(64);index:refresh_token"`
	UserAgent    string    `gorm:"type:varchar(225)"`
	ExpireAt     time.Time
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}
