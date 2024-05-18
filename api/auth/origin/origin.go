package originAPI

import (
	"fmt"
	"net/http"

	"github.com/capdale/was/auth"
	"github.com/capdale/was/email"
	baselogger "github.com/capdale/was/logger"
	"github.com/capdale/was/types/binaryuuid"
	"github.com/capdale/was/types/claimer"
	"github.com/gin-gonic/gin"
)

var logger = baselogger.Logger

type database interface {
	IsEmailUsed(email string) (bool, error)
	IsTicketAvailable(ticketUUID *binaryuuid.UUID) (bool, error)
	CreateTicketByEmail(email string) (*binaryuuid.UUID, error)
	GetEmailByTicket(ticketUUID *binaryuuid.UUID) (string, error)
	CreateOriginViaTicket(ticket *binaryuuid.UUID, username string, password string) error
	GetOriginUserClaim(username string, password string) (*claimer.Claimer, error)
}

type OriginAPI struct {
	DB               database
	Auth             *auth.Auth
	Email            email.EmailService
	CreateVerifyLink func(identifier string) string
}

func New(d database, auth *auth.Auth, email email.EmailService, createVerifyLink func(string) string) *OriginAPI {
	return &OriginAPI{
		DB:               d,
		Auth:             auth,
		Email:            email,
		CreateVerifyLink: createVerifyLink,
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

	verifyLink := o.CreateVerifyLink(ticketUUID.String())
	if err := o.Email.SendTicketVerifyLink(ctx, form.Email, verifyLink); err != nil {
		ctx.Status(http.StatusInternalServerError)
		logger.ErrorWithCTX(ctx, "create email error", err)
		return
	}
	ctx.Status(http.StatusAccepted)
}

type registerTicketForm struct {
	Ticket   string `form:"ticket" binding:"required,uuid"`
	Username string `form:"username" json:"username" binding:"required,min=6,max=20"`
	Password string `form:"password" json:"password" binding:"required,min=8,max=32"`
}

func (o *OriginAPI) RegisterTicketHandler(ctx *gin.Context) {
	form := &registerTicketForm{}
	err := ctx.ShouldBind(form)
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "binding error", err)
		return
	}

	if !validatePassword(&form.Password) {
		err = ErrInvalidPasswordForm
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "password check error", err)
		return
	}

	ticketUUID := binaryuuid.MustParse(form.Ticket)

	err = o.DB.CreateOriginViaTicket(&ticketUUID, form.Username, form.Password)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "create origin via ticket", err)
		return
	}
	ctx.Status(http.StatusAccepted)
}

type registerTicketViewUri struct {
	TicketUUID string `uri:"ticket" binding:"required,uuid"`
}

func (o *OriginAPI) RegisterTicketView(ctx *gin.Context) {
	uri := &registerTicketViewUri{}
	if err := ctx.BindUri(uri); err != nil {
		// TODO: change to 404 page
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "bind uri", err)
		return
	}

	ticketUUID := binaryuuid.MustParse(uri.TicketUUID)

	ticketEmail, err := o.DB.GetEmailByTicket(&ticketUUID)
	if err != nil {
		// TODO: change to 404 page
		ctx.Status(http.StatusNotFound)
		logger.ErrorWithCTX(ctx, "bind uri", err)
		return
	}

	fmt.Println(ticketEmail)
	censoredEmail := email.CensorEmail(ticketEmail)
	ctx.HTML(http.StatusOK, "email_register.tmpl", gin.H{
		"endpoint": "/auth/regist",
		"ticket":   ticketUUID,
		"email":    censoredEmail,
	})
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

	claimer, err := o.DB.GetOriginUserClaim(form.Username, form.Password)
	if err != nil {
		ctx.Status(http.StatusUnauthorized)
		logger.ErrorWithCTX(ctx, "get origin user", err)
		return
	}

	userAgent := ctx.Request.UserAgent()
	tokenString, refreshToken, err := o.Auth.IssueToken(*claimer, &userAgent)
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
