package auth

import "github.com/gin-gonic/gin"

type database interface {
}

type store interface {
	IsBlacklist(token string) (bool, error)
	SetBlacklist(token string) error
}

type Auth struct {
	DB     database
	Store  store
	secret []byte
}

func (a *Auth) New(database database, store store) *Auth {
	return &Auth{
		DB:    database,
		Store: store,
	}
}

func (a *Auth) tokenFromCTX(ctx *gin.Context) string {
	if token := a.tokenFromQuery(ctx); token != "" {
		return token
	} else if token := a.tokenFromXToken(ctx); token != "" {
		return token
	}
	return ""
}

func (a *Auth) tokenFromQuery(ctx *gin.Context) string {
	return ctx.Request.URL.Query().Get("token")
}

func (a *Auth) tokenFromXToken(ctx *gin.Context) string {
	return ctx.Request.Header.Get("X-MD-Token")
}
