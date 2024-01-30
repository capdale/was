package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *Auth) AuthorizeRequiredMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString, err := TokenFromRequest(ctx.Request)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "invalid request",
			})
			return
		}

		claims, err := a.ValidateToken(tokenString)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "invalid token",
			})
			return
		}

		ctx.Set("token", tokenString)
		ctx.Set("claims", claims)
		ctx.Next()
	}
}

func (a *Auth) AuthorizeOptionalMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString, err := TokenFromRequest(ctx.Request)
		if err != nil || tokenString == "" {
			ctx.Next()
			return
		}

		claims, err := a.ValidateToken(tokenString)
		if err != nil {
			ctx.Next() // consider
			return
		}

		ctx.Set("token", tokenString)
		ctx.Set("claims", claims)
		ctx.Next()
	}
}
