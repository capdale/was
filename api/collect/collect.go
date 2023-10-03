package collect

import (
	"net/http"
	"os"
	"path"

	"github.com/capdale/was/auth"
	"github.com/capdale/was/location"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type database interface {
	PutInImageQueue(userUUID *uuid.UUID, fileUUID *uuid.UUID, geoLocation *location.GeoLocation) error
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

	fileUUID, err := a.saveImage(b)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	claims := ctx.MustGet("claims").(*auth.AuthClaims)
	err = a.DB.PutInImageQueue(&claims.UserUUID, fileUUID, geoLocation)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"ok": "ok",
	})
}

func (a *CollectAPI) saveImage(imageBytes *[]byte) (*uuid.UUID, error) {
	fileUUID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path.Join("./secret", fileUUID.String()), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	_, err = f.Write(*imageBytes)
	if err != nil {
		return nil, err
	}
	return &fileUUID, nil
}
