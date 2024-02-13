package collect

import (
	"mime/multipart"
	"net/http"

	"github.com/capdale/was/api"
	"github.com/capdale/was/auth"
	baseLogger "github.com/capdale/was/logger"
	"github.com/capdale/was/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var logger = baseLogger.Logger

type database interface {
	GetCollections(userUUID *uuid.UUID, offset int, limit int) (*[]uuid.UUID, error)
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
	Collections []uuid.UUID `json:"collections"`
}

type PostCollectionReq struct {
	ImageFile *multipart.FileHeader `form:"image" binding:"required"`
	Info      Collection            `form:"info" binding:"required"`
}

type Collection = model.CollectionAPI

func (a *CollectAPI) GetCollectection(ctx *gin.Context) {
	var req GetCollectionReq
	if err := ctx.ShouldBind(&req); err != nil {
		logger.ErrorWithCTX(ctx, "binding form", err)
		api.BasicBadRequestError(ctx)
		return
	}

	claims := ctx.MustGet("claims").(*auth.AuthClaims)
	collections, err := a.DB.GetCollections(&claims.UserUUID, *req.Offset, *req.Limit)
	if err != nil {
		logger.ErrorWithCTX(ctx, "db get collections", err)
		api.BasicInternalServerError(ctx)
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
		logger.ErrorWithCTX(ctx, "binding form", err)
		api.BasicBadRequestError(ctx)
		return
	}

	err := isValidImageFromFile(body.ImageFile)
	if err != nil {
		if err == ErrImageInValid {
			logger.ErrorWithCTX(ctx, "invalid image", err)
			api.BasicBadRequestError(ctx)
			return
		}
		logger.ErrorWithCTX(ctx, "validate image", err)
		api.BasicBadRequestError(ctx)
		return
	}

	issuedUUID, err := a.DB.CreateCollectionWithUserUUID(&body.Info, &claims.UserUUID)
	if err != nil {
		logger.ErrorWithCTX(ctx, "create collection with useruuid", err)
		api.BasicInternalServerError(ctx)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"uuid":    issuedUUID,
	})
}

func (a *CollectAPI) GetCollectionByUUID(ctx *gin.Context) {
	uuidParam := ctx.Param("uuid")
	collectionUUID, err := uuid.Parse(uuidParam)
	if err != nil {
		logger.ErrorWithCTX(ctx, "binding form", err)
		api.BasicBadRequestError(ctx)
		return
	}

	collection, err := a.DB.GetCollectionByUUID(&collectionUUID)
	if err != nil {
		logger.ErrorWithCTX(ctx, "get collection by uuid", err)
		api.BasicInternalServerError(ctx)
		return
	}

	ctx.JSON(http.StatusOK, collection)
}
