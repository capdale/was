package kakao

import (
	"encoding/base64"
	"net/http"

	"github.com/capdale/was/api"
	authapi "github.com/capdale/was/api/auth"
	"github.com/capdale/was/auth"
	baseLogger "github.com/capdale/was/logger"
	"github.com/capdale/was/model"
	"github.com/capdale/was/types/claimer"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

const userInfoEndpoint = "https://kapi.kakao.com/v2/user/me"

var logger = baseLogger.Logger

type database interface {
	GetUserByEmail(email string) (*model.User, error)
	CreateWithKakao(username string, email string) (*model.User, error)
}

type KakaoAuth struct {
	DB          database
	Auth        *auth.Auth
	OAuthConfig *oauth2.Config
}

func New(database database, auth *auth.Auth, oauthConfig *oauth2.Config) *KakaoAuth {
	return &KakaoAuth{
		DB:          database,
		Auth:        auth,
		OAuthConfig: oauthConfig,
	}
}

func (k *KakaoAuth) LoginHandler(ctx *gin.Context) {
	session := sessions.Default(ctx)
	session.Options(sessions.Options{
		Path:   "/auth",
		MaxAge: 900,
	})
	rand32, err := auth.RandToken(32)
	if err != nil {
		api.BasicInternalServerError(ctx)
		logger.ErrorWithCTX(ctx, "generate random token", err)
		return
	}
	state := base64.StdEncoding.EncodeToString(*rand32)
	session.Set("state", state)
	session.Save()
	if ctx.Query("type") == "json" {
		ctx.JSON(http.StatusOK, gin.H{"url": k.OAuthConfig.AuthCodeURL(state)})
		return
	}
	ctx.Redirect(http.StatusFound, k.OAuthConfig.AuthCodeURL(state))
}

func (k *KakaoAuth) CallbackHandler(ctx *gin.Context) {
	err := authapi.CheckState(ctx)
	if err != nil {
		api.BasicUnAuthorizedError(ctx)
		logger.ErrorWithCTX(ctx, "cannot find state", err)
		return
	}

	token, err := k.OAuthConfig.Exchange(ctx.Request.Context(), ctx.Query("code"))
	if err != nil {
		api.BasicUnAuthorizedError(ctx)
		logger.ErrorWithCTX(ctx, "exchane oauth failed", err)
	}

	if !token.Valid() {
		api.BasicUnAuthorizedError(ctx)
		logger.ErrorWithCTX(ctx, "token invalid", nil)
	}

	email, err := k.getUserEmail(ctx, token)
	if err != nil {
		api.BasicBadRequestError(ctx)
		logger.ErrorWithCTX(ctx, "get user email from kakao error", err)
		return
	}

	user, err := k.getUserFromSocialByEmail(email)

	if err == gorm.ErrRecordNotFound {
		user, err = k.DB.CreateWithKakao("username", email)
		if err != nil {
			api.BasicInternalServerError(ctx)
			logger.ErrorWithCTX(ctx, "create kakao account", err)
			return
		}
	} else if err == authapi.ErrAlreayExistEmail {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "email already used"})
		logger.ErrorWithCTX(ctx, "email alreay exist", err)
		return
	} else if err != nil {
		api.BasicInternalServerError(ctx)
		logger.ErrorWithCTX(ctx, "query social account by email", err)
		return
	}

	userAgent := ctx.Request.UserAgent()
	claimer := claimer.New(&user.AuthUUID)
	tokenString, refreshToken, err := k.Auth.IssueToken(*claimer, &userAgent)
	if err != nil {
		api.BasicInternalServerError(ctx)
		logger.ErrorWithCTX(ctx, "issue token", err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  tokenString,
		"refresh_token": refreshToken,
	})
}

func (k *KakaoAuth) getUserFromSocialByEmail(email string) (user *model.User, err error) {
	user, err = k.DB.GetUserByEmail(email)
	if err != nil {
		return
	}
	if user.AccountType != model.AccountTypeKakao {
		return user, authapi.ErrAlreayExistEmail
	}
	return
}
