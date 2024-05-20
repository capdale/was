package articleAPI

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/capdale/was/api"
	baselogger "github.com/capdale/was/logger"
	"github.com/capdale/was/model"
	"github.com/capdale/was/types/binaryuuid"
	"github.com/capdale/was/types/claimer"
	"github.com/gin-gonic/gin"
)

var logger = baselogger.Logger

type storage interface {
	GetArticleJPG(ctx context.Context, uuid binaryuuid.UUID) (*[]byte, error)
	UploadArticleJPGs(ctx context.Context, uuids *[]binaryuuid.UUID, readers *[]io.Reader) error
}

type database interface {
	IsCollectionOwned(claimer *claimer.Claimer, collectionUUIDs *[]binaryuuid.UUID) error
	GetArticleLinkIdsByUsername(claimer *claimer.Claimer, username *string, offset int, limit int) (*[]*binaryuuid.UUID, error)
	GetPublicArticleLinks(offset int, limit int) (*[]*binaryuuid.UUID, error)
	GetArticle(claimer *claimer.Claimer, linkId binaryuuid.UUID) (*model.ArticleAPI, error)
	CreateNewArticle(claimer *claimer.Claimer, title string, content string, collectionUUIDs *[]binaryuuid.UUID, imageUUIDs *[]binaryuuid.UUID, collectionOrder *[]uint8) error
	HasAccessPermissionArticleImage(claimer *claimer.Claimer, imageUUID *binaryuuid.UUID) (bool, error)
	DeleteArticle(claimer *claimer.Claimer, articleLinkId *binaryuuid.UUID) error

	Comment(claimer *claimer.Claimer, articleId *binaryuuid.UUID, comment *string) error
	GetComments(claimer *claimer.Claimer, articleId *binaryuuid.UUID, offset uint, limit uint) (*[]model.ArticleCommentAPI, error)
	GetHeartState(claimer *claimer.Claimer, articleUUID *binaryuuid.UUID) (bool, error)
	DoHeart(claimer *claimer.Claimer, aritcleId *binaryuuid.UUID, action int) error
	CountHeart(claimer *claimer.Claimer, articleId *binaryuuid.UUID) (uint64, error)
}

type ArticleAPI struct {
	d       database
	Storage storage
}

func New(d database, storage storage) *ArticleAPI {
	return &ArticleAPI{
		d:       d,
		Storage: storage,
	}
}

var ErrInvalidForm = errors.New("form is invalid")

type createArticleForm struct {
	Article      articleForm             `form:"article" binding:"required"`
	ImageHeaders []*multipart.FileHeader `form:"image[]"`
}

type articleForm struct {
	Title           string           `form:"title" json:"title" binding:"required,min=4,max=32"`
	Content         string           `form:"content" json:"content" binding:"required,min=8,max=512"`
	CollectionInfos []collectionInfo `form:"collections" json:"collections" binding:"required,min=1"`
}

type collectionInfo struct {
	UUID  string `form:"uuid" binding:"required,uuid"`
	Order *uint8 `form:"order" binding:"required"`
}

var ErrInvalidOrder = errors.New("invalid order")

func (a *ArticleAPI) CreateArticleHandler(ctx *gin.Context) {
	form := &createArticleForm{}
	if err := ctx.Bind(form); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid form"})
		logger.ErrorWithCTX(ctx, "binding error", err)
		return
	}

	imageCount := len(form.ImageHeaders)
	collectionCount := uint8(len(form.Article.CollectionInfos))
	collectionUUIDs := make([]binaryuuid.UUID, len(form.Article.CollectionInfos))
	orders := make([]uint8, len(form.Article.CollectionInfos))
	for i, collectionInfo := range form.Article.CollectionInfos {
		if *collectionInfo.Order > collectionCount { // uint8, so no need to check sign of number
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "bad request, order is invalid"})
			logger.ErrorWithCTX(ctx, "order invalid", ErrInvalidOrder)
			return
		}
		cuid := binaryuuid.MustParse(collectionInfo.UUID)
		collectionUUIDs[i] = cuid
		orders[i] = *collectionInfo.Order
	}

	imageUUIDs := make([]binaryuuid.UUID, imageCount)
	for i := 0; i < imageCount; i++ {
		buid, err := binaryuuid.NewRandom()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
			logger.ErrorWithCTX(ctx, "create image uuid", err)
			return
		}
		imageUUIDs[i] = buid
	}

	claimerAuthUUID := api.MustGetClaimer(ctx)

	// no check uuids duplicated, since uuidv4 duplicate probability is very low, err when insert to DB with unique key

	// check collection is owned
	if err := a.d.IsCollectionOwned(claimerAuthUUID, &collectionUUIDs); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
		logger.ErrorWithCTX(ctx, "bad request", err)
		return
	}

	// upload image first, for consistency, if database write success and imag write file, need to rollback but rollback can be also failed. Then its hard to track and recover
	if err := a.uploadImagesWithUUID(ctx, &imageUUIDs, &form.ImageHeaders); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		logger.ErrorWithCTX(ctx, "upload image", err)
		return
	}

	if err := a.d.CreateNewArticle(claimerAuthUUID, form.Article.Title, form.Article.Content, &collectionUUIDs, &imageUUIDs, &orders); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		logger.ErrorWithCTX(ctx, "create new article", err)
		return
	}

	ctx.JSON(http.StatusAccepted, gin.H{"message": "ok"})
}

