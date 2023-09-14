package server

import (
	"net/http"
	"time"

	"github.com/capdale/was/api/collect"
	"github.com/capdale/was/config"
	"github.com/capdale/was/database"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() (r *gin.Engine) {
	r = gin.Default()

	// config := cors.DefaultConfig()

	r.Use(cors.New(
		cors.Config{
			AllowOrigins:     []string{"http://localhost:5500"}, // test front addr
			AllowMethods:     []string{"POST", "GET", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "content-type"},
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

	r.GET("/", func(ctx *gin.Context) {
		ctx.AbortWithStatusJSON(http.StatusOK, gin.H{
			"ok": "ok",
		})
	})

	collectAPI := collect.New(d)

	collectRouter := r.Group("/collect")
	{
		collectRouter.POST("/", collectAPI.PostCollectHandler)
	}

	return
}
