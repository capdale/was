package reportAPI

import (
	"net/http"

	"github.com/capdale/was/api"
	articleAPI "github.com/capdale/was/api/article"
	"github.com/capdale/was/auth"
	baseLogger "github.com/capdale/was/logger"
	"github.com/capdale/was/model"
	"github.com/capdale/was/types/binaryuuid"
	"github.com/gin-gonic/gin"
)

var logger = baseLogger.Logger

type database interface {
	GetUserIdByUUID(userUUID binaryuuid.UUID) (int64, error)
	CreateReportUser(issuerId int64, targetUsername string, detailType int, description string) error
	CreateReportArticle(issuerId int64, targetarticleUUID binaryuuid.UUID, detailType int, description string) error
	CreateReportBug(issuerId int64, title string, description string) error
	CreateReportHelp(issuerId int64, title string, description string) error
	CreateReportEtc(issuerId int64, title string, description string) error
}

type ReportAPI struct {
	d database
}

func New(d database) *ReportAPI {
	return &ReportAPI{
		d: d,
	}
}

type postReportUserForm struct {
	ReportDetailType *int    `json:"report_detail_type" binding:"required"`
	TargetUsername   *string `json:"username" binding:"required"`
	Description      string  `json:"description"`
}

func (r *ReportAPI) PostUserReportHandler(ctx *gin.Context) {
	// validate form start
	form := &postReportUserForm{}
	if err := ctx.Bind(form); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "bind form", err)
		return
	}

	if *form.ReportDetailType < model.ReportUserMin || *form.ReportDetailType > model.ReportUserMax {
		api.BasicBadRequestError(ctx)
		logger.ErrorWithCTX(ctx, "binding form, detail type", nil)
		return
	}
	// validate form end

	var issuerId int64 = -1 // defulat issuer is -1 = anonymous
	claimsPtr, isExist := ctx.Get("claims")
	if isExist {
		var err error
		claims := claimsPtr.(*auth.AuthClaims)
		issuerId, err = r.d.GetUserIdByUUID(claims.UUID)
		if err != nil {
			api.BasicInternalServerError(ctx)
			logger.ErrorWithCTX(ctx, "exchange id - uuid", err)
			return
		}
	}

	if err := r.d.CreateReportUser(issuerId, *form.TargetUsername, *form.ReportDetailType, form.Description); err != nil {
		api.BasicInternalServerError(ctx)
		logger.ErrorWithCTX(ctx, "create report user", err)
		return
	}
	ctx.JSON(http.StatusAccepted, gin.H{"message": "ok"})
}

type postReportArticleForm struct {
	ReportDetailType  *int    `json:"report_detail_type" binding:"required"`
	TargetArticleLink *string `json:"target_article_link" binding:"required"`
	Description       string  `json:"description"`
}

func (r *ReportAPI) PostReportArticleHandler(ctx *gin.Context) {
	// validate form start
	form := &postReportArticleForm{}
	if err := ctx.Bind(form); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "binding form", err)
		return
	}

	if *form.ReportDetailType < model.ReportArticleMin || *form.ReportDetailType > model.ReportArticleMax {
		api.BasicBadRequestError(ctx)
		logger.ErrorWithCTX(ctx, "binding form, detail type", nil)
		return
	}

	articleUUID, err := articleAPI.DecodeLink(*form.TargetArticleLink)
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "link decode", err)
		return
	}
	// validate form end

	var issuerId int64 = -1 // defulat issuer is -1 = anonymous
	claimsPtr, isExist := ctx.Get("claims")
	if isExist {
		var err error
		claims := claimsPtr.(*auth.AuthClaims)
		issuerId, err = r.d.GetUserIdByUUID(claims.UUID)
		if err != nil {
			api.BasicInternalServerError(ctx)
			logger.ErrorWithCTX(ctx, "exchange id - uuid", err)
			return
		}
	}

	if err := r.d.CreateReportArticle(issuerId, *articleUUID, *form.ReportDetailType, form.Description); err != nil {
		api.BasicInternalServerError(ctx)
		logger.ErrorWithCTX(ctx, "create report article", err)
		return
	}
	ctx.JSON(http.StatusAccepted, gin.H{"message": "ok"})
}

