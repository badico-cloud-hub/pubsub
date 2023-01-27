package main

import (
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/badico-cloud-hub/pubsub/api"
	"github.com/badico-cloud-hub/pubsub/infra"
	"github.com/badico-cloud-hub/pubsub/setup"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
	isServerEnv := os.Getenv("IS_SERVER")
	isServer, err := strconv.ParseBool(isServerEnv)
	if err != nil {
		log.Fatal(err)
	}

	if isServer {
		port := os.Getenv("PORT")
		api := api.NewServer(port)
		api.AllRouters()
		if err := api.Run(); err != nil {
			log.Fatal(err)
		}
	} else {
		wg := &sync.WaitGroup{}
		sqs := infra.NewSqsClient()
		if err := sqs.Setup(); err != nil {
			log.Fatal(err)
		}
		setup.SetupNotifyEventConsumer(sqs.Client, wg)
		wg.Wait()
	}

}
