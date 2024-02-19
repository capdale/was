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
)

var logger = baseLogger.Logger

type database interface {
	GetUserIdByUUID(userUUID binaryuuid.UUID) (int64, error)
	GetCollectionUUIDs(userId int64, offset int, limit int) (*[]binaryuuid.UUID, error)
	GetCollectionByUUID(collectionUUID *binaryuuid.UUID) (*Collection, error)
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

type Collection = model.CollectionAPI

type getCollectionform struct {
	Offset int `json:"offset" form:"offset" binding:"required,min=0"`
	Limit  int `json:"limit" form:"limit" binding:"required,min=1,max=100"`
}

type getCollectionRes struct {
	Collections []binaryuuid.UUID `json:"collections"`
}

func (a *CollectAPI) GetCollectection(ctx *gin.Context) {
	form := getCollectionform{}
	if err := ctx.Bind(form); err != nil {
		logger.ErrorWithCTX(ctx, "binding form", err)
		return
	}

	claims := ctx.MustGet("claims").(*auth.AuthClaims)
	userId, err := a.DB.GetUserIdByUUID(claims.UUID)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "query user id by uuid", err)
		return
	}

	collections, err := a.DB.GetCollectionUUIDs(userId, form.Offset, form.Limit)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "db get collections", err)
		return
	}

	res := &getCollectionRes{
		Collections: *collections,
	}
	ctx.JSON(http.StatusOK, res)
}

type postCollectionForm struct {
	ImageFile *multipart.FileHeader `form:"image" binding:"required"`
	Info      Collection            `form:"info" binding:"required"`
}

func (a *CollectAPI) PostCollection(ctx *gin.Context) {
	claims := ctx.MustGet("claims").(*auth.AuthClaims)
	form := postCollectionForm{}
	if err := ctx.ShouldBind(form); err != nil {
		api.BasicBadRequestError(ctx)
		logger.ErrorWithCTX(ctx, "binding form", err)
		return
	}

	err := isValidImageFromFile(form.ImageFile)
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
		api.BasicBadRequestError(ctx)
		logger.ErrorWithCTX(ctx, "query user id by uuid", err)
		return
	}

	issuedUUID, err := a.DB.CreateCollection(&form.Info, userId)
	if err != nil {
		api.BasicInternalServerError(ctx)
		logger.ErrorWithCTX(ctx, "create collection with useruuid", err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"uuid": issuedUUID,
	})
}

func (a *CollectAPI) GetCollectionByUUID(ctx *gin.Context) {
	uuidParam := ctx.Param("uuid")
	collectionUUID, err := binaryuuid.Parse(uuidParam)
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
