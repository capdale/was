package server

import (
	"time"

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

	config, err := parseConfig("config.yaml")
	if err != nil {
		panic("parse config error")
	}

	return
}
