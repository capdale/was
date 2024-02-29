package articleAPI

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/capdale/was/auth"
	baselogger "github.com/capdale/was/logger"
	"github.com/capdale/was/model"
	"github.com/capdale/was/types/binaryuuid"
	"github.com/gin-gonic/gin"
)

var logger = baselogger.Logger

type storage interface {
	UploadJPGs(ctx context.Context, filenames *[]string, readers *[]io.Reader) error
	DeleteJPG(filename string) (*s3.DeleteObjectOutput, error)
}

type database interface {
	GetArticleLinkIdsByUserId(userId int64, offset int, limit int) (*[]binaryuuid.UUID, error)
	GetUserIdByUUID(userUUID binaryuuid.UUID) (int64, error)
	GetArticle(writerId int64, linkId binaryuuid.UUID) (*model.ArticleAPI, error)
	CreateNewArticle(userId int64, title string, content string, collectionUUIDs *[]binaryuuid.UUID, imageUUIDs *[]binaryuuid.UUID, collectionOrder *[]uint8) error
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
	Article      articleForm             `form:"article"`
	ImageHeaders []*multipart.FileHeader `form:"image[]" json:"image[]"`
}

type articleForm struct {
	Title           string   `form:"title" json:"title" binding:"required,min=4,max=32"`
	Content         string   `form:"content" json:"content" binding:"required,min=8,max=512"`
	CollectionUUIDs []string `form:"collections" json:"collections" binding:"required,min=1,max=10,dive,uuid"`
	Order           []uint8  `form:"order" json:"order" binding:"required"` // provide order information where collection will be ordered in
}

var ErrInvalidOrder = errors.New("invalid order")

func (a *ArticleAPI) CreateArticleHandler(ctx *gin.Context) {
	claims := ctx.MustGet("claims").(*auth.AuthClaims)
	form := &createArticleForm{}
	if err := ctx.Bind(form); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid form"})
		logger.ErrorWithCTX(ctx, "binding error", err)
		return
	}

	imageCount := len(form.ImageHeaders)
	collectionCount := uint8(len(form.Article.CollectionUUIDs))
	if collectionCount != uint8(len(form.Article.Order)) {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid form"})
		logger.ErrorWithCTX(ctx, "order / image count is not equal", nil)
		return
	}

	for _, order := range form.Article.Order {
		if order > collectionCount { // uint8, so no need to check sign of number
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "bad request, order is invalid"})
			logger.ErrorWithCTX(ctx, "order invalid", ErrInvalidOrder)
			return
		}
	}

	collectionUUIDs := make([]binaryuuid.UUID, len(form.Article.CollectionUUIDs))
	for i, cuidStr := range form.Article.CollectionUUIDs {
		// validate while bind (Validator)
		cuid := binaryuuid.MustParse(cuidStr)
		collectionUUIDs[i] = cuid
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

	// no check uuids duplicated, since uuidv4 duplicate probability is very low, err when insert to DB with unique key

	// upload image first, for consistency, if database write success and imag write file, need to rollback but rollback can be also failed. Then its hard to track and recover
	if err := a.uploadImagesWithUUID(ctx, &imageUUIDs, &form.ImageHeaders); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		logger.ErrorWithCTX(ctx, "upload image", err)
		return
	}

	userId, err := a.d.GetUserIdByUUID(claims.UUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		logger.ErrorWithCTX(ctx, "get userid by uuid", err)
		return
	}

	if err := a.d.CreateNewArticle(userId, form.Article.Title, form.Article.Content, &collectionUUIDs, &imageUUIDs, &form.Article.Order); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		logger.ErrorWithCTX(ctx, "create new article", err)
		return
	}

	ctx.JSON(http.StatusAccepted, gin.H{"message": "ok"})
}

type getArticlesByUserUUIDUri struct {
	UserUUIDStr string `uri:"useruuid" binding:"required,uuid"`
}

type getArticlesByUserUUIDForm struct {
	Offset int `form:"offset,default=0" binding:"min=0"`
	Limit  int `form:"limit,default=20" binding:"min=1,max=20"`
}

func (a *ArticleAPI) GetUserArticleLinksHandler(ctx *gin.Context) {
	form := &getArticlesByUserUUIDForm{}
	if err := ctx.Bind(form); err != nil {
		logger.ErrorWithCTX(ctx, "bind form", err)
		return
	}

	uri := &getArticlesByUserUUIDUri{}
	if err := ctx.BindUri(uri); err != nil {
		logger.ErrorWithCTX(ctx, "bind uri", err)
		return
	}

	// validate with Validator,
	userUUID := binaryuuid.MustParse(uri.UserUUIDStr)
	userId, err := a.d.GetUserIdByUUID(userUUID)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "get userid by uuid", err)
		return
	}

	articles, err := a.d.GetArticleLinkIdsByUserId(userId, form.Offset, form.Limit)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "query linkids by user id", err)
		return
	}

	links := make([]string, len(*articles))
	for i, article := range *articles {
		links[i] = base64.URLEncoding.EncodeToString(append(userUUID[:], article[:]...))
	}

	ctx.JSON(http.StatusOK, gin.H{"links": links})
}

func (a *ArticleAPI) GetArticleHandler(ctx *gin.Context) {
	link := ctx.Param("link")
	linkBytes, err := base64.URLEncoding.DecodeString(link)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "decode link", nil)
		return
	}

	if len(linkBytes) != 32 {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "link bytes error", nil)
		return
	}

	userUUIDBytes := linkBytes[0:16]
	linkIdBytes := linkBytes[16:]
	userUUID, err := binaryuuid.FromBytes(userUUIDBytes)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "parse user uuid", err)
		return
	}
	linkId, err := binaryuuid.FromBytes(linkIdBytes)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "parse link id", err)
		return
	}

	userId, err := a.d.GetUserIdByUUID(userUUID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
		logger.ErrorWithCTX(ctx, "query user id by uuid", err)
		return
	}

	article, err := a.d.GetArticle(userId, linkId)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "get article", err)
		return
	}
	ctx.JSON(http.StatusOK, article)
}
