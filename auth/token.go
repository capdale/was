package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthClaims struct {
	UserUUID uuid.UUID `json:"uuid"`
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

func (a *Auth) GenerateClaim(userUUID *uuid.UUID) *AuthClaims {
	expirationTime := time.Now().Add(time.Hour)
	return &AuthClaims{
		UserUUID: *userUUID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
}

func (a *Auth) GenerateToken(claims *AuthClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(a.secret)
	if err != nil {
		return "", err
	}
	return signedToken[37:], nil
}

func (a *Auth) ParseToken(tokenString string) (claims *AuthClaims, err error) {
	claims = &AuthClaims{}
	token, err := jwt.ParseWithClaims("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9."+tokenString, claims, func(t *jwt.Token) (interface{}, error) { return a.secret, nil })
	if err != nil {
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
