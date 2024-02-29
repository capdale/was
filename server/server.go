package server

import (
	"net/http"
	"time"

	articleAPI "github.com/capdale/was/api/article"
	authapi "github.com/capdale/was/api/auth"
	githubAuth "github.com/capdale/was/api/auth/github"
	originAPI "github.com/capdale/was/api/auth/origin"
	collect "github.com/capdale/was/api/collection"
	reportAPI "github.com/capdale/was/api/report"
	"github.com/capdale/was/auth"
	"github.com/capdale/was/config"
	"github.com/capdale/was/database"
	"github.com/capdale/was/logger"
	"github.com/capdale/was/s3"
	"github.com/capdale/was/store"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"gopkg.in/natefinch/lumberjack.v2"
)

func SetupRouter(config *config.Config) (r *gin.Engine, err error) {
	r = gin.New()

	isProduction := gin.Mode() == gin.ReleaseMode

	// if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
	// 	v.RegisterCustomTypeFunc(binaryuuid.ValidateUUID, binaryuuid.UUID{})
	// }

	routerLogger := logger.New(&lumberjack.Logger{
		Filename:   config.Service.Log.Path,
		MaxSize:    config.Service.Log.MaxSize,
		MaxBackups: config.Service.Log.MaxBackups,
		MaxAge:     config.Service.Log.MaxAge,
	}, isProduction, config.Service.Log.Console)
	logger.Init(routerLogger)

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

	s3storage, err := s3.New(&config.S3)
	if err != nil {
		return
	}

	auth := auth.New(d, store)

	r.GET("/", func(ctx *gin.Context) {
		logger.Logger.InfoWithCTX(ctx, "log check", zap.String("asdf", "A"))
		ctx.JSON(http.StatusOK, gin.H{
			"ok": "ok",
		})
	})

	collectAPI := collect.New(d, s3storage)

	collectRouter := r.Group("/collection").Use(auth.AuthorizeRequiredMiddleware())
	{
		collectRouter.GET("/", auth.AuthorizeRequiredMiddleware(), collectAPI.GetCollectection)
		collectRouter.POST("/", auth.AuthorizeRequiredMiddleware(), collectAPI.CreateCollectionHandler)
		collectRouter.GET("/:uuid", collectAPI.GetCollectionByUUID)
	}

	authAPI := authapi.New(d, auth)
	originAPI := originAPI.New(d, auth)
	authRouter := r.Group("/auth")
	{
		authRouter.GET("/whoami", auth.AuthorizeRequiredMiddleware(), authAPI.WhoAmIHandler)
		authRouter.POST("/blacklist", auth.AuthorizeRequiredMiddleware(), authAPI.SetBlacklistHandler)
		authRouter.POST("/refresh", authAPI.RefreshTokenHandler)
		githubAuth := githubAuth.New(d, auth, &oauth2.Config{
			ClientID:     config.Oauth.Github.Id,
			ClientSecret: config.Oauth.Github.Secret,
			RedirectURL:  config.Oauth.Github.Redirect,
			Scopes:       []string{"user:email"},
			Endpoint:     github.Endpoint,
		})
		authRouter.POST("/regist-email", originAPI.CreateEmailTicketHandler)
		authRouter.POST("/regist", originAPI.RegisterTicketHandler)
		authRouter.POST("/login", originAPI.LoginHandler)
		githubAuthRouter := authRouter.Group("/github").Use(sessions.Sessions("state", st))
		{
			githubAuthRouter.GET("/login", githubAuth.LoginHandler)
			githubAuthRouter.GET("/callback", githubAuth.CallbackHandler)
		}
	}

	reportAPI := reportAPI.New(d)
	reportRouter := r.Group("/report", auth.AuthorizeOptionalMiddleware()) // anonymous can report too
	{
		reportRouter.POST("/user", reportAPI.PostUserReportHandler)
		reportRouter.POST("/article", reportAPI.PostReportArticleHandler)
		reportRouter.POST("/bug", reportAPI.PostReportBugHandler)
		reportRouter.POST("/help", reportAPI.PostReportHelpHandler)
		reportRouter.POST("/etc", reportAPI.PostReportEtcHandler)
	}

	articleAPI := articleAPI.New(d, s3storage)
	articleRouter := r.Group("/article")
	{
		articleRouter.POST("/", auth.AuthorizeRequiredMiddleware(), articleAPI.CreateArticleHandler)
		articleRouter.GET("/get-links/:useruuid", articleAPI.GetUserArticleLinksHandler)
		articleRouter.GET("/:link", articleAPI.GetArticleHandler)
	}

	return r, nil
}
