package server

import (
	"context"
	"net/http"
	"time"

	authapi "github.com/capdale/was/api/auth"
	githubAuth "github.com/capdale/was/api/auth/github"
	"github.com/capdale/was/api/collect"
	"github.com/capdale/was/auth"
	"github.com/capdale/was/config"
	"github.com/capdale/was/database"
	imagequeue "github.com/capdale/was/queue/image_queue"
	rpcclient "github.com/capdale/was/rpc"
	"github.com/capdale/was/store"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type RouterOpened struct {
	Rpc *rpcclient.RpcService
}

func (r *RouterOpened) Close() {
	for _, service := range *r.Rpc.ImageClassifies {
		service.Conn.Close()
	}
}

func SetupRouter(config *config.Config) (*gin.Engine, *RouterOpened, error) {
	r := gin.Default()

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

	rpcClient, err := rpcclient.New(&config.Rpc)
	if err != nil {
		return nil, nil, err
	}

	d, err := database.New(&config.Mysql)
	if err != nil {
		return nil, nil, err
	}

	store, err := store.New(&config.Redis)
	if err != nil {
		return nil, nil, err
	}

	st, err := redis.NewStore(10, "tcp", config.Redis.Address, config.Redis.Password, []byte(config.Key.SessionStateKey))
	if err != nil {
		return nil, nil, err
	}

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
		githubAuthRouter := authRouter.Group("/github").Use(sessions.Sessions("state", st))
		{
			githubAuthRouter.GET("/login", githubAuth.LoginHandler)
			githubAuthRouter.GET("/callback", githubAuth.CallbackHandler)
		}
	}

	imageQueueCTX := context.Background()
	imageQueue := imagequeue.New(d, time.Duration(time.Second*10), rpcClient.ImageClassifies)
	imageQueue.Run(&imageQueueCTX)

	return r, &RouterOpened{
		Rpc: rpcClient,
	}, nil
}
