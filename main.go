package main

import (
	"github.com/capdale/was/config"
	"github.com/capdale/was/server"
)

func main() {
	config, err := config.ParseConfig("config.yaml")
	if err != nil {
		panic(err)
	}

	r, c, err := server.SetupRouter(config)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	r.Run(config.Service.Address)
}
