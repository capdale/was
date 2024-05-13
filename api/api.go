package api

import (
	"net/http"

	"github.com/capdale/was/types/claimer"
	"github.com/gin-gonic/gin"
)

func GetClaimer(ctx *gin.Context) *claimer.Claimer {
	claimerPtr, ok := ctx.Get("claimer")
	if !ok {
		return nil
	}
	claimer := claimerPtr.(*claimer.Claimer)
	return claimer
}

func MustGetClaimer(ctx *gin.Context) *claimer.Claimer {
	claimer := ctx.MustGet("claimer").(*claimer.Claimer)
	return claimer
}

func BasicInternalServerError(ctx *gin.Context) {
	ctx.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
}

func BasicUnAuthorizedError(ctx *gin.Context) {
	ctx.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized error"})
}

func BasicBadRequestError(ctx *gin.Context) {
	ctx.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
}
