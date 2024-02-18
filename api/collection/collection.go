package collect

import (
	"mime/multipart"
	"net/http"

	"github.com/capdale/was/api"
	"github.com/capdale/was/auth"
	baseLogger "github.com/capdale/was/logger"
	"github.com/capdale/was/model"
	"github.com/capdale/was/types/binaryuuid"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var logger = baseLogger.Logger

type database interface {
	GetUserIdByUUID(userUUID binaryuuid.UUID) (int64, error)
	GetCollectionUUIDs(userId int64, offset int, limit int) (*[]binaryuuid.UUID, error)
	GetCollectionByUUID(collectionUUID *uuid.UUID) (*Collection, error)
	CreateCollection(collection *Collection, useId int64) (collectionUUID binaryuuid.UUID, err error)
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
	Collections []binaryuuid.UUID `json:"collections"`
}

type PostCollectionReq struct {
	ImageFile *multipart.FileHeader `form:"image" binding:"required"`
	Info      Collection            `form:"info" binding:"required"`
}

type Collection = model.CollectionAPI

func (a *CollectAPI) GetCollectection(ctx *gin.Context) {
	var req GetCollectionReq
	if err := ctx.ShouldBind(&req); err != nil {
		api.BasicBadRequestError(ctx)
		logger.ErrorWithCTX(ctx, "binding form", err)
		return
	}

	claims := ctx.MustGet("claims").(*auth.AuthClaims)
	userId, err := a.DB.GetUserIdByUUID(claims.UUID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
		logger.ErrorWithCTX(ctx, "query user id by uuid", err)
		return
	}

	collections, err := a.DB.GetCollectionUUIDs(userId, *req.Offset, *req.Limit)
	if err != nil {
		api.BasicInternalServerError(ctx)
		logger.ErrorWithCTX(ctx, "db get collections", err)
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
		api.BasicBadRequestError(ctx)
		logger.ErrorWithCTX(ctx, "binding form", err)
		return
	}

	err := isValidImageFromFile(body.ImageFile)
	if err != nil {
		if err == ErrImageInValid {
			api.BasicBadRequestError(ctx)
			logger.ErrorWithCTX(ctx, "invalid image", err)
			return
		}
		api.BasicBadRequestError(ctx)
		logger.ErrorWithCTX(ctx, "validate image", err)
		return
	}

	userId, err := a.DB.GetUserIdByUUID(claims.UUID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
		logger.ErrorWithCTX(ctx, "query user id by uuid", err)
		return
	}

	issuedUUID, err := a.DB.CreateCollection(&body.Info, userId)
	if err != nil {
		api.BasicInternalServerError(ctx)
		logger.ErrorWithCTX(ctx, "create collection with useruuid", err)
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
		api.BasicBadRequestError(ctx)
		logger.ErrorWithCTX(ctx, "binding form", err)
		return
	}

	collection, err := a.DB.GetCollectionByUUID(&collectionUUID)
	if err != nil {
		api.BasicInternalServerError(ctx)
		logger.ErrorWithCTX(ctx, "get collection by uuid", err)
		return
	}

	ctx.JSON(http.StatusOK, collection)
}
