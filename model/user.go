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
	Id              int64           `gorm:"primaryKey"`
	Username        string          `gorm:"type:varchar(36);uniqueIndex:username;not null"`
	UUID            binaryuuid.UUID `gorm:"uniqueIndex;"`
	AccountType     int
	Email           string           `gorm:"size:64;uniqueIndex;not null"`
	CreatedAt       time.Time        `gorm:"autoCreateTime"`
	UpdateAt        time.Time        `gorm:"autoUpdateTime"`
	Collections     []Collection     `gorm:"foreignkey:UserId;references:Id"`
	OriginUser      OriginUser       `gorm:"foreignkey:Id;references:Id;constraint:OnDelete:CASCADE"`
	SocialUser      SocialUser       `gorm:"foreignkey:Id;references:Id;constraint:OnDelete:CASCADE"`
	UserDisplayType *UserDisplayType `gorm:"foreignKey:UserId;references:Id;constraint:OnDelete:CASCADE"`
	Tokens          *[]*Token        `gorm:"foreignKey:UserId;references:Id;constraint:OnUpdate:SET NULL,OnDelete:CASCADE"`
	UserFollowers   *[]*UserFollow   `gorm:"foreginKey:UserId;references:Id;constraint:OnDelete:CASCADE"`
	UserFollowings  *[]*UserFollow   `gorm:"foreginKey:TargetId;references:Id;constraint:OnDelete:CASCADE"`
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
	Id           int64           `gorm:"primaryKey"`
	UserId       int64           `gorm:"index;"`
	UUID         binaryuuid.UUID `gorm:"index"` // token uuid to identify token
	RefreshToken []byte          `gorm:"size:60;"`
	UserAgent    string          `gorm:"type:varchar(225)"`
	NotBefore    time.Time       // jwt expired at, refresh token cannot be used before this, also used when make jwt token
	ExpireAt     time.Time       // refresh token expired at, after can't refresh with this
	CreatedAt    time.Time       `gorm:"autoCreateTime"`
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

func (t *Ticket) BeforeCreate(tx *gorm.DB) error {
	uid, err := uuid.NewRandom()
	t.UUID = binaryuuid.UUID(uid)
	return err
}

type UserDisplayType struct {
	UserId    int64 `gorm:"index"`
	IsPrivate bool
}

type UserFollow struct {
	UserId   int64 `gorm:"index:user_idx;uniqueIndex:user_target_idx"`
	TargetId int64 `gorm:"index:target_idx;uniqueIndex:user_target_idx"`
}

type UserFollowRequest struct {
	UserId   int64 `gorm:"index:user_idx;uniqueIndex:user_target_idx"`
	TargetId int64 `gorm:"index:target_idx;uniqueIndex:user_target_idx"`
}
