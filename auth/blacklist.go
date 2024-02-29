package auth

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

var ErrInValidRequest = errors.New("in valid request cannot find token")

func (a *Auth) IsBlacklist(token string) (bool, error) {
	return a.Store.IsBlacklist(token)
}

func (a *Auth) BlackToken(tokenString *string, refreshTokenString *string) error {
	claims, err := a.ValidateToken(*tokenString)
	if err != nil {
		return err
	}
	refreshToken, err := base64.StdEncoding.DecodeString(*refreshTokenString)
	if err != nil {
		return err
	}
	userId, err := a.DB.GetUserIdByUUID(claims.UUID)
	if err != nil {
		return err
	}

	if err := a.DB.IsTokenPair(userId, claims.ExpiresAt.Time, &refreshToken); err != nil {
		return err
	}
	return nil
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
