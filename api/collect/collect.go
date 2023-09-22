package collect

import (
	"fmt"
	"net/http"

	"github.com/capdale/was/location"
	rpc_protocol "github.com/capdale/was/rpc/proto"
	"github.com/gin-gonic/gin"
)

type database interface {
}

type CollectAPI struct {
	DB                   database
	ImageClassifiyClient *rpc_protocol.ImageClassifyClient
}

func New(database database, imageClassifyClient *rpc_protocol.ImageClassifyClient) *CollectAPI {
	return &CollectAPI{
		DB:                   database,
		ImageClassifiyClient: imageClassifyClient,
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

	reply, err := (*a.ImageClassifiyClient).ClassifyImage(ctx, &rpc_protocol.ImageClassifierRequest{Image: *b})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"ok":    "ok",
		"class": reply.ClassIndex,
	})
}
