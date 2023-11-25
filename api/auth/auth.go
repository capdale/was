package authapi

import (
	"errors"
	"net/http"
	"time"

	"github.com/capdale/was/auth"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type database interface {
}

type AuthAPI struct {
	DB   database
	Auth *auth.Auth
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
	err := a.Auth.Store.SetBlacklist(tokenString, time.Until(claims.ExpiresAt.Time))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "internal server error",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}
