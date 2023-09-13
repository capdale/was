package main

import "github.com/capdale/was/server"

func main() {
	r := server.SetupRouter()

	r.Run(":8080")
}
