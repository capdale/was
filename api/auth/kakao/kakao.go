package kakao

import (
	"encoding/base64"
	"math/rand"
	"net/http"
	"time"

	"github.com/capdale/was/api"
	authapi "github.com/capdale/was/api/auth"
	"github.com/capdale/was/auth"
	baseLogger "github.com/capdale/was/logger"
	"github.com/capdale/was/model"
	"github.com/capdale/was/types/claimer"
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

type state interface {
	SetState(state string, expired time.Duration) error
	PopState(state string) error
}

type KakaoAuth struct {
	DB          database
	Auth        *auth.Auth
	State       state
	OAuthConfig *oauth2.Config
}

func New(database database, auth *auth.Auth, state state, oauthConfig *oauth2.Config) *KakaoAuth {
	return &KakaoAuth{
		DB:          database,
		Auth:        auth,
		State:       state,
		OAuthConfig: oauthConfig,
	}
}

func (k *KakaoAuth) LoginHandler(ctx *gin.Context) {
	rand32, err := auth.RandToken(32)
	if err != nil {
		api.BasicInternalServerError(ctx)
		logger.ErrorWithCTX(ctx, "generate random token", err)
		return
	}
	state := base64.StdEncoding.EncodeToString(*rand32)
	k.State.SetState(state, time.Minute*10)
	if ctx.Query("type") == "json" {
		ctx.JSON(http.StatusOK, gin.H{"url": k.OAuthConfig.AuthCodeURL(state)})
		return
	}
	ctx.Redirect(http.StatusFound, k.OAuthConfig.AuthCodeURL(state))
}

type loginWithAccessTokenForm struct {
	AccessToken string `form:"access_token"`
}

func (k *KakaoAuth) LoginWithAccessTokenHandler(ctx *gin.Context) {
	form := &loginWithAccessTokenForm{}
	if err := ctx.Bind(form); err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "bind form", err)
		return
	}

	email, err := getUserEmailWithAccessToken(form.AccessToken)
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		logger.ErrorWithCTX(ctx, "get user email", err)
		return
	}

	user, err := k.getUserFromSocialByEmail(email)

	if err == gorm.ErrRecordNotFound {
		user, err = k.DB.CreateWithKakao(RandStringRunes(8), email)
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

func (k *KakaoAuth) CallbackHandler(ctx *gin.Context) {
	state := ctx.Query("state")
	err := k.State.PopState(state)
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
		user, err = k.DB.CreateWithKakao(RandStringRunes(8), email)
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

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz1234567890")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
