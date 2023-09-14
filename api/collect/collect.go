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

func (a *CollectAPI) PostCollectHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"ok": "ok",
	})
}
