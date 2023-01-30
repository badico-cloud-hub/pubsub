package setup

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/badico-cloud-hub/log-driver/producer"
	"github.com/badico-cloud-hub/pubsub/consumer"
)

func SetupNotifyEventConsumer(sqs *sqs.SQS, wg *sync.WaitGroup, logManager *producer.LoggerManager) {
	queueUrl := os.Getenv("NOTIFY_SUBSCRIBERS_QUEUE_URL")
	dlq := os.Getenv("NOTIFY_SUBSCRIBERS_QUEUE_URL_DLQ")
	interval := 100 * time.Millisecond

	setupLog := logManager.NewLogger("logger setup notify event consumer - ", os.Getenv("MACHINE_IP"))

	consumer, err := consumer.NewSQSConsumer(queueUrl, dlq, sqs, consumer.NewNotifyEventHandler(logManager), 10, interval, logManager)

	if err != nil {
		fmt.Println(err)
	}

	consumer.Init(wg)

	setupLog.Infoln("Running NotifyEventConsumer")
}
