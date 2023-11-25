package collect

import (
	"net/http"
	"os"
	"path"
	"time"

	"github.com/capdale/was/auth"
	"github.com/capdale/was/location"
	"github.com/capdale/was/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type database interface {
	PutInImageQueue(userUUID *uuid.UUID, fileUUID *uuid.UUID, geoLocation *location.GeoLocation) error
	GetCollectection(userUUID *uuid.UUID, offset int, limit int) (*[]model.Collection, error)
}

type CollectAPI struct {
	DB database
}

func New(database database) *CollectAPI {
	return &CollectAPI{
		DB: database,
	}
}

type GetCollectionReq struct {
	Offset *int `json:"offset" form:"offset" binding:"required"`
	Limit  *int `json:"limit" form:"limit" binding:"required"`
}

type GetCollectionRes struct {
	Collections *[]*Collections `json:"collections"`
}

type Collections struct {
	UUID            uuid.UUID `json:"uuid"`
	CollectionIndex int64     `json:"index"`
	Longtitude      float64   `json:"long"`
	Latitude        float64   `json:"lat"`
	Altitude        float64   `json:"alt"`
	OriginAt        time.Time `json:"origin_at"`
}

func (a *CollectAPI) GetCollectection(ctx *gin.Context) {
	var req GetCollectionReq
	err := ctx.ShouldBind(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	claims := ctx.MustGet("claims").(*auth.AuthClaims)
	collections, err := a.DB.GetCollectection(&claims.UserUUID, *req.Offset, *req.Limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	res := &GetCollectionRes{
		Collections: &[]*Collections{},
	}
	for _, collection := range *collections {
		(*res.Collections) = append((*res.Collections), &Collections{
			UUID:            collection.UUID,
			CollectionIndex: collection.CollectionIndex,
			Longtitude:      collection.Longtitude,
			Latitude:        collection.Latitude,
			Altitude:        collection.Altitude,
			OriginAt:        collection.OriginAt,
		})
	}
	ctx.JSON(http.StatusOK, res)
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
