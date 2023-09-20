package server

import (
	"net/http"
	"time"

	authapi "github.com/capdale/was/api/auth"
	githubAuth "github.com/capdale/was/api/auth/github"
	"github.com/capdale/was/api/collect"
	"github.com/capdale/was/auth"
	"github.com/capdale/was/config"
	"github.com/capdale/was/database"
	"github.com/capdale/was/store"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

func SetupRouter() (r *gin.Engine) {
	r = gin.Default()

	// config := cors.DefaultConfig()

	r.Use(cors.New(
		cors.Config{
			AllowOrigins: []string{"http://localhost:5500"}, // test front addr
			AllowMethods: []string{"POST", "GET", "DELETE", "OPTIONS", "PUT"},
			AllowHeaders: []string{
				"Origin", "Content-Type", "Upgrade",
				"X-MD-Token", "Accept-Encoding", "Accept-Language",
				"Authorization", "Host"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		},
	))

	config, err := config.ParseConfig("config.yaml")
	if err != nil {
		panic(err)
	}

	d, err := database.New(config)
	if err != nil {
		panic(err)
	}

	store, err := store.New(&config.Redis)
	if err != nil {
		panic(err)
	}

	st, err := redis.NewStore(10, "tcp", config.Redis.Address, config.Redis.Password, []byte("secret"))
	if err != nil {
		panic(err)
	}
	r.Use(sessions.Sessions("authstate", st))

	auth := auth.New(d, store)

	r.GET("/", func(ctx *gin.Context) {
		ctx.AbortWithStatusJSON(http.StatusOK, gin.H{
			"ok": "ok",
		})
	})

	collectAPI := collect.New(d)

	collectRouter := r.Group("/collect").Use(auth.AuthorizeRequiredMiddleware())
	{
		collectRouter.GET("/", collectAPI.GetCollectection)
		collectRouter.POST("/", collectAPI.PostCollectHandler)
	}

	authAPI := authapi.New(d, auth)
	authRouter := r.Group("/auth")
	{
		r.Use(auth.AuthorizeRequiredMiddleware()).POST("/blacklist", authAPI.SetBlacklistHandler)

		githubAuth := githubAuth.New(d, auth, &oauth2.Config{
			ClientID:     config.Oauth.Github.Id,
			ClientSecret: config.Oauth.Github.Secret,
			RedirectURL:  config.Oauth.Github.Redirect,
			Scopes:       []string{"user:email"},
			Endpoint:     github.Endpoint,
		})
		githubAuthRouter := authRouter.Group("/github")
		{
			githubAuthRouter.GET("/login", githubAuth.LoginHandler)
			githubAuthRouter.GET("/callback", githubAuth.CallbackHandler)
		}
	}

	return
}
