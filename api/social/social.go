package socialAPI

import (
	"net/http"
	"strconv"

	"github.com/capdale/was/api"
	baselogger "github.com/capdale/was/logger"

	"github.com/capdale/was/types/binaryuuid"
	"github.com/capdale/was/types/claimer"
	"github.com/gin-gonic/gin"
)

var logger = baselogger.Logger

type database interface {
	GetFollowers(userUUID *binaryuuid.UUID, offset int, limit int) (*[]string, error)
	GetFollowings(userUUID *binaryuuid.UUID, offset int, limit int) (*[]string, error)
	RequestFollow(claimer *claimer.Claimer, targetUUID *binaryuuid.UUID) error
	IsFollower(claimer *claimer.Claimer, targetUUID *binaryuuid.UUID) (bool, error)
	IsFollowing(claimer *claimer.Claimer, targetUUID *binaryuuid.UUID) (bool, error)
	AcceptRequestFollow(claimer *claimer.Claimer, code *binaryuuid.UUID) error
	RejectRequestFollow(claimer *claimer.Claimer, code *binaryuuid.UUID) error
	RemoveFollower(claimer *claimer.Claimer, targetUUID *binaryuuid.UUID) error
	RemoveFollowing(claimer *claimer.Claimer, targetUUID *binaryuuid.UUID) error
}

type SocialAPI struct {
	DB database
}

func New(database database) *SocialAPI {
	return &SocialAPI{
		DB: database,
	}
}

type getFollowersHandlerUri struct {
	Target string `uri:"target" binding:"required,uri"`
}

type getFollowersHandlerForm struct {
	Offset *int `form:"offset,default=0" binding:"min=0"`
	Limit  *int `form:"limit,default=64" binding:"min=1,max=64"`
}

func (a *SocialAPI) GetFollowersHandler(ctx *gin.Context) {
	uri := &getFollowersHandlerUri{}
	if err := ctx.BindUri(uri); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "bind uri", err)
		return
	}
	form := &getFollowersHandlerForm{}
	if err := ctx.Bind(form); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "bind form", err)
		return
	}

	userUUID := binaryuuid.MustParse(uri.Target)
	followernames, err := a.DB.GetFollowers(&userUUID, *form.Offset, *form.Limit)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "get followers", err)
		return
	}
	ctx.JSON(http.StatusOK, followernames)
}

type getFollowingsHandlerUri = getFollowersHandlerUri
type getFollowingsHandlerForm = getFollowersHandlerForm

func (a *SocialAPI) GetFollowingsHandler(ctx *gin.Context) {
	uri := &getFollowingsHandlerUri{}
	if err := ctx.BindUri(uri); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "bind uri", err)
		return
	}
	form := &getFollowingsHandlerForm{}
	if err := ctx.Bind(form); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "bind form", err)
		return
	}

	targetUUID := binaryuuid.MustParse(uri.Target)
	followingnames, err := a.DB.GetFollowings(&targetUUID, *form.Offset, *form.Limit)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "get followings", err)
		return
	}
	ctx.JSON(http.StatusOK, followingnames)
}

type requestFollowHandlerUri struct {
	TargetUUID string `uri:"target" binding:"required, uuid"`
}

func (a *SocialAPI) RequestFollowHandler(ctx *gin.Context) {
	uri := &requestFollowHandlerUri{}
	if err := ctx.BindUri(uri); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "bind uri", err)
		return
	}

	targetUUID := binaryuuid.MustParse(uri.TargetUUID)
	claimer := api.MustGetClaimer(ctx)
	if err := a.DB.RequestFollow(claimer, &targetUUID); err != nil {
		ctx.Status(http.StatusInternalServerError)
		logger.ErrorWithCTX(ctx, "request follow", err)
		return
	}
	ctx.Status(http.StatusAccepted)
}

type getFollowerRelationHandlerUri struct {
	TargetUUID string `uri:"target" binding:"required,uuid"`
}

