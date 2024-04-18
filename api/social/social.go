package socialAPI

import (
	"net/http"
	"strconv"

	"github.com/capdale/was/auth"
	baselogger "github.com/capdale/was/logger"

	"github.com/capdale/was/types/binaryuuid"
	"github.com/gin-gonic/gin"
)

var logger = baselogger.Logger

type database interface {
	GetFollowers(username string, offset int, limit int) (*[]string, error)
	GetFollowings(username string, offset int, limit int) (*[]string, error)
	RequestFollow(claimer binaryuuid.UUID, targetname string) error
	IsFollower(claimerUUID binaryuuid.UUID, targetname string) (bool, error)
	IsFollowing(claimerUUID binaryuuid.UUID, targetname string) (bool, error)
	AcceptRequestFollow(claimerUUID *binaryuuid.UUID, requestUUID *binaryuuid.UUID) error
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
	Name string `uri:"username" binding:"required"`
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

	followernames, err := a.DB.GetFollowers(uri.Name, *form.Offset, *form.Limit)
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

	followingnames, err := a.DB.GetFollowings(uri.Name, *form.Offset, *form.Limit)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "get followings", err)
		return
	}
	ctx.JSON(http.StatusOK, followingnames)
}

type requestFollowHandlerUri struct {
	Targetname string `uri:"username" binding:"required"`
}

func (a *SocialAPI) RequestFollowHandler(ctx *gin.Context) {
	uri := &requestFollowHandlerUri{}
	if err := ctx.BindUri(uri); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "bind uri", err)
		return
	}

	claims := ctx.MustGet("claims").(*auth.AuthClaims)

	if err := a.DB.RequestFollow(claims.UUID, uri.Targetname); err != nil {
		ctx.Status(http.StatusInternalServerError)
		logger.ErrorWithCTX(ctx, "request follow", err)
		return
	}
	ctx.Status(http.StatusAccepted)
}

type getFollowerRelationHandlerUri struct {
	TargetName string `uri:"username" binding:"required"`
}

func (a *SocialAPI) GetFollowerRelationHandler(ctx *gin.Context) {
	claims := ctx.MustGet("claims").(*auth.AuthClaims)
	uri := &getFollowerRelationHandlerUri{}
	if err := ctx.Bind(uri); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "binding uri", err)
		return
	}

	isFollower, err := a.DB.IsFollower(claims.UUID, uri.TargetName)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		logger.ErrorWithCTX(ctx, "query follower", err)
		return
	}
	ctx.String(http.StatusOK, strconv.FormatBool(isFollower))
}

type getFollowingRelationHandlerUri = getFollowerRelationHandlerUri

func (a *SocialAPI) GetFollowingRelationHandler(ctx *gin.Context) {
	claims := ctx.MustGet("claims").(*auth.AuthClaims)
	uri := &getFollowingRelationHandlerUri{}
	if err := ctx.BindUri(uri); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "binding uri", err)
		return
	}

	isFollowing, err := a.DB.IsFollowing(claims.UUID, uri.TargetName)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		logger.ErrorWithCTX(ctx, "query following", err)
		return
	}
	ctx.String(http.StatusOK, strconv.FormatBool(isFollowing))
}

type acceptRequestFollowRequestUri struct {
	RequestUUID string `uri:"request_uuid" binding:"required,uuid"`
}

func (a *SocialAPI) AcceptRequestFollowHandler(ctx *gin.Context) {
	uri := &acceptRequestFollowRequestUri{}
	if err := ctx.BindUri(uri); err != nil {
		ctx.Status(http.StatusBadGateway)
		logger.ErrorWithCTX(ctx, "bind uri", err)
		return 
	}
	
	claims := ctx.MustGet("claims").(*auth.AuthClaims)
	requestUUID := binaryuuid.MustParse(uri.RequestUUID)
	
	if err := a.DB.AcceptRequestFollow(&claims.UUID, &requestUUID); err != nil {
		ctx.Status(http.StatusBadGateway)
		logger.ErrorWithCTX(ctx, "bind uri", err)
		return 
	}
	
	ctx.Status(http.StatusAccepted)
}