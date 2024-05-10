package auth

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/capdale/was/model"
	"github.com/capdale/was/types/binaryuuid"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
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

func (a *AuthClaims) IsExpired() bool {
	return a.ExpiresAt.Time.Before(time.Now())
}

func (a *Auth) IssueToken(userAuthUUID binaryuuid.UUID, agent *string) (tokenString string, refreshTokenString string, err error) {
	// this function manage all secure process, store refresh token in db, validate token etc
	expireAt := time.Now().Add(time.Minute * 30)
	claims, err := a.generateClaim(userAuthUUID, expireAt)
	if err != nil {
		return
	}

	tokenString, err = a.generateToken(claims)
	if err != nil {
		return
	}

	refreshTokenUID, refreshToken, err := a.generateRefreshToken()
	if err != nil {
		return
	}

	userId, err := a.DB.GetUserIdByAuthUUID(userAuthUUID)
	if err != nil {
		return
	}

	refreshTokenExpireAt := time.Now().Add(time.Hour * 24 * 7)
	if err = a.DB.CreateRefreshToken(userId, refreshTokenUID, refreshToken, claims.ExpiresAt.Time, refreshTokenExpireAt, agent); err != nil {
		return
	}

	refreshTokenUIDStr := base64.URLEncoding.EncodeToString((*refreshTokenUID)[:])
	refrehTokenStr := base64.URLEncoding.EncodeToString(*refreshToken)
	refreshTokenString = fmt.Sprintf("%s.%s", refreshTokenUIDStr, refrehTokenStr)
	return
}

func (a *Auth) generateClaim(userUUID binaryuuid.UUID, expireAt time.Time) (c *AuthClaims, err error) {
	c = &AuthClaims{
		UUID: userUUID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireAt),
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

func (a *Auth) ParseTokenIgnoreExpired(tokenString *string) (*AuthClaims, error) {
	claims, err := a.ParseToken(*tokenString)
	if err != nil {
		if errors.Is(jwt.ErrTokenExpired, err) && !errors.Is(jwt.ErrTokenMalformed, err) {
			return claims, err
		} else {
			return nil, err
		}
	}
	return claims, err
}

func (a *Auth) RefreshToken(refreshTokenString string, agent *string) (newTokenString string, newRefreshTokenString string, err error) {
	parts := strings.Split(refreshTokenString, ".")
	if len(parts) != 2 {
		err = ErrTokenInvalid
		return
	}
	refreshTokenUIDBytes, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		return
	}
	refreshTokenUID, err := binaryuuid.FromBytes(refreshTokenUIDBytes)
	if err != nil {
		return
	}
	refreshTokenBytes, err := base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		return
	}

	refreshToken, err := a.popRefreshToken(&refreshTokenUID, &refreshTokenBytes)
	if err != nil {
		return
	}

	if err = a.IsRefreshTokenValid(refreshToken); err != nil {
		// if errors.Is(ErrTokenNotExpiredYet, err) {
		// SECURITY TODO: if err token not expired yet, then need to expire access token too?
		// }
		return
	}

	user, err := a.DB.GetUserById(refreshToken.UserId)
	if err != nil {
		return
	}

	newTokenString, newRefreshTokenString, err = a.IssueToken(user.UUID, agent)
	return
}

func (a *Auth) popRefreshToken(refreshTokenUID *binaryuuid.UUID, refreshToken *[]byte) (*model.Token, error) {
	token, err := a.DB.PopRefreshToken(refreshTokenUID)
	if err != nil {
		return nil, err
	}

	if err = bcrypt.CompareHashAndPassword(token.RefreshToken, *refreshToken); err != nil {
		return nil, err
	}
	return token, nil
}

func (a *Auth) IsRefreshTokenValid(token *model.Token) error {
	if token.NotBefore.After(time.Now()) { //token is not expired yet
		return ErrTokenNotExpiredYet
	}
	if token.ExpireAt.Before(time.Now()) { // refresh token is expired1
		return errors.New("refresh token expired")
	}
	return nil
}

// func (a *Auth) generateClaimsFromRefreshToken(token *model.Token) (*AuthClaims, error) {
// 	user, err := a.DB.GetUserById(token.UserId)
// 	if err != nil {
// 		return nil, err
// 	}
// 	claims, err := a.generateClaim(user.UUID, token.ExpireAt)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return claims, err
// }

// func (a *Auth) refreshTransaction(tokenString string) (err error) {
// 	claims, err := a.ParseTokenIgnoreExpired(&tokenString)
// 	if err != nil {
// 		return
// 	}
// 	if !claims.IsExpired() {
// 		err = a.Store.SetBlacklist(tokenString, time.Until(claims.ExpiresAt.Time))
// 		if err != nil {
// 			return
// 		}
// 	}
// 	return
// }

// func (a *Auth) ActiveTokensByUserUUID(userUUID *uuid.UUID) (*[]string, error) {
// 	return a.DB.QueryAllTokensByUserUUID(userUUID)
// }
