package articleAPI

import (
	"encoding/base64"
	"net/http"

	"github.com/capdale/was/auth"
	baselogger "github.com/capdale/was/logger"
	"github.com/capdale/was/model"
	"github.com/capdale/was/types/binaryuuid"
	"github.com/gin-gonic/gin"
)

var logger = baselogger.Logger

type database interface {
	GetArticleLinkIdsByUserId(userId int64, offset int, limit int) (*[]binaryuuid.UUID, error)
	GetUserIdByUUID(userUUID binaryuuid.UUID) (int64, error)
	GetArticle(writerId int64, linkId binaryuuid.UUID) (*model.ArticleAPI, error)
	CreateNewArticle(userId int64, title string, content string, collectionUUIDs *[]binaryuuid.UUID) error
}

type ArticleAPI struct {
	d database
}

func New(d database) *ArticleAPI {
	return &ArticleAPI{
		d: d,
	}
}

type createArticleForm struct {
	Title           string   `json:"title" binding:"required,min=4,max=32"`
	Content         string   `json:"content" binding:"required,min=8,max=512"`
	CollectionUUIDs []string `json:"collections" binding:"required,dive,uuid"`
}

func (a *ArticleAPI) CreateArticleHandler(ctx *gin.Context) {
	claims := ctx.MustGet("claims").(*auth.AuthClaims)
	form := &createArticleForm{}
	if err := ctx.Bind(form); err != nil {
		logger.ErrorWithCTX(ctx, "binding error", err)
		return
	}

	collectionUUIDs := []binaryuuid.UUID{}
	for _, cuidStr := range form.CollectionUUIDs {
		// validate while bind (Validator)
		cuid := binaryuuid.MustParse(cuidStr)
		collectionUUIDs = append(collectionUUIDs, cuid)
	}

	userId, err := a.d.GetUserIdByUUID(claims.UUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		logger.ErrorWithCTX(ctx, "get userid by uuid", err)
		return
	}

	if err = a.d.CreateNewArticle(userId, form.Title, form.Content, &collectionUUIDs); err != nil {
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
