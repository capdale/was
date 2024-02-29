package auth

import (
	"time"

	"github.com/capdale/was/model"
	"github.com/capdale/was/types/binaryuuid"
)

type database interface {
	CreateRefreshToken(userId int64, tokenUID *binaryuuid.UUID, refreshToken *[]byte, notBefore time.Time, expireAt time.Time, agent *string) error
	IsTokenPair(userId int64, tokenExpiredAt time.Time, refreshToken *[]byte) error
	PopRefreshToken(refreshTokenUID *binaryuuid.UUID) (*model.Token, error)
	GetUserById(userId int64) (user *model.User, err error)
	GetUserIdByUUID(userUUID binaryuuid.UUID) (int64, error)
}

type store interface {
	IsBlacklist(token string) (bool, error)
	SetBlacklist(token string, expiration time.Duration) error
}

type Auth struct {
	DB     database
	Store  store
	secret []byte
}

func New(database database, store store) *Auth {
	return &Auth{
		DB:    database,
		Store: store,
	}
}
