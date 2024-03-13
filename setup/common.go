package setup

import (
	"log"
	"sync"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/badico-cloud-hub/pubsub/consumer"
	"github.com/badico-cloud-hub/pubsub/infra"
)

func SetupNotifyEventConsumer(sqs *sqs.SQS, wg *sync.WaitGroup) {
	// setupLog := logManager.NewLogger("logger setup notify event consumer - ", os.Getenv("MACHINE_IP"))
	dynamoClient := infra.NewDynamodbClient()
	if err := dynamoClient.Setup(); err != nil {
		log.Printf("logger setup notify event consumer - %+v\n", err.Error())
		// setupLog.Errorln(err.Error())
	}
	rabbitMqClient := infra.NewRabbitMQ()
	battery := infra.NewBattery()
	consumer, err := consumer.NewPubsubConsumer(consumer.NewNotifyEventHandler(rabbitMqClient), dynamoClient, rabbitMqClient, battery)
	if err != nil {
		log.Fatal(err)
	}
	consumer.Consume(wg)

	// setupLog.Infoln("Running NotifyEventConsumer")
	log.Printf("logger setup notify event consumer Running NotifyEventConsumer\n")

}
