package githubAuth

import (
	"encoding/base64"
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

var logger = baseLogger.Logger

const (
	emailInfoEndpoint = "https://api.github.com/user/emails"
	userInfoEndpoint  = "https://api.github.com/user"
)

type database interface {
	GetUserByEmail(email string) (*model.User, error)
	CreateWithGithub(username string, email string) (*model.User, error)
}

type state interface {
	SetState(state string, expired time.Duration) error
	PopState(state string) error
}

type GithubAuth struct {
	DB          database
	Auth        *auth.Auth
	State       state
	OAuthConfig *oauth2.Config
}

func New(database database, auth *auth.Auth, state state, oauthConfig *oauth2.Config) *GithubAuth {
	return &GithubAuth{
		DB:          database,
		Auth:        auth,
		State:       state,
		OAuthConfig: oauthConfig,
	}
}

func (g *GithubAuth) LoginHandler(ctx *gin.Context) {
	rand32, err := auth.RandToken(32)
	if err != nil {
		api.BasicInternalServerError(ctx)
		logger.ErrorWithCTX(ctx, "generate random token", err)
		return
	}
	state := base64.StdEncoding.EncodeToString(*rand32)
	g.State.SetState(state, time.Minute*10)
	if ctx.Query("type") == "json" {
		ctx.JSON(http.StatusOK, gin.H{"url": g.OAuthConfig.AuthCodeURL(state)})
		return
	}
	ctx.Redirect(http.StatusFound, g.OAuthConfig.AuthCodeURL(state))
}

func (g *GithubAuth) CallbackHandler(ctx *gin.Context) {
	state := ctx.Query("state")
	err := g.State.PopState(state)
	if err != nil {
		api.BasicUnAuthorizedError(ctx)
		logger.ErrorWithCTX(ctx, "cannot find state", err)
		return
	}

	token, err := g.OAuthConfig.Exchange(ctx.Request.Context(), ctx.Query("code"))
	if err != nil {
		api.BasicUnAuthorizedError(ctx)
		logger.ErrorWithCTX(ctx, "exchane oauth failed", err)
	}
	if !token.Valid() {
		api.BasicUnAuthorizedError(ctx)
		logger.ErrorWithCTX(ctx, "token invalid", nil)
	}

	userId, err := g.getUserId(ctx, token)
	if err != nil {
		api.BasicInternalServerError(ctx)
		logger.ErrorWithCTX(ctx, "get user id failed", err)
		return
	}

	userEmail, err := g.getEmail(ctx, token)
	if err == authapi.ErrNoValidEmail {
		api.BasicInternalServerError(ctx)
		logger.ErrorWithCTX(ctx, "no valid email", err)
		return
	} else if err != nil {
		api.BasicInternalServerError(ctx)
		logger.ErrorWithCTX(ctx, "get email failed", err)
		return
	}

	user, err := g.getUserFromSocialByEmail(userEmail)

	if err == gorm.ErrRecordNotFound {
		user, err = g.createSocial(userId, userEmail)
		if err != nil {
			api.BasicInternalServerError(ctx)
			logger.ErrorWithCTX(ctx, "create social account", err)
			return
		}
	} else if err == authapi.ErrAlreayExistEmail {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "email already used"})
		logger.ErrorWithCTX(ctx, "try to create duplicated email account", err)
		return
	} else if err != nil {
		api.BasicInternalServerError(ctx)
		logger.ErrorWithCTX(ctx, "query social account by email", err)
		return
	}

	userAgent := ctx.Request.UserAgent()
	claimer := claimer.New(&user.AuthUUID)
	tokenString, refreshToken, err := g.Auth.IssueToken(*claimer, &userAgent)
	if err != nil {
		api.BasicInternalServerError(ctx)
		logger.ErrorWithCTX(ctx, "issue token", err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"username":      user.Username,
		"access_token":  tokenString,
		"refresh_token": refreshToken,
	})
}

func (g *GithubAuth) createSocial(username string, email string) (user *model.User, err error) {
	user, err = g.DB.CreateWithGithub(username, email)
	return
}

func (g *GithubAuth) getUserFromSocialByEmail(email string) (user *model.User, err error) {
	user, err = g.DB.GetUserByEmail(email)
	if err != nil {
		return
	}
	if user.AccountType != model.AccountTypeGithub {
		return user, authapi.ErrAlreayExistEmail
	}
	return
}
