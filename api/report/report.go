package reportAPI

import (
	"net/http"

	"github.com/capdale/was/api"
	articleAPI "github.com/capdale/was/api/article"
	baseLogger "github.com/capdale/was/logger"
	"github.com/capdale/was/model"
	"github.com/capdale/was/types/binaryuuid"
	"github.com/capdale/was/types/claimer"
	"github.com/gin-gonic/gin"
)

var logger = baseLogger.Logger

type database interface {
	CreateReportUser(issuerUUID *claimer.Claimer, targetUUID *binaryuuid.UUID, detailType int, description string) error
	CreateReportArticle(issuerUUID *claimer.Claimer, targetarticleUUID *binaryuuid.UUID, detailType int, description string) error
	CreateReportBug(issuerUUID *claimer.Claimer, title string, description string) error
	CreateReportHelp(issuerUUID *claimer.Claimer, title string, description string) error
	CreateReportEtc(issuerUUID *claimer.Claimer, title string, description string) error
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
	TargetUUID       *string `json:"target" binding:"required,uuid"`
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

	issuerUUID := api.GetClaimer(ctx)
	targetUUID := binaryuuid.MustParse(*form.TargetUUID)
	if err := r.d.CreateReportUser(issuerUUID, &targetUUID, *form.ReportDetailType, form.Description); err != nil {
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

	issuerUUID := api.GetClaimer(ctx)
	if err := r.d.CreateReportArticle(issuerUUID, articleUUID, *form.ReportDetailType, form.Description); err != nil {
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

	issuerUUID := api.GetClaimer(ctx)
	if err := r.d.CreateReportBug(issuerUUID, *form.Title, *form.Description); err != nil {
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

	issuerUUID := api.GetClaimer(ctx)
	if err = r.d.CreateReportHelp(issuerUUID, *form.Title, *form.Description); err != nil {
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

	issuerUUID := api.GetClaimer(ctx)
	if err = r.d.CreateReportEtc(issuerUUID, *form.Title, *form.Description); err != nil {
		api.BasicInternalServerError(ctx)
		logger.ErrorWithCTX(ctx, "create report etc", err)
		return
	}
	ctx.JSON(http.StatusAccepted, gin.H{"message": "ok"})
}
