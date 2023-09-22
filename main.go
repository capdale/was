package main

import "github.com/capdale/was/server"

func main() {
	r, c := server.SetupRouter()
	defer c.Close()

	r.Run(":8080")
}
