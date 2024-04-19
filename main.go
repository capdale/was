package main

import (
	"flag"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/capdale/was/config"
	"github.com/capdale/was/logger"
	"github.com/capdale/was/server"
	"github.com/gin-gonic/autotls"
	"go.uber.org/zap"
)

func main() {
	configPath := flag.String("configpath", "config.yaml", "config path")
	config, err := config.ParseConfig(*configPath)
	if err != nil {
		panic(err)
	}

	r, err := server.SetupRouter(config)
	if err != nil {
		panic(err)
	}
	logger.Logger.Info("Server Start", zap.Time("time", time.Now().Local()))

	if strings.HasPrefix(config.Service.Address, "localhost") {
		svr := http.Server{
			Addr:    config.Service.Address,
			Handler: r,
		}
		if err := svr.ListenAndServe(); err != nil {
			log.Fatalf("listen: %s\n", err)
			return
		}
	} else {
		if err := autotls.Run(r, config.Service.Address); err != nil {
			log.Fatalf("tls listen %s\n", err)
			return
		}
	}

	// srv := &http.Server{
	// 	Addr:    config.Service.Address,
	// 	Handler: r,
	// }

	// go func() {
	// 	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
	// 		log.Fatalf("listen: %s\n", err)
	// 	}
	// }()

	// quit := make(chan os.Signal, 2)

	// signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// <-quit
	// log.Println("Shutdown server...")

	// ctx, cancel := context.WithTimeout(context.Background(), 0*time.Second)
	// defer cancel()
	// if err := srv.Shutdown(ctx); err != nil {
	// 	log.Fatal("Server Shutdown:", err)
	// }
	// select {
	// case <-ctx.Done():
	// 	log.Println("timeout of 10 seconds")
	// }
	// log.Println("Server exiting")
}
