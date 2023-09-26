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
	rpcclient "github.com/capdale/was/rpc"
	"github.com/capdale/was/store"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"google.golang.org/grpc"
)

type RouterOpened struct {
	Rpc *grpc.ClientConn
}

func (r *RouterOpened) Close() {
	r.Rpc.Close()
}

func SetupRouter(config *config.Config) (*gin.Engine, *RouterOpened, error) {
	r := gin.Default()

	r.Use(cors.New(
		cors.Config{
			AllowOrigins:           config.Service.Cors.AllowOrigins, // test front addr
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

	conn, rpcClient, err := rpcclient.New(&config.Rpc)
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

	r.Use(sessions.Sessions("authstate", st))

	auth := auth.New(d, store)

	r.GET("/", func(ctx *gin.Context) {
		ctx.AbortWithStatusJSON(http.StatusOK, gin.H{
			"ok": "ok",
		})
	})

	collectAPI := collect.New(d, rpcClient.ImageClassifyClient)

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

	return r, &RouterOpened{
		Rpc: conn,
	}, nil
}
