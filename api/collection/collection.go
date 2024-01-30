package collect

import (
	"mime/multipart"
	"net/http"

	"github.com/capdale/was/auth"
	"github.com/capdale/was/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type database interface {
	GetCollections(userUUID *uuid.UUID, offset int, limit int) (*[]Collection, error)
	GetCollectionByUUID(collectionUUID *uuid.UUID) (*Collection, error)
	CreateCollectionWithUserUUID(collection *Collection, userUUID *uuid.UUID) (collectionUUID *uuid.UUID, err error)
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
	Collections []Collection `json:"collections"`
}

type PostCollectionReq struct {
	ImageFile *multipart.FileHeader `form:"image" binding:"required"`
	Info      Collection            `form:"info" binding:"required"`
}
type GetCollectionByUUIDReq struct {
	CollectionUUID *uuid.UUID `json:"uuid" binding:"required"`
}

type Collection = model.CollectionAPI

func (a *CollectAPI) GetCollectection(ctx *gin.Context) {
	var req GetCollectionReq
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	claims := ctx.MustGet("claims").(*auth.AuthClaims)
	collections, err := a.DB.GetCollections(&claims.UserUUID, *req.Offset, *req.Limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	res := &GetCollectionRes{
		Collections: *collections,
	}
	ctx.JSON(http.StatusOK, res)
}

func (a *CollectAPI) PostCollection(ctx *gin.Context) {
	claims := ctx.MustGet("claims").(*auth.AuthClaims)
	var body PostCollectionReq

	if err := ctx.ShouldBind(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "bad request form",
		})
		return
	}

	err := isValidImageFromFile(body.ImageFile)
	if err != nil {
		if err == ErrImageInValid {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "invalid image",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "internal server error",
		})
		return
	}

	issuedUUID, err := a.DB.CreateCollectionWithUserUUID(&body.Info, &claims.UserUUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "internal server error",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"uuid":    issuedUUID,
	})
}

func (a *CollectAPI) GetCollectionByUUID(ctx *gin.Context) {
	var req GetCollectionByUUIDReq
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	collection, err := a.DB.GetCollectionByUUID(req.CollectionUUID)
	if err != nil {
		ctx.JSON(http.StatusNoContent, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, collection)
}