func (a *SocialAPI) GetFollowerRelationHandler(ctx *gin.Context) {
	uri := &getFollowerRelationHandlerUri{}
	if err := ctx.Bind(uri); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "binding uri", err)
		return
	}

	targetUUID := binaryuuid.MustParse(uri.TargetUUID)
	claimer := api.MustGetClaimer(ctx)
	isFollower, err := a.DB.IsFollower(claimer, &targetUUID)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		logger.ErrorWithCTX(ctx, "query follower", err)
		return
	}
	ctx.String(http.StatusOK, strconv.FormatBool(isFollower))
}

type getFollowingRelationHandlerUri = getFollowerRelationHandlerUri

func (a *SocialAPI) GetFollowingRelationHandler(ctx *gin.Context) {
	uri := &getFollowingRelationHandlerUri{}
	if err := ctx.BindUri(uri); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "binding uri", err)
		return
	}

	targetUUID := binaryuuid.MustParse(uri.TargetUUID)
	claimer := api.MustGetClaimer(ctx)
	isFollowing, err := a.DB.IsFollowing(claimer, &targetUUID)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		logger.ErrorWithCTX(ctx, "query following", err)
		return
	}
	ctx.String(http.StatusOK, strconv.FormatBool(isFollowing))
}

type deleteFollowerUri struct {
	TargetUUID string `uri:"target" binding:"required,uuid"`
}

func (a *SocialAPI) DeleteFollowerHandler(ctx *gin.Context) {
	uri := &deleteFollowerUri{}
	if err := ctx.BindUri(uri); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "bind uri", err)
		return
	}

	targetUUID := binaryuuid.MustParse(uri.TargetUUID)
	claimer := api.MustGetClaimer(ctx)
	if err := a.DB.RemoveFollower(claimer, &targetUUID); err != nil {
		ctx.Status(http.StatusNotAcceptable)
		logger.ErrorWithCTX(ctx, "remove follower", err)
		return
	}
	ctx.Status(http.StatusAccepted)
}

type deleteFollowingUri = deleteFollowerUri

func (a *SocialAPI) DeleteFollowingHandler(ctx *gin.Context) {
	uri := &deleteFollowingUri{}
	if err := ctx.BindUri(uri); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "bind uri", err)
		return
	}

	targetUUID := binaryuuid.MustParse(uri.TargetUUID)
	claimer := api.MustGetClaimer(ctx)
	if err := a.DB.RemoveFollowing(claimer, &targetUUID); err != nil {
		ctx.Status(http.StatusNotAcceptable)
		logger.ErrorWithCTX(ctx, "remove following", err)
		return
	}
	ctx.Status(http.StatusAccepted)
}

type acceptRequestFollowRequestUri struct {
	Code string `uri:"code" binding:"required,uuid"`
}

func (a *SocialAPI) AcceptRequestFollowHandler(ctx *gin.Context) {
	uri := &acceptRequestFollowRequestUri{}
	if err := ctx.BindUri(uri); err != nil {
		ctx.Status(http.StatusBadGateway)
		logger.ErrorWithCTX(ctx, "bind uri", err)
		return
	}

	codeUUID := binaryuuid.MustParse(uri.Code)
	claimer := api.MustGetClaimer(ctx)
	if err := a.DB.AcceptRequestFollow(claimer, &codeUUID); err != nil {
		ctx.Status(http.StatusBadGateway)
		logger.ErrorWithCTX(ctx, "accept request follow", err)
		return
	}

	ctx.Status(http.StatusAccepted)
}

type rejectRequestFollowRequestUri = acceptRequestFollowRequestUri

func (a *SocialAPI) RejectRequestFollowHandler(ctx *gin.Context) {
	uri := &rejectRequestFollowRequestUri{}
	if err := ctx.BindUri(uri); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "bind uri", err)
		return
	}

	codeUUID := binaryuuid.MustParse(uri.Code)
	claimer := api.MustGetClaimer(ctx)
	if err := a.DB.RejectRequestFollow(claimer, &codeUUID); err != nil {
		ctx.Status(http.StatusBadGateway)
		logger.ErrorWithCTX(ctx, "reject request follow", err)
		return
	}
	ctx.Status(http.StatusNoContent)
}