type getArticleLinksUri struct {
	Targetname string `uri:"targetname" binding:"required"`
}

type getArticleLinksForm struct {
	Offset int `form:"offset,default=0" binding:"min=0"`
	Limit  int `form:"limit,default=20" binding:"min=1,max=20"`
}

func (a *ArticleAPI) GetUserArticleLinksHandler(ctx *gin.Context) {
	form := &getArticleLinksForm{}
	if err := ctx.Bind(form); err != nil {
		logger.ErrorWithCTX(ctx, "bind form", err)
		return
	}

	uri := &getArticleLinksUri{}
	if err := ctx.BindUri(uri); err != nil {
		logger.ErrorWithCTX(ctx, "bind uri", err)
		return
	}

	claimerAuthUUID := api.GetClaimer(ctx)
	articles, err := a.d.GetArticleLinkIdsByUsername(claimerAuthUUID, &uri.Targetname, form.Offset, form.Limit)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "query linkids by user id", err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"links": articles})
}

type getArticleHandlerUri struct {
	UUID string `uri:"link" binding:"required,uuid"`
}

func (a *ArticleAPI) GetArticleHandler(ctx *gin.Context) {
	uri := &getArticleHandlerUri{}
	if err := ctx.BindUri(uri); err != nil {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "binding uri error", err)
		return
	}

	linkId := binaryuuid.MustParse(uri.UUID)
	claimerAuthUUID := api.GetClaimer(ctx)
	article, err := a.d.GetArticle(claimerAuthUUID, linkId)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "get article", err)
		return
	}
	ctx.JSON(http.StatusOK, article)
}

type getArticleImageHandlerUri struct {
	ImageUUID string `uri:"uuid" binding:"required,uuid"`
}

func (a *ArticleAPI) GetArticleImageHandler(ctx *gin.Context) {
	uri := &getArticleImageHandlerUri{}
	if err := ctx.BindUri(uri); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "binding uri", err)
		return
	}

	imageUUID := binaryuuid.MustParse(uri.ImageUUID)
	claimerAuthUUID := api.GetClaimer(ctx)
	hasPermission, err := a.d.HasAccessPermissionArticleImage(claimerAuthUUID, &imageUUID)
	if err != nil || !hasPermission {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "check permission", err)
		return
	}

	imageBytes, err := a.Storage.GetArticleJPG(ctx, imageUUID)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "get jpg", err)
		return
	}
	ctx.Data(http.StatusOK, "image/jpeg", *imageBytes)
}

type deleteArticleHandlerUri struct {
	ArticleLink string `uri:"link" binding:"required,uuid"`
}

func (a *ArticleAPI) DeleteArticleHandler(ctx *gin.Context) {
	uri := &deleteArticleHandlerUri{}
	if err := ctx.BindUri(uri); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "binding uri", err)
		return
	}

	articleId := binaryuuid.MustParse(uri.ArticleLink)
	claimerAuthUUID := api.MustGetClaimer(ctx)
	if err := a.d.DeleteArticle(claimerAuthUUID, &articleId); err != nil {
		ctx.Status(http.StatusInternalServerError)
		logger.ErrorWithCTX(ctx, "delete article", err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

type postCommentHandlerUri struct {
	ArticleId string `uri:"link" binding:"uuid"`
}

type postCommentHandlerForm struct {
	Comment string `form:"comment" binding:"min=1,max=255"`
}

func (a *ArticleAPI) PostCommentHandler(ctx *gin.Context) {
	uri := &postCommentHandlerUri{}
	if err := ctx.BindUri(uri); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "bind uri error", err)
		return
	}

	form := &postCommentHandlerForm{}
	if err := ctx.Bind(form); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "bind form error", err)
		return
	}

	articleId := binaryuuid.MustParse(uri.ArticleId)
	claimer := api.MustGetClaimer(ctx)

	if err := a.d.Comment(claimer, &articleId, &form.Comment); err != nil {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "comment error", err)
		return
	}
	ctx.Status(http.StatusAccepted)
}

