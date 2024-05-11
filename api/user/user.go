package userAPI

import (
	"net/http"

	"github.com/capdale/was/api"
	baselogger "github.com/capdale/was/logger"
	"github.com/capdale/was/types/claimer"
	"github.com/gin-gonic/gin"
)

var logger = baselogger.Logger

type database interface {
	UserVisibilityPublic() int
	UserVisibilityPrivate() int
	ChangeVisibility(claimer *claimer.Claimer, visibilityType int) error
}

type UserAPI struct {
	d database
}

func New(database database) *UserAPI {
	return &UserAPI{
		d: database,
	}
}

type changeVisibilityUri struct {
	Type string `uri:"type" binding:"oneof=public private"`
}

func (a *UserAPI) ChangeVisibilityHandler(ctx *gin.Context) {
	uri := &changeVisibilityUri{}
	if err := ctx.BindUri(uri); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "bind uri", err)
		return
	}

	claimer := api.MustGetClaimer(ctx)

	changeType := a.d.UserVisibilityPrivate()
	// don't check error state, bind check it already
	if uri.Type == "public" {
		changeType = a.d.UserVisibilityPublic()
	}

	if err := a.d.ChangeVisibility(claimer, changeType); err != nil {
		ctx.Status(http.StatusInternalServerError)
		logger.ErrorWithCTX(ctx, "change visibility", err)
		return
	}
	ctx.Status(http.StatusAccepted)
}
