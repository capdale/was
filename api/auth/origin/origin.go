package originAPI

import (
	"fmt"
	"net/http"

	"github.com/capdale/was/auth"
	baselogger "github.com/capdale/was/logger"
	"github.com/capdale/was/types/binaryuuid"
	"github.com/gin-gonic/gin"
)

var logger = baselogger.Logger

type database interface {
	IsEmailUsed(email string) (bool, error)
	IsTicketAvailable(ticketUUID binaryuuid.UUID) (bool, error)
	CreateTicketByEmail(email string) (*binaryuuid.UUID, error)
	CreateOriginViaTicket(ticket binaryuuid.UUID, username string, password string) error
	GetOriginUserUUID(username string, password string) (*binaryuuid.UUID, error)
}

type OriginAPI struct {
	DB   database
	Auth *auth.Auth
}

func New(d database, auth *auth.Auth) *OriginAPI {
	return &OriginAPI{
		DB:   d,
		Auth: auth,
	}
}

type createEmailTicketForm struct {
	Email string `form:"email" json:"email" binding:"required,email"`
}

func (o *OriginAPI) CreateEmailTicketHandler(ctx *gin.Context) {
	form := &createEmailTicketForm{}
	err := ctx.ShouldBind(form)
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "binding error", err)
		return
	}

	ticketUUID, err := o.DB.CreateTicketByEmail(form.Email)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		logger.ErrorWithCTX(ctx, "create ticket error", err)
		return
	}
	fmt.Println(ticketUUID)
	ctx.Status(http.StatusAccepted)
}

type registerTicketForm struct {
	Ticket   string `form:"ticket" binding:"required,uuid"`
	Username string `form:"username" json:"username" binding:"required,min=6,max=12"`
	Password string `form:"password" json:"password" binding:"required,min=8,max=22"`
}

func (o *OriginAPI) RegisterTicketHandler(ctx *gin.Context) {
	form := &registerTicketForm{}
	err := ctx.ShouldBind(form)
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "binding error", err)
		return
	}

	ticketUUID := binaryuuid.MustParse(form.Ticket)

	err = o.DB.CreateOriginViaTicket(ticketUUID, form.Username, form.Password)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "create origin via ticket", err)
		return
	}
	ctx.Status(http.StatusAccepted)
}

type loginForm struct {
	Username string `json:"username" binding:"required,min=6,max=12"`
	Password string `json:"password" binding:"required,min=8,max=22"`
}

func (o *OriginAPI) LoginHandler(ctx *gin.Context) {
	form := &loginForm{}
	if err := ctx.ShouldBind(form); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "binding error", err)
		return
	}

	userUUID, err := o.DB.GetOriginUserUUID(form.Username, form.Password)
	if err != nil {
		ctx.Status(http.StatusUnauthorized)
		logger.ErrorWithCTX(ctx, "get origin user", err)
		return
	}

	userAgent := ctx.Request.UserAgent()
	tokenString, refreshToken, err := o.Auth.IssueToken(*userUUID, &userAgent)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		logger.ErrorWithCTX(ctx, "issue token", err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  tokenString,
		"refresh_token": refreshToken,
	})
}
