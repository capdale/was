package server

import (
	"time"

	"github.com/capdale/was/config"
	"github.com/gin-contrib/cors"
)

func createCorsConfig(conf *config.Cors) *cors.Config {
	return &cors.Config{
		AllowOrigins:           conf.AllowOrigins,
		AllowMethods:           conf.AllowMethods,
		AllowHeaders:           conf.AllowHeaders,
		AllowCredentials:       conf.AllowCredentials,
		ExposeHeaders:          conf.ExposeHeaders,
		MaxAge:                 time.Duration(conf.MaxAge) * time.Second,
		AllowWildcard:          conf.AllowWildcard,
		AllowBrowserExtensions: conf.AllowBrowserExtensions,
		AllowWebSockets:        conf.AllowWebSockets,
		AllowFiles:             conf.AllowFiles,
	}
}
