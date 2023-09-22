package collect

import (
	"fmt"
	"net/http"

	"github.com/capdale/was/location"
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
	geoLocation, err := location.GeoLocationFromCTX(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "no geolocation data",
		})
		return
	}

	b, err := getByteFromCTX(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "no b data",
		})
		return
	}

	if !isValidImage(b) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid data",
		})
		return
	}

	fmt.Println(*geoLocation)

	ctx.JSON(http.StatusOK, gin.H{
		"ok": "ok",
	})
}
