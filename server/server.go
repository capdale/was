package server

import (
	"fmt"
	"net/http"
	"time"

	articleAPI "github.com/capdale/was/api/article"
	authapi "github.com/capdale/was/api/auth"
	githubAuth "github.com/capdale/was/api/auth/github"
	originAPI "github.com/capdale/was/api/auth/origin"
	collect "github.com/capdale/was/api/collection"
	reportAPI "github.com/capdale/was/api/report"
	socialAPI "github.com/capdale/was/api/social"
	userAPI "github.com/capdale/was/api/user"
	"github.com/capdale/was/auth"
	"github.com/capdale/was/config"
	"github.com/capdale/was/database"
	"github.com/capdale/was/email"
	"github.com/capdale/was/email/ses"
	"github.com/capdale/was/logger"
	"github.com/capdale/was/storage"
	localstorage "github.com/capdale/was/storage/local"
	"github.com/capdale/was/storage/s3"
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
	r.LoadHTMLGlob("templates/**/*")
	r.Static("/static", "./static")

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
	r.Use(cors.New(*createCorsConfig(&config.Service.Cors)))

	d, err := database.New(&config.Database, 0)
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

	var storage storage.Storage
	if config.Storage.Local != nil {
		storage, err = localstorage.New(config.Storage.Local.BaseDir)
	} else {
		storage, err = s3.New(config.Storage.S3)
	}
	if err != nil {
		return
	}

	auth := auth.New(d, store)

	var emailService email.EmailService
	if config.Email.Mock != nil {
		emailService = email.NewEmailMock()
	} else if config.Email.Ses != nil {
		emailService, err = ses.New(config.Email.Ses)
	}

	if err != nil {
		return
	}

	r.GET("/", func(ctx *gin.Context) {
		logger.Logger.InfoWithCTX(ctx, "log check")
		ctx.JSON(http.StatusOK, gin.H{
			"ok": "ok",
		})
	})

	collectAPI := collect.New(d, storage)

	collectRouter := r.Group("/collection")
	{
		collectRouter.GET("/list/:user", collectAPI.GetUserCollectections)
		collectRouter.POST("/", auth.AuthorizeRequiredMiddleware(), collectAPI.CreateCollectionHandler)
		collectRouter.GET("/:uuid", auth.AuthorizeOptionalMiddleware(), collectAPI.GetCollectionHandler)
		collectRouter.DELETE("/:uuid", auth.AuthorizeRequiredMiddleware(), collectAPI.DeleteCollectionHandler)
		collectRouter.GET("/image/:uuid", auth.AuthorizeOptionalMiddleware(), collectAPI.GetCollectionImageHandler)
	}

	authAPI := authapi.New(d, auth)
	createVerifyLink := func(identifier string) string {
		return fmt.Sprintf("https://%s/auth/register/%s", config.Service.Address, identifier)
	}
	originAPI := originAPI.New(d, auth, emailService, createVerifyLink)
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
		authRouter.POST("/regist-email", originAPI.CreateEmailTicketHandler)
		authRouter.POST("/regist", originAPI.RegisterTicketHandler)
		authRouter.POST("/login", originAPI.LoginHandler)
		githubAuthRouter := authRouter.Group("/github").Use(sessions.Sessions("state", st))
		{
			githubAuthRouter.GET("/login", githubAuth.LoginHandler)
			githubAuthRouter.GET("/callback", githubAuth.CallbackHandler)
		}
		authRouter.DELETE("/", auth.AuthorizeRequiredMiddleware(), authAPI.DeleteUserAccountHandler)

		authRouter.GET("/register/:ticket", originAPI.RegisterTicketView)
	}

	userAPI := userAPI.New(d)
	userRouter := r.Group("/user")
	{
		userRouter.POST("/visibility/:type", auth.AuthorizeRequiredMiddleware(), userAPI.ChangeVisibilityHandler)
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

	articleAPI := articleAPI.New(d, storage)
	articleRouter := r.Group("/article")
	{
		articleRouter.POST("/", auth.AuthorizeRequiredMiddleware(), articleAPI.CreateArticleHandler)
		articleRouter.GET("/get-links/:uuid", articleAPI.GetUserArticleLinksHandler)
		articleRouter.GET("/:link", auth.AuthorizeOptionalMiddleware(), articleAPI.GetArticleHandler)
		articleRouter.DELETE("/:link", auth.AuthorizeRequiredMiddleware(), articleAPI.DeleteArticleHandler)
		articleRouter.GET("/image/:uuid", auth.AuthorizeOptionalMiddleware(), articleAPI.GetArticleImageHandler)
	}

	socialAPI := socialAPI.New(d)
	socialRouter := r.Group("/social")
	{
		// TODO: auth for secret account
		socialRouter.GET("/follower/:target", auth.AuthorizeOptionalMiddleware(), socialAPI.GetFollowersHandler)
		socialRouter.GET("/following/:target", auth.AuthorizeOptionalMiddleware(), socialAPI.GetFollowingsHandler)

		socialRouter.GET("/is-following/:target", auth.AuthorizeRequiredMiddleware(), socialAPI.GetFollowingRelationHandler)
		socialRouter.GET("/is-follower/:target", auth.AuthorizeRequiredMiddleware(), socialAPI.GetFollowerRelationHandler)
		socialRouter.GET("/relation/:target", auth.AuthorizeRequiredMiddleware(), socialAPI.GetRelationHandler)
		socialRouter.DELETE("/follower/:target", auth.AuthorizeRequiredMiddleware(), socialAPI.DeleteFollowerHandler)
		socialRouter.DELETE("/following/:target", auth.AuthorizeRequiredMiddleware(), socialAPI.DeleteFollowingHandler)
		// request follow
		socialRouter.POST("/follow/:target", auth.AuthorizeRequiredMiddleware(), socialAPI.RequestFollowHandler)
		socialRouter.GET("/requests", auth.AuthorizeRequiredMiddleware(), socialAPI.GetFollowRequestsHandler)
		socialRouter.POST("/follow/accept/:request_user", auth.AuthorizeRequiredMiddleware(), socialAPI.AcceptRequestFollowHandler)
		socialRouter.POST("/follow/reject/:request_user", auth.AuthorizeRequiredMiddleware(), socialAPI.RejectRequestFollowHandler)
	}

	return r, nil
}
