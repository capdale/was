package model

import (
	"time"

	"github.com/capdale/was/types/binaryuuid"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	Id          int64           `gorm:"primaryKey"`
	Username    string          `gorm:"type:varchar(36);uniqueIndex:username;not null"`
	UUID        binaryuuid.UUID `gorm:"uniqueIndex;"`
	AccountType int
	Email       string       `gorm:"size:64;uniqueIndex;not null"`
	CreatedAt   time.Time    `gorm:"autoCreateTime"`
	UpdateAt    time.Time    `gorm:"autoUpdateTime"`
	Collections []Collection `gorm:"foreignkey:UserId;references:Id"`
}

type Token struct {
	// this token is same as jwt token, write when token generated, delete when token blacklist, query when refresh request comes in
	ID           []byte `gorm:"type:binary(16);primaryKey"`
	UserId       int64  `gorm:"index;not null"`
	Token        string `gorm:"type:varchar(225);not null"` // need type tuning
	RefreshToken []byte `gorm:"type:binary(64);index:refresh_token"`
	UserAgent    string `gorm:"type:varchar(225)"`
	ExpireAt     time.Time
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	uid, err := uuid.NewRandom()
	u.UUID = binaryuuid.UUID(uid)
	return err
}
