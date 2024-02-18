package auth

import (
	"time"

	"github.com/capdale/was/model"
	"github.com/capdale/was/types/binaryuuid"
	"github.com/google/uuid"
)

type database interface {
	GetUserIdByUUID(userUUID binaryuuid.UUID) (int64, error)
	SaveToken(userId int64, tokenString string, refreshToken *[]byte, agent *string) error
	IfTokenExistRemoveElseErr(tokenString string, until time.Duration, blackToken func(string, time.Duration) error) error
	PopTokenByRefreshToken(refreshToken *[]byte, transactionF func(string) error) (tokenString *string, err error)
	QueryAllTokensByUserUUID(userUUID *uuid.UUID) (*[]*model.Token, error)
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
