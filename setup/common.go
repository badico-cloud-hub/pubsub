package setup

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/badico-cloud-hub/pubsub/consumer"
)

func SetupNotifyEventConsumer(sqs *sqs.SQS, wg *sync.WaitGroup) {
	queueUrl := os.Getenv("NOTIFY_SUBSCRIBERS_QUEUE_URL")
	dlq := os.Getenv("NOTIFY_SUBSCRIBERS_QUEUE_URL_DLQ")
	interval := 100 * time.Millisecond

	// logManager := infra.NewLogManager()
	// logManager.StartProducer()
	// defer func() {
	// logManager.StopProducer()
	// }()

	// setupLog := logManager.NewLogger("logger setup notify event consumer - ", os.Getenv("MACHINE_IP"))

	consumer, err := consumer.NewSQSConsumer(queueUrl, dlq, sqs, consumer.NewNotifyEventHandler(), 10, interval)

	if err != nil {
		fmt.Println(err)
	}

	consumer.Init(wg)

	// setupLog.Infoln("Running NotifyEventConsumer")
}
