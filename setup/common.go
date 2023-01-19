package setup

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/badico-cloud-hub/pubsub/consumer"
)

func SetupNotifyEventConsumer(sqs *sqs.SQS) {
	queueUrl := os.Getenv("NOTIFY_SUBSCRIBERS_QUEUE_URL")
	interval := 100 * time.Millisecond

	consumer, err := consumer.NewSQSConsumer(queueUrl, sqs, consumer.NewNotifyEventHandler(), 10, interval)

	if err != nil {
		fmt.Println(err)
	}

	consumer.Init()

	fmt.Println("Running NotifyEventConsumer")
}
