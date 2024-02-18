package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/capdale/was/types/binaryuuid"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthClaims struct {
	UUID binaryuuid.UUID `json:"user"`
	jwt.RegisteredClaims
}

var (
	ErrTokenInvalid       = errors.New("token is invalid")
	ErrTypeParse          = errors.New("fail parse custom claim")
	ErrTokenBlacklist     = errors.New("token is blacklist")
	ErrTokenNotExpiredYet = errors.New("token not expired yet")
)

func (a *AuthClaims) isExpired() bool {
	return time.Until(a.ExpiresAt.Time) < 0
}

func (a *Auth) IssueTokenByUserUUID(userUUID binaryuuid.UUID, userId int64, agent *string) (tokenString string, refreshToken *[]byte, err error) {
	// this function manage all secure process, store refresh token in db, validate token etc
	claims, err := a.generateClaim(userUUID)
	if err != nil {
		return
	}
	tokenString, err = a.generateToken(claims)
	if err != nil {
		return
	}
	refreshToken, err = a.generateRefreshToken()
	if err != nil {
		return
	}
	if err = a.DB.SaveToken(userId, tokenString, refreshToken, agent); err != nil {
		return
	}
	return
}

func (a *Auth) generateRefreshToken() (*[]byte, error) {
	refreshToken := make([]byte, 64)
	randomUUID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	randBackBytes, err := randomUUID.MarshalBinary()
	if err != nil {
		return nil, err
	}
	randFrontBytes, err := RandToken(48)
	if err != nil {
		return nil, err
	}
	refreshToken = append(*randFrontBytes, randBackBytes...)
	fmt.Println(len(refreshToken))
	return &refreshToken, nil
}

func (a *Auth) generateClaim(userUUID binaryuuid.UUID) (c *AuthClaims, err error) {
	expirationTime := time.Now().Add(time.Minute * 30)
	c = &AuthClaims{
		UUID: userUUID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	return
}

func (a *Auth) generateToken(claims *AuthClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(a.secret)
	if err != nil {
		return "", err
	}
	return signedToken[37:], nil
}

func (a *Auth) ParseToken(tokenString string) (claims *AuthClaims, err error) {
	claims = &AuthClaims{}
	_, err = jwt.ParseWithClaims("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9."+tokenString, claims, func(t *jwt.Token) (interface{}, error) { return a.secret, nil })
	return
}

func (a *Auth) ParseTokenIgnoreExpired(tokenString string) (claims *AuthClaims, err error) {
	claims, err = a.ParseToken(tokenString)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) && claims.isExpired() {
			return claims, nil
		}
		return
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

func (a *Auth) RefreshToken(refreshToken *[]byte, agent *string) (newToken string, newRefreshToken *[]byte, err error) {
	tokenString, err := a.DB.PopTokenByRefreshToken(refreshToken, a.refreshTransaction)
	if err != nil {
		return
	}
	claims, err := a.ParseTokenIgnoreExpired(*tokenString)
	if err != nil {
		return
	}
	userId, err := a.DB.GetUserIdByUUID(claims.UUID)
	if err != nil {
		return
	}
	newToken, newRefreshToken, err = a.IssueTokenByUserUUID(claims.UUID, userId, agent)
	return
}

func (a *Auth) refreshTransaction(tokenString string) (err error) {
	claims, err := a.ParseTokenIgnoreExpired(tokenString)
	if err != nil {
		return
	}
	if !claims.isExpired() {
		err = a.Store.SetBlacklist(tokenString, time.Until(claims.ExpiresAt.Time))
		if err != nil {
			return
		}
		return ErrTokenNotExpiredYet
	}
	return
}

// func (a *Auth) ActiveTokensByUserUUID(userUUID *uuid.UUID) (*[]string, error) {
// 	return a.DB.QueryAllTokensByUserUUID(userUUID)
// }
