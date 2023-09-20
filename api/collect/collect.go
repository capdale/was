package collect

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type database interface {
}

type CollectAPI struct {
	DB database
}

func New(database database) *CollectAPI {
	return &CollectAPI{
		DB: database,
	}
}

func (a *CollectAPI) GetCollectection(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"collection": "",
	})
}

func (a *CollectAPI) PostCollectHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"ok": "ok",
	})
}