type postReportBugForm struct {
	Title       *string `json:"title" binding:"required"`
	Description *string `json:"description" binding:"required"`
}

func (r *ReportAPI) PostReportBugHandler(ctx *gin.Context) {
	// validate form start
	form := &postReportBugForm{}
	if err := ctx.Bind(form); err != nil {
		api.BasicBadRequestError(ctx)
		logger.ErrorWithCTX(ctx, "binding form", err)
		return
	}
	// validate form end

	var issuerId int64 = -1 // defulat issuer is -1 = anonymous
	claimsPtr, isExist := ctx.Get("claims")
	if isExist {
		var err error
		claims := claimsPtr.(*auth.AuthClaims)
		issuerId, err = r.d.GetUserIdByUUID(claims.UUID)
		if err != nil {
			api.BasicInternalServerError(ctx)
			logger.ErrorWithCTX(ctx, "exchange id - uuid", err)
			return
		}
	}

	if err := r.d.CreateReportBug(issuerId, *form.Title, *form.Description); err != nil {
		api.BasicInternalServerError(ctx)
		logger.ErrorWithCTX(ctx, "creat report bug", err)
	}
	ctx.JSON(http.StatusAccepted, gin.H{"message": "ok"})
}

type postReportHelpForm struct {
	Title       *string `json:"title" binding:"required"`
	Description *string `json:"description" binding:"required"`
}

func (r *ReportAPI) PostReportHelpHandler(ctx *gin.Context) {
	// validate form start
	form := new(postReportHelpForm)
	err := ctx.ShouldBind(form)
	if err != nil {
		api.BasicBadRequestError(ctx)
		logger.ErrorWithCTX(ctx, "binding form", err)
		return
	}
	// validate form end

	var issuerId int64 = -1 // defulat issuer is -1 = anonymous
	claimsPtr, isExist := ctx.Get("claims")
	if isExist {
		claims := claimsPtr.(*auth.AuthClaims)
		issuerId, err = r.d.GetUserIdByUUID(claims.UUID)
		if err != nil {
			api.BasicInternalServerError(ctx)
			logger.ErrorWithCTX(ctx, "exchange id - uuid", err)
			return
		}
	}

	if err = r.d.CreateReportHelp(issuerId, *form.Title, *form.Description); err != nil {
		api.BasicInternalServerError(ctx)
		logger.ErrorWithCTX(ctx, "create report help", err)
		return
	}
	ctx.JSON(http.StatusAccepted, gin.H{"message": "ok"})
}

type postReportEtcForm struct {
	Title       *string `json:"title" binding:"required"`
	Description *string `json:"description" binding:"required"`
}

func (r *ReportAPI) PostReportEtcHandler(ctx *gin.Context) {
	// validate form start
	form := new(postReportEtcForm)
	err := ctx.ShouldBind(form)
	if err != nil {
		api.BasicBadRequestError(ctx)
		logger.ErrorWithCTX(ctx, "binding form", err)
		return
	}
	// validate form end

	var issuerId int64 = -1 // defulat issuer is -1 = anonymous
	claimsPtr, isExist := ctx.Get("claims")
	if isExist {
		claims := claimsPtr.(*auth.AuthClaims)
		issuerId, err = r.d.GetUserIdByUUID(claims.UUID)
		if err != nil {
			api.BasicInternalServerError(ctx)
			logger.ErrorWithCTX(ctx, "exchange id - uuid", err)
			return
		}
	}

	if err = r.d.CreateReportEtc(issuerId, *form.Title, *form.Description); err != nil {
		api.BasicInternalServerError(ctx)
		logger.ErrorWithCTX(ctx, "create report etc", err)
		return
	}
	ctx.JSON(http.StatusAccepted, gin.H{"message": "ok"})
}
