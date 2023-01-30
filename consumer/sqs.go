package consumer

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/badico-cloud-hub/log-driver/producer"
	"github.com/badico-cloud-hub/pubsub/dto"
	"github.com/badico-cloud-hub/pubsub/infra"
	"github.com/rabbitmq/amqp091-go"
)

type ConsumerMessage struct {
	QueueMessage  *dto.QueueMessage
	handleChannel chan *dto.QueueMessage
}

type ErrorMessage struct {
	Source        string                 `json:"source,omitempty"`
	Input         map[string]interface{} `json:"input,omitempty"`
	Reason        string                 `json:"reason"`
	Output        map[string]interface{} `json:"output,omitempty"`
	SourceMessage *dto.QueueMessage      `json:"-"`
}

type ConsumerHandler interface {
	Handle(ConsumerMessage) (map[string]interface{}, error)
}

type SQSConsumer struct {
	consumer            ConsumerHandler
	sqsClient           *sqs.SQS
	queueUrl            string
	deadLetterQueueURL  string
	err                 chan *ErrorMessage
	handle              chan *dto.QueueMessage
	maxNumberOfMessages int64
	pollInterval        time.Duration
	rabbitMqClient      *infra.RabbitMQ
	logManager          *producer.LoggerManager
}

func NewSQSConsumer(queueUrl, dlq string, sqsClient *sqs.SQS, consumer ConsumerHandler, maxNumberOfMessages int64, pollInterval time.Duration, logManager *producer.LoggerManager) (*SQSConsumer, error) {

	err := make(chan *ErrorMessage)
	handle := make(chan *dto.QueueMessage)

	rabbitMqClient := infra.NewRabbitMQ()
	if err := rabbitMqClient.Setup(); err != nil {
		fmt.Printf("Error Setup RabbitMQ: %s", err.Error())
	}

	newConsumer := &SQSConsumer{
		consumer,
		sqsClient,
		queueUrl,
		dlq,
		err,
		handle,
		maxNumberOfMessages,
		pollInterval,
		rabbitMqClient,
		logManager,
	}

	return newConsumer, nil
}

func (c *SQSConsumer) Init(wg *sync.WaitGroup) {
	wg.Add(2)
	go c.initChannels(wg)
	go c.consume()
}

func (c *SQSConsumer) initChannels(wg *sync.WaitGroup) {
	semaphore := make(chan struct{}, 10)
	for {
		wg.Add(1)
		fmt.Printf("LENGTH GO ROUTINES: %+v\n", runtime.NumGoroutine())
		select {
		case queueMessage := <-c.handle:
			semaphore <- struct{}{}
			go c.handleMessage(queueMessage, wg, &semaphore)

		case err := <-c.err:
			semaphore <- struct{}{}
			go c.handleError(err, wg, &semaphore)

		}
	}

}

func (c *SQSConsumer) consume() {
	consumerLog := c.logManager.NewLogger("logger consumer queue - ", os.Getenv("MACHINE_IP"))
	consumerLog.Infoln("Start consume queue messages...")

	msgs, err := c.rabbitMqClient.Consumer()
	if err != nil {
		fmt.Printf("Error aqui")
	}
	for msg := range msgs {
		queueMessage, err := adaptQueueMessage(msg)
		if err != nil {
			consumerLog.Errorf("failed to fetch sqs message %v in a queue %s", err, c.queueUrl)
		}
		c.handle <- &queueMessage

		time.Sleep(c.pollInterval)
	}

}

func (c *SQSConsumer) handleMessage(queueMessage *dto.QueueMessage, wg *sync.WaitGroup, semaphore *chan struct{}) {
	handleMessageLog := c.logManager.NewLogger("logger handle message - ", os.Getenv("MACHINE_IP"))
	consumeMessage := ConsumerMessage{
		handleChannel: c.handle,
		QueueMessage:  queueMessage,
	}
	handleMessageLog.AddTraceRef(fmt.Sprintf("ClientId: %s", queueMessage.ClientId))
	handleMessageLog.AddTraceRef(fmt.Sprintf("AssociationId: %s", queueMessage.AssociationId))
	handleMessageLog.AddTraceRef(fmt.Sprintf("QueueUrl: %s", c.queueUrl))

	handleMessageLog.Infoln("Handling message...")
	output, err := c.consumer.Handle(consumeMessage)

	if err != nil {
		c.err <- &ErrorMessage{
			Reason:        err.Error(),
			SourceMessage: consumeMessage.QueueMessage,
			Output:        output,
		}

		return
	}

	wg.Done()
	<-*semaphore
}

func (c *SQSConsumer) handleError(errorMessage *ErrorMessage, wg *sync.WaitGroup, semaphore *chan struct{}) {
	handleErrorLog := c.logManager.NewLogger("logger handle error - ", os.Getenv("MACHINE_IP"))
	handleErrorLog.AddTraceRef(fmt.Sprintf("ClientId: %s", errorMessage.SourceMessage.ClientId))
	handleErrorLog.AddTraceRef(fmt.Sprintf("AssociationId: %s", errorMessage.SourceMessage.AssociationId))
	handleErrorLog.AddTraceRef(fmt.Sprintf("QueueUrl: %s", c.queueUrl))

	err := c.rabbitMqClient.Dlq(*errorMessage.SourceMessage)

	if err != nil {
		handleErrorLog.Errorln(err.Error())
	}

	handleErrorLog.Errorf("Error: %s", errorMessage.Reason)
	handleErrorLog.Infoln("Sending to dlq...")

	fmt.Printf("Error to process message: %+v\n", errorMessage)
	wg.Done()
	<-*semaphore
}

func adaptQueueMessage(message amqp091.Delivery) (dto.QueueMessage, error) {
	queueMessage := dto.QueueMessage{}
	err := json.Unmarshal(message.Body, &queueMessage)

	if err != nil {
		return dto.QueueMessage{}, err
	}

	return queueMessage, nil
}
