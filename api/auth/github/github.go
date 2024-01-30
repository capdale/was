package githubAuth

import (
	"encoding/base64"
	"net/http"

	authapi "github.com/capdale/was/api/auth"
	"github.com/capdale/was/auth"
	"github.com/capdale/was/model"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

const (
	emailInfoEndpoint = "https://api.github.com/user/emails"
	userInfoEndpoint  = "https://api.github.com/user"
)

type database interface {
	GetUserByEmail(email string) (*model.User, error)
	CreateWithGithub(username string, email string) (*model.User, error)
}

type GithubAuth struct {
	DB          database
	Auth        *auth.Auth
	OAuthConfig *oauth2.Config
}

func New(database database, auth *auth.Auth, oauthConfig *oauth2.Config) *GithubAuth {
	return &GithubAuth{
		DB:          database,
		Auth:        auth,
		OAuthConfig: oauthConfig,
	}
}

func (g *GithubAuth) LoginHandler(ctx *gin.Context) {

	session := sessions.Default(ctx)
	session.Options(sessions.Options{
		Path:   "/auth",
		MaxAge: 900,
	})
	state := auth.RandToken(32)
	session.Set("state", state)
	session.Save()
	if ctx.Query("type") == "json" {
		ctx.JSON(http.StatusOK, gin.H{"url": g.OAuthConfig.AuthCodeURL(state)})
		return
	}
	ctx.Redirect(http.StatusFound, g.OAuthConfig.AuthCodeURL(state))
}

func (g *GithubAuth) CallbackHandler(ctx *gin.Context) {
	err := authapi.CheckState(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}

	token, err := g.OAuthConfig.Exchange(ctx.Request.Context(), ctx.Query("code"))
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}
	if !token.Valid() {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}

	userId, err := g.getUserId(ctx, token)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	userEmail, err := g.getEmail(ctx, token)
	if err == authapi.ErrNoValidEmail {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	} else if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	user, err := g.getUserFromSocialByEmail(userEmail)

	if err == gorm.ErrRecordNotFound {
		user, err = g.createSocial(userId, userEmail)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}
	} else if err == authapi.ErrAlreayExistEmail {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	} else if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	userAgent := ctx.Request.UserAgent()

	tokenString, refreshToken, err := g.Auth.IssueTokenByUUID(&user.UUID, &userAgent)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  tokenString,
		"refresh_token": base64.StdEncoding.EncodeToString(*refreshToken),
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
	if user.AccountType != 0 {
		return user, authapi.ErrAlreayExistEmail
	}
	return
}
