package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *Auth) AuthorizeRequiredMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString := a.tokenFromCTX(ctx)
		if tokenString == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "no x-token",
			})
			return
		}

		claims, err := a.ValidateToken(tokenString)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"message": "invalid token",
			})
			return
		}

		ctx.Set("token", tokenString)
		ctx.Set("claims", claims)
		ctx.Next()
	}
}
