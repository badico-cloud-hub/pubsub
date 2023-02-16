package setup

import (
	"log"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/badico-cloud-hub/log-driver/producer"
	"github.com/badico-cloud-hub/pubsub/consumer"
	"github.com/badico-cloud-hub/pubsub/infra"
)

func SetupNotifyEventConsumer(sqs *sqs.SQS, wg *sync.WaitGroup, logManager *producer.LoggerManager) {
	setupLog := logManager.NewLogger("logger setup notify event consumer - ", os.Getenv("MACHINE_IP"))
	dynamoClient := infra.NewDynamodbClient()
	if err := dynamoClient.Setup(); err != nil {
		setupLog.Errorln(err.Error())
	}
	consumer, err := consumer.NewPubsubConsumer(consumer.NewNotifyEventHandler(logManager), logManager, dynamoClient)
	if err != nil {
		log.Fatal(err)
	}
	consumer.Consume(wg)

	setupLog.Infoln("Running NotifyEventConsumer")
}
