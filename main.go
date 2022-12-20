package main

import (
	"log"
	"os"

	"github.com/badico-cloud-hub/pubsub/api"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
	port := os.Getenv("PORT")
	api := api.NewServer(port)
	api.AllRouters()
	if err := api.Run(); err != nil {
		log.Fatal(err)
	}

}
