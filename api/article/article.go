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
	Title           string   `json:"title"`
	Content         string   `json:"content"`
	CollectionUUIDs []string `json:"collections"`
}

func (a *ArticleAPI) CreateArticleHandler(ctx *gin.Context) {
	claims := ctx.MustGet("claims").(*auth.AuthClaims)
	form := &createArticleForm{}
	err := ctx.ShouldBind(form)
	collectionUUIDs := []binaryuuid.UUID{}
	for _, cuidStr := range form.CollectionUUIDs {
		cuid, err := binaryuuid.Parse(cuidStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
			return
		}
		collectionUUIDs = append(collectionUUIDs, cuid)
	}

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
		return
	}

	userId, err := a.d.GetUserIdByUUID(claims.UUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}

	if err = a.d.CreateNewArticle(userId, form.Title, form.Content, &collectionUUIDs); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}
	ctx.JSON(http.StatusAccepted, gin.H{"message": "ok"})
}

type getArticlesByUserUUIDForm struct {
	Offset *int `form:"offset"`
	Limit  *int `form:"limit"`
}

func (a *ArticleAPI) GetUserArticleLinksHandler(ctx *gin.Context) {
	userUUIDStr := ctx.Param("useruuid")
	userUUID, err := binaryuuid.Parse(userUUIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
		logger.ErrorWithCTX(ctx, "useruuid parse", err)
		return
	}
	form := &getArticlesByUserUUIDForm{}
	err = ctx.ShouldBind(form)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
		logger.ErrorWithCTX(ctx, "bind form", err)
		return
	}

	offset := 0
	limit := 20
	if form.Limit != nil {
		limit = *form.Limit
	}
	if form.Offset != nil {
		offset = *form.Offset
	}
	if offset < 0 || limit < 1 || limit > 100 {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
		logger.ErrorWithCTX(ctx, "bind limit error", nil)
		return
	}

	userId, err := a.d.GetUserIdByUUID(userUUID)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "get userid by uuid", err)
		return
	}

	articles, err := a.d.GetArticleLinkIdsByUserId(userId, offset, limit)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "query linkids by user id", err)
		return
	}

	links := make([]string, len(*articles))
	for i, article := range *articles {
		links[i] = base64.URLEncoding.EncodeToString(append(userUUID[:], article[:]...))
	}

	ctx.JSON(http.StatusOK, links)
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
