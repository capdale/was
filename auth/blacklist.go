package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrInValidRequest = errors.New("in valid request cannot find token")

func (a *Auth) IsBlacklist(token string) (bool, error) {
	return a.Store.IsBlacklist(token)
}

func (a *Auth) SetBlacklistByToken(claims *AuthClaims) error {
	tokenString, err := a.generateToken(claims)
	if err != nil {
		return err
	}
	return a.DB.IfTokenExistRemoveElseErr(tokenString, time.Until(claims.ExpiresAt.Time), a.Store.SetBlacklist)
}

func (a *Auth) SetBlacklistByUserUUID(userUUID *uuid.UUID) error {
	return nil
}

func TokenFromRequest(req *http.Request) (string, error) {
	authString := req.Header.Get("Authorization")
	authStruct := strings.Split(authString, " ")
	if len(authStruct) != 2 {
		return "", ErrInValidRequest
	}
	authScheme, authParam := authStruct[0], authStruct[1]
	if authScheme != "Bearer" {
		return "", ErrInValidRequest
	}
	return authParam, nil
}
