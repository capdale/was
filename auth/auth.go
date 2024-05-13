package auth

import (
	"time"

	"github.com/capdale/was/model"
	"github.com/capdale/was/types/binaryuuid"
	"github.com/capdale/was/types/claimer"
)

type database interface {
	CreateRefreshToken(claimer claimer.Claimer, tokenUID *binaryuuid.UUID, refreshToken *[]byte, notBefore time.Time, expireAt time.Time, agent *string) error
	IsTokenPair(claimer claimer.Claimer, tokenExpiredAt time.Time, refreshToken *[]byte) error
	PopRefreshToken(refreshTokenUID *binaryuuid.UUID) (*model.Token, error)
	GetUserClaimByID(claimerId uint64) (*claimer.Claimer, error)
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
