package server

import (
	"net/http"
	"time"

	authapi "github.com/capdale/was/api/auth"
	githubAuth "github.com/capdale/was/api/auth/github"
	collect "github.com/capdale/was/api/collection"
	"github.com/capdale/was/auth"
	"github.com/capdale/was/config"
	"github.com/capdale/was/database"
	"github.com/capdale/was/logger"
	"github.com/capdale/was/store"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"gopkg.in/natefinch/lumberjack.v2"
)

func SetupRouter(config *config.Config) (r *gin.Engine, err error) {
	r = gin.New()

	isProduction := gin.Mode() == gin.ReleaseMode

	routerLogger := logger.New(&lumberjack.Logger{
		Filename:   config.Service.Log.Path,
		MaxSize:    config.Service.Log.MaxSize,
		MaxBackups: config.Service.Log.MaxBackups,
		MaxAge:     config.Service.Log.MaxAge,
	}, isProduction, config.Service.Log.Console)

	r.Use(ginzap.Ginzap(routerLogger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(routerLogger, true))

	r.Use(cors.New(
		cors.Config{
			AllowOrigins:           config.Service.Cors.AllowOrigins,
			AllowMethods:           config.Service.Cors.AllowMethods,
			AllowHeaders:           config.Service.Cors.AllowHeaders,
			AllowCredentials:       config.Service.Cors.AllowCredentials,
			ExposeHeaders:          config.Service.Cors.ExposeHeaders,
			MaxAge:                 time.Duration(config.Service.Cors.MaxAge) * time.Second,
			AllowWildcard:          config.Service.Cors.AllowWildcard,
			AllowBrowserExtensions: config.Service.Cors.AllowBrowserExtensions,
			AllowWebSockets:        config.Service.Cors.AllowWebSockets,
			AllowFiles:             config.Service.Cors.AllowFiles,
		},
	))

	d, err := database.New(&config.Mysql)
	if err != nil {
		return
	}

	store, err := store.New(&config.Redis)
	if err != nil {
		return
	}

	st, err := redis.NewStore(10, "tcp", config.Redis.Address, config.Redis.Password, []byte(config.Key.SessionStateKey))
	if err != nil {
		return nil, err
	}

	// s3, err := s3.New(&config.S3)
	// if err != nil {
	// 	return
	// }
	// print(s3) // declare

	auth := auth.New(d, store)

	r.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"ok": "ok",
		})
	})

	collectAPI := collect.New(d)

	collectRouter := r.Group("/collection").Use(auth.AuthorizeRequiredMiddleware())
	{
		collectRouter.GET("/", auth.AuthorizeRequiredMiddleware(), collectAPI.GetCollectection)
		collectRouter.POST("/", auth.AuthorizeRequiredMiddleware(), collectAPI.PostCollection)
		collectRouter.GET("/:uuid", collectAPI.GetCollectionByUUID)
	}

	authAPI := authapi.New(d, auth)
	authRouter := r.Group("/auth")
	{
		authRouter.POST("/blacklist", auth.AuthorizeRequiredMiddleware(), authAPI.SetBlacklistHandler)
		authRouter.POST("/refresh", authAPI.RefreshTokenHandler)
		githubAuth := githubAuth.New(d, auth, &oauth2.Config{
			ClientID:     config.Oauth.Github.Id,
			ClientSecret: config.Oauth.Github.Secret,
			RedirectURL:  config.Oauth.Github.Redirect,
			Scopes:       []string{"user:email"},
			Endpoint:     github.Endpoint,
		})
		githubAuthRouter := authRouter.Group("/github").Use(sessions.Sessions("state", st))
		{
			githubAuthRouter.GET("/login", githubAuth.LoginHandler)
			githubAuthRouter.GET("/callback", githubAuth.CallbackHandler)
		}
	}

	return r, nil
}
