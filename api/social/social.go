package socialAPI

import (
	"net/http"
	"strconv"

	"github.com/capdale/was/api"
	baselogger "github.com/capdale/was/logger"
	"github.com/capdale/was/model"

	"github.com/capdale/was/types/binaryuuid"
	"github.com/capdale/was/types/claimer"
	"github.com/gin-gonic/gin"
)

var logger = baselogger.Logger

type database interface {
	GetFollowers(claimer *claimer.Claimer, username *string, offset int, limit int) (*[]string, error)
	GetFollowings(claimer *claimer.Claimer, username *string, offset int, limit int) (*[]string, error)
	RequestFollow(claimer *claimer.Claimer, targetname *string) error
	IsFollower(claimer *claimer.Claimer, targetname *string) (bool, error)
	IsFollowing(claimer *claimer.Claimer, targetname *string) (bool, error)
	AcceptRequestFollow(claimer *claimer.Claimer, code *binaryuuid.UUID) error
	RejectRequestFollow(claimer *claimer.Claimer, code *binaryuuid.UUID) error
	GetFollowRequests(claimer *claimer.Claimer, offset int, limit int) (*[]model.FollowRequest, error)
	RemoveFollower(claimer *claimer.Claimer, targetname *string) error
	RemoveFollowing(claimer *claimer.Claimer, targetUUID *string) error
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
	Targetname string `uri:"targetname" binding:"required"`
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

	claimer := api.GetClaimer(ctx)
	followers, err := a.DB.GetFollowers(claimer, &uri.Targetname, *form.Offset, *form.Limit)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "get followers", err)
		return
	}
	ctx.JSON(http.StatusOK, &gin.H{
		"followers": followers,
	})
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

	claimer := api.GetClaimer(ctx)
	followings, err := a.DB.GetFollowings(claimer, &uri.Targetname, *form.Offset, *form.Limit)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "get followings", err)
		return
	}
	ctx.JSON(http.StatusOK, &gin.H{
		"followings": followings,
	})
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

	claimer := api.MustGetClaimer(ctx)
	if err := a.DB.RequestFollow(claimer, &uri.TargetUUID); err != nil {
		ctx.Status(http.StatusInternalServerError)
		logger.ErrorWithCTX(ctx, "request follow", err)
		return
	}
	ctx.Status(http.StatusAccepted)
}

type getFollowerRelationHandlerUri struct {
	Targetname string `uri:"target" binding:"required"`
}

func (a *SocialAPI) GetFollowerRelationHandler(ctx *gin.Context) {
	uri := &getFollowerRelationHandlerUri{}
	if err := ctx.Bind(uri); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "binding uri", err)
		return
	}

	claimer := api.MustGetClaimer(ctx)

	isFollower, err := a.DB.IsFollower(claimer, &uri.Targetname)
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

	claimer := api.MustGetClaimer(ctx)
	isFollowing, err := a.DB.IsFollowing(claimer, &uri.Targetname)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		logger.ErrorWithCTX(ctx, "query following", err)
		return
	}
	ctx.String(http.StatusOK, strconv.FormatBool(isFollowing))
}

type deleteFollowerUri struct {
	Targetname string `uri:"targetname" binding:"required"`
}

func (a *SocialAPI) DeleteFollowerHandler(ctx *gin.Context) {
	uri := &deleteFollowerUri{}
	if err := ctx.BindUri(uri); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "bind uri", err)
		return
	}

	claimer := api.MustGetClaimer(ctx)
	if err := a.DB.RemoveFollower(claimer, &uri.Targetname); err != nil {
		ctx.Status(http.StatusNotAcceptable)
		logger.ErrorWithCTX(ctx, "remove follower", err)
		return
	}
	ctx.Status(http.StatusAccepted)
}

// duplicated, but remain (different function)
type getRelationHandlerUri struct {
	Targetname string `uri:"targetname" binding:"required"`
}

func (a *SocialAPI) GetRelationHandler(ctx *gin.Context) {
	uri := &getRelationHandlerUri{}
	if err := ctx.Bind(uri); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "binding uri", err)
		return
	}

	claimer := api.MustGetClaimer(ctx)

	// [TODO] integrate to database.GetRelationship for resource
	isFollower, err := a.DB.IsFollower(claimer, &uri.Targetname)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		logger.ErrorWithCTX(ctx, "query follower", err)
		return
	}

	isFollowing, err := a.DB.IsFollowing(claimer, &uri.Targetname)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		logger.ErrorWithCTX(ctx, "query follower", err)
		return
	}

	ctx.JSON(http.StatusOK, &gin.H{
		"is_following": isFollowing,
		"is_follower":  isFollower,
	})
}

type deleteFollowingUri = deleteFollowerUri

func (a *SocialAPI) DeleteFollowingHandler(ctx *gin.Context) {
	uri := &deleteFollowingUri{}
	if err := ctx.BindUri(uri); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "bind uri", err)
		return
	}

	claimer := api.MustGetClaimer(ctx)
	if err := a.DB.RemoveFollowing(claimer, &uri.Targetname); err != nil {
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

	code := binaryuuid.MustParse(uri.Code)
	claimer := api.MustGetClaimer(ctx)
	if err := a.DB.AcceptRequestFollow(claimer, &code); err != nil {
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

	code := binaryuuid.MustParse(uri.Code)
	claimer := api.MustGetClaimer(ctx)
	if err := a.DB.RejectRequestFollow(claimer, &code); err != nil {
		ctx.Status(http.StatusBadGateway)
		logger.ErrorWithCTX(ctx, "reject request follow", err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

type getFollowRequestsHandlerForm struct {
	Offset *int `form:"offset,default=0" binding:"min=0"`
	Limit  *int `form:"limit,default=64" binding:"min=1,max=64"`
}

func (a *SocialAPI) GetFollowRequestsHandler(ctx *gin.Context) {
	form := &getFollowRequestsHandlerForm{}
	if err := ctx.Bind(form); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "bind form", err)
		return
	}

	claimer := api.MustGetClaimer(ctx)
	requests, err := a.DB.GetFollowRequests(claimer, *form.Offset, *form.Limit)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		logger.ErrorWithCTX(ctx, "get follow request", err)
		return
	}

	ctx.JSON(http.StatusOK, &gin.H{
		"requests": requests,
	})

}
