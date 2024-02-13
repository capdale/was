package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func BasicInternalServerError(ctx *gin.Context) {
	ctx.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
}

func BasicUnAuthorizedError(ctx *gin.Context) {
	ctx.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized error"})
}

func BasicBadRequestError(ctx *gin.Context) {
	ctx.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
}
