package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/capdale/was/config"
	"github.com/capdale/was/server"
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

	srv := &http.Server{
		Addr:    config.Service.Address,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 2)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown server...")

	ctx, cancel := context.WithTimeout(context.Background(), 0*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	select {
	case <-ctx.Done():
		log.Println("timeout of 10 seconds")
	}
	log.Println("Server exiting")
}
