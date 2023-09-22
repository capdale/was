package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthPayload struct {
	UserUUID uuid.UUID
	Username string
}

type AuthClaims struct {
	Username     string    `json:"username"`
	UserUUID     uuid.UUID `json:"uuid"`
	RefreshToken string    `json:"refresh_token"`
	jwt.RegisteredClaims
}

var (
	ErrTokenInvalid   = errors.New("token is invalid")
	ErrTypeParse      = errors.New("fail parse custom claim")
	ErrTokenBlacklist = errors.New("token is blacklist")
)

func (a *AuthClaims) isExpired() bool {
	return time.Until(a.ExpiresAt.Time) < 0
}

func (a *Auth) GenerateToken(dp *AuthPayload) (string, error) {
	refreshToken, err := createRandomToken()
	if err != nil {
		return "", err
	}

	expirationTime := time.Now().Add(time.Hour)
	claims := &AuthClaims{
		Username:     dp.Username,
		UserUUID:     dp.UserUUID,
		RefreshToken: refreshToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(a.secret)
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func (a *Auth) ParseToken(tokenString string) (claims *AuthClaims, err error) {
	claims = &AuthClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) { return a.secret, nil })
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if !token.Valid {
		return nil, ErrTokenInvalid
	}

	if claims.isExpired() {
		return claims, jwt.ErrTokenExpired
	}

	return
}

func (a *Auth) ValidateToken(tokenString string) (*AuthClaims, error) {
	claims, err := a.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	isBlacklist, err := a.IsBlacklist(tokenString)
	if err != nil {
		return nil, err
	}
	if isBlacklist {
		return nil, ErrTokenBlacklist
	}

	return claims, nil
}