type getCommentsHandlerUri struct {
	ArticleId string `uri:"link" binding:"uuid"`
}

type getCommentsHandlerForm struct {
	Offset uint `form:"offset,default=0" binding:"min=0"`
	Limit  uint `form:"limit,default=16" binding:"min=1,max=64"`
}

func (a *ArticleAPI) GetCommentsHandler(ctx *gin.Context) {
	uri := &getCommentsHandlerUri{}
	if err := ctx.BindUri(uri); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "binding error", err)
		return
	}

	form := &getCommentsHandlerForm{}
	if err := ctx.Bind(form); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "binding error", err)
		return
	}

	claimer := api.GetClaimer(ctx)
	linkId := binaryuuid.MustParse(uri.ArticleId)
	comments, err := a.d.GetComments(claimer, &linkId, form.Offset, form.Limit)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "get comment", err)
		return
	}

	ctx.JSON(http.StatusOK, comments)
}

type heartHandlerUri = postCommentHandlerUri
type heartHandlerForm struct {
	Action string `form:"action" binding:"oneof=apply cancel"`
}

func (a *ArticleAPI) HeartHandler(ctx *gin.Context) {
	uri := &heartHandlerUri{}
	if err := ctx.BindUri(uri); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "bind uri error", err)
		return
	}
	form := &heartHandlerForm{}
	if err := ctx.Bind(form); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "bind form error", err)
		return
	}
	claimer := api.MustGetClaimer(ctx)
	articleId := binaryuuid.MustParse(uri.ArticleId)
	var action int
	if form.Action == "apply" {
		action = 1
	} else if form.Action == "cancel" {
		action = 0
	}
	if err := a.d.DoHeart(claimer, &articleId, action); err != nil {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "do heart error", err)
		return
	}

	ctx.Status(http.StatusAccepted)
}

type getHeartCountHandlerUri = postCommentHandlerUri

func (a *ArticleAPI) GetHeartCountHandler(ctx *gin.Context) {
	uri := &getHeartCountHandlerUri{}
	if err := ctx.BindUri(uri); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "bind uri error", err)
		return
	}

	claimer := api.GetClaimer(ctx)
	articleId := binaryuuid.MustParse(uri.ArticleId)
	heartCounts, err := a.d.CountHeart(claimer, &articleId)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "count heart error", err)
		return
	}

	ctx.JSON(http.StatusOK, &gin.H{
		"counts": heartCounts,
	})
}

type getHeartStateHandlerUri = postCommentHandlerUri

func (a *ArticleAPI) GetHeartStateHandler(ctx *gin.Context) {
	uri := &getHeartStateHandlerUri{}
	if err := ctx.BindUri(uri); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "bind uri", err)
		return
	}

	claimer := api.MustGetClaimer(ctx)
	articleUUID := binaryuuid.MustParse(uri.ArticleId)

	heart, err := a.d.GetHeartState(claimer, &articleUUID)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "get heart state", err)
		return
	}

	ctx.JSON(http.StatusOK, &gin.H{
		"heart": heart,
	})
}

type getPublicArticlesHandlerForm struct {
	Offset int `form:"offset,default=0" binding:"min=0"`
	Limit  int `form:"limit,default=16" binding:"min=1,max=64"`
}

func (a *ArticleAPI) GetPublicArticlesHandler(ctx *gin.Context) {
	form := &getPublicArticlesHandlerForm{}
	if err := ctx.Bind(form); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "bind form", err)
		return
	}

	links, err := a.d.GetPublicArticleLinks(form.Offset, form.Limit)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		logger.ErrorWithCTX(ctx, "get public articles", err)
		return
	}

	ctx.JSON(http.StatusOK, &gin.H{
		"links": links,
	})
}
