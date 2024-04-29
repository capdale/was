package collect

import (
	"context"
	"io"
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

type storage interface {
	GetCollectionJPG(ctx context.Context, uuid binaryuuid.UUID) (*[]byte, error)
	UploadCollectionJPG(ctx context.Context, uuid binaryuuid.UUID, reader io.Reader) error
	DeleteCollectionJPG(ctx context.Context, uuid binaryuuid.UUID) error
}

type database interface {
	GetUserIdByUUID(userUUID binaryuuid.UUID) (int64, error)
	GetCollectionUUIDs(userId int64, offset int, limit int) (*[]binaryuuid.UUID, error)
	GetCollectionByUUID(claimerId int64, collectionUUID *binaryuuid.UUID) (*Collection, error)
	CreateCollection(userId int64, collection *Collection, collectionUUID binaryuuid.UUID) error
	HasAccessPermissionCollection(claimerId int64, collectionUUID binaryuuid.UUID) error
	DeleteCollection(userUUID *binaryuuid.UUID, collectionUUID *binaryuuid.UUID) error
}

type CollectAPI struct {
	DB      database
	Storage storage
}

func New(database database, storage storage) *CollectAPI {
	return &CollectAPI{
		DB:      database,
		Storage: storage,
	}
}

type Collection = model.CollectionAPI

type getCollectionform struct {
	Offset *int `form:"offset" binding:"required,min=0"`
	Limit  *int `form:"limit" binding:"required,min=1,max=100"` // this not need pointer (because min is 1 never be 0), but for consistency
}

type getCollectionRes struct {
	Collections []binaryuuid.UUID `json:"collections"`
}

func (a *CollectAPI) GetCollectection(ctx *gin.Context) {
	form := &getCollectionform{}
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

	collections, err := a.DB.GetCollectionUUIDs(userId, *form.Offset, *form.Limit)
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

type createCollectionForm struct {
	Image *multipart.FileHeader `form:"image" binding:"required"`
	Info  Collection            `form:"info" json:"info" binding:"required"`
}

func (a *CollectAPI) CreateCollectionHandler(ctx *gin.Context) {
	claims := ctx.MustGet("claims").(*auth.AuthClaims)
	form := &createCollectionForm{}
	if err := ctx.Bind(form); err != nil {
		api.BasicBadRequestError(ctx)
		logger.ErrorWithCTX(ctx, "binding form", err)
		return
	}

	err := isValidImageFromFile(form.Image)
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

	collectionUUID, err := binaryuuid.NewRandom()
	if err != nil {
		api.BasicInternalServerError(ctx)
		logger.ErrorWithCTX(ctx, "create uuid", err)
		return
	}

	// s3 upload
	multipartFile, err := form.Image.Open()
	if err != nil {
		api.BasicInternalServerError(ctx)
		logger.ErrorWithCTX(ctx, "image upload", err)
		return
	}
	defer multipartFile.Close()

	if err := a.Storage.UploadCollectionJPG(ctx, collectionUUID, multipartFile); err != nil {
		api.BasicInternalServerError(ctx)
		logger.ErrorWithCTX(ctx, "upload jpg", err)
		return
	}

	err = a.DB.CreateCollection(userId, &form.Info, collectionUUID)
	if err != nil {
		api.BasicInternalServerError(ctx)
		logger.ErrorWithCTX(ctx, "create collection with useruuid", err)
		go a.Storage.DeleteCollectionJPG(context.Background(), collectionUUID) // send to delete queue
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"uuid": collectionUUID,
	})
}

type getCollectionForm struct {
	CollectionUUID string `uri:"uuid" binding:"required,uuid"`
}

func (a *CollectAPI) GetCollectionByUUID(ctx *gin.Context) {
	uri := &getCollectionForm{}
	if err := ctx.BindUri(uri); err != nil {
		api.BasicBadRequestError(ctx)
		logger.ErrorWithCTX(ctx, "binding", err)
		return
	}

	collectionUUID := binaryuuid.MustParse(uri.CollectionUUID)

	claimePtr, exist := ctx.Get("claims")
	var claimerId int64 = -1
	if exist {
		var err error
		claim := claimePtr.(*auth.AuthClaims)
		claimerId, err = a.DB.GetUserIdByUUID(claim.UUID)
		if err != nil {
			ctx.Status(http.StatusNotFound)
			logger.ErrorWithCTX(ctx, "id by uuid", err)
			return
		}
	}

	collection, err := a.DB.GetCollectionByUUID(claimerId, &collectionUUID)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "get collection by uuid", err)
		return
	}

	ctx.JSON(http.StatusOK, collection)
}

type getCollectionImageUri struct {
	ImageUUID string `uri:"uuid" binding:"required,uuid"`
}

func (a *CollectAPI) GetCollectionImageHandler(ctx *gin.Context) {
	uri := &getCollectionImageUri{}
	if err := ctx.BindUri(uri); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "binding uri", err)
		return
	}

	var claimerId int64 = -1

	claimsPtr, isExists := ctx.Get("claims")
	if isExists {
		var err error
		claims := claimsPtr.(*auth.AuthClaims)
		claimerId, err = a.DB.GetUserIdByUUID(claims.UUID)
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			return
		}
	}

	imageUUID := binaryuuid.MustParse(uri.ImageUUID)
	err := a.DB.HasAccessPermissionCollection(claimerId, imageUUID)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "check permission", err)
		return
	}

	imageBytes, err := a.Storage.GetCollectionJPG(ctx, imageUUID)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "get jpg", err)
		return
	}
	ctx.Data(http.StatusOK, "image/jpeg", *imageBytes)
}

type deleteCollectionUri struct {
	CollectionUUID string `uri:"uuid" binding:"required,uuid"`
}

func (a *CollectAPI) DeleteCollectionHandler(ctx *gin.Context) {
	uri := &deleteCollectionUri{}
	if err := ctx.BindUri(uri); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "binding uri", err)
		return
	}

	claims := ctx.MustGet("claims").(*auth.AuthClaims)
	collectionUUID := binaryuuid.MustParse(uri.CollectionUUID)

	if err := a.DB.DeleteCollection(&claims.UUID, &collectionUUID); err != nil {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "delete collection", err)
		return
	}

	ctx.Status(http.StatusNoContent)
	return
}
