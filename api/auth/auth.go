package authapi

import (
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"github.com/capdale/was/api"
	"github.com/capdale/was/auth"
	"github.com/capdale/was/logger"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type database interface {
}

type AuthAPI struct {
	DB   database
	Auth *auth.Auth
}

type RefreshTokenReq struct {
	RefreshToken *string `json:"refresh_token" binding:"required"`
}

var (
	ErrStateNotEqual    = errors.New("state is not equal")
	ErrNoValidEmail     = errors.New("no valid email")
	ErrAlreayExistEmail = errors.New("already exist email")
)

var (
	AccessDenied = gin.H{
		"message": "access denied",
	}
)

func New(database database, auth *auth.Auth) *AuthAPI {
	return &AuthAPI{
		DB:   database,
		Auth: auth,
	}
}

func CheckState(ctx *gin.Context) error {
	session := sessions.Default(ctx)
	state := session.Get("state")
	if state == nil {
		return errors.New("state not found")
	}
	session.Clear()
	return nil
}

func (a *AuthAPI) SetBlacklistHandler(ctx *gin.Context) {
	claims := ctx.MustGet("claims").(*auth.AuthClaims)
	tokenString := ctx.MustGet("token").(string)
	a.Auth.SetBlacklistByToken(claims)
	err := a.Auth.Store.SetBlacklist(tokenString, time.Until(claims.ExpiresAt.Time))
	if err != nil {
		api.BasicInternalServerError(ctx)
		logger.Logger.ErrorWithCTX(ctx, "set token blacklist", err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

func (a *AuthAPI) RefreshTokenHandler(ctx *gin.Context) {
	form := new(RefreshTokenReq)
	err := ctx.BindJSON(form)
	if err != nil {
		api.BasicBadRequestError(ctx)
		logger.Logger.ErrorWithCTX(ctx, "binding form", err)
		return
	}

	userAgent := ctx.Request.UserAgent()
	oldRefreshToken, err := base64.StdEncoding.DecodeString(*form.RefreshToken)
	if err != nil {
		api.BasicBadRequestError(ctx)
		logger.Logger.ErrorWithCTX(ctx, "binding refresh token", err)
		return
	}

	newToken, newRefreshToken, err := a.Auth.RefreshToken(&oldRefreshToken, &userAgent)
	if err != nil {
		api.BasicInternalServerError(ctx)
		logger.Logger.ErrorWithCTX(ctx, "refresh token failed", err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  newToken,
		"refresh_token": base64.StdEncoding.EncodeToString(*newRefreshToken),
	})
}

func (a *AuthAPI) WhoAmIHandler(ctx *gin.Context) {
	claims := ctx.MustGet("claims").(*auth.AuthClaims)
	ctx.JSON(http.StatusOK, gin.H{"uuid": claims.UUID.String()})
}
