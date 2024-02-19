package model

import (
	"time"

	"github.com/capdale/was/types/binaryuuid"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	AccountTypeOrigin = 0
	AccountTypeGithub = 1
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
	OriginUser  OriginUser   `gorm:"foreignkey:Id;references:Id"`
	SocialUser  SocialUser   `gorm:"foreignkey:Id;references:Id"`
}

type OriginUser struct {
	Id     int64  `gorm:"index"`
	Hashed []byte `gorm:"size:60;not null"`
}

type SocialUser struct {
	Id          int64
	AccountType int
}

type Token struct {
	// this token is same as jwt token, write when token generated, delete when token blacklist, query when refresh request comes in
	IssuerUUID   binaryuuid.UUID `gorm:"index;not null"`
	Token        string          `gorm:"type:varchar(225);not null"` // need type tuning
	RefreshToken []byte          `gorm:"type:binary(64);index:refresh_token"`
	UserAgent    string          `gorm:"type:varchar(225)"`
	ExpireAt     time.Time
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	uid, err := uuid.NewRandom()
	u.UUID = binaryuuid.UUID(uid)
	return err
}

type Ticket struct {
	Email     string          `gorm:"size:64;not null"`
	UUID      binaryuuid.UUID `gorm:"uniqueIndex;not null"`
	CreatedAt time.Time       `gorm:"autoCreateTime"`
}
