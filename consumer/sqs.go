package consumer

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sync"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/badico-cloud-hub/log-driver/producer"
	"github.com/badico-cloud-hub/pubsub/dto"
	"github.com/badico-cloud-hub/pubsub/entity"
	"github.com/badico-cloud-hub/pubsub/infra"
	"github.com/badico-cloud-hub/pubsub/utils"
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
	rabbitMqClient      *infra.RabbitMQ
	logManager          *producer.LoggerManager
	dynamo              *infra.DynamodbClient
}

func NewSQSConsumer(queueUrl, dlq string, sqsClient *sqs.SQS, consumer ConsumerHandler, maxNumberOfMessages int64, logManager *producer.LoggerManager, dynamoClient *infra.DynamodbClient) (*SQSConsumer, error) {

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
		rabbitMqClient,
		logManager,
		dynamoClient,
	}

	return newConsumer, nil
}

func (c *SQSConsumer) Init(wg *sync.WaitGroup) {
	wg.Add(3)
	go c.initChannels(wg)
	go c.notify(wg)
	go c.consume()
}

func (c *SQSConsumer) initChannels(wg *sync.WaitGroup) {
	for {
		wg.Add(1)
		fmt.Printf("LENGTH GO ROUTINES: %+v\n", runtime.NumGoroutine())
		select {
		case queueMessage := <-c.handle:
			go c.handleMessage(queueMessage, wg)

		case err := <-c.err:
			go c.handleError(err, wg)

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
			consumerLog.Errorf("failed to fetch queue message %v in a queue %s", err, c.queueUrl)
		}
		c.handle <- &queueMessage

	}

}

func (c *SQSConsumer) notify(wg *sync.WaitGroup) {
	notifyLog := c.logManager.NewLogger("logger consumer queue - ", os.Getenv("MACHINE_IP"))
	notifyLog.Infoln("Start consume notify queue...")

	msgs, err := c.rabbitMqClient.ConsumerNotifyQueue()
	if err != nil {
		notifyLog.Errorf("failed to fetch queue message %v in a queue %s", err, c.queueUrl)
	}

	semaphore := make(chan struct{}, 500)

	for msg := range msgs {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(msgQeue amqp091.Delivery, wg *sync.WaitGroup, semaphore *chan struct{}) {
			defer wg.Done()
			notif, err := adaptQueueMessageFromNotifyQueue(msgQeue)
			if err != nil {
				notifyLog.Errorf("failed to fetch queue message %v in a queue %s", err, c.queueUrl)
				return
			}
			notifyLog.AddTraceRef(fmt.Sprintf("AssociationsId: %+v", notif.AssociationsId))
			notifyLog.AddTraceRef(fmt.Sprintf("Event: %+v", notif.Event))

			subscriptions := []entity.Subscription{}
			subscriptionsFiltered := []dto.SubscriptionDTO{}

			for _, associationId := range notif.AssociationsId {
				newSubsciptions, err := c.dynamo.GetSubscriptionByAssociationIdAndEvent(associationId, notif.Event)
				if err != nil && err == infra.ErrorSubscriptinEventNotFound {
					notifyLog.Errorf("Subscription with AssociationId %s and Event %s not found", associationId, notif.Event)
					return
				}

				if err != nil {
					notifyLog.Errorln(err.Error())
					return
				}

				subscriptions = append(subscriptions, newSubsciptions...)
			}

			for _, subscription := range subscriptions {
				if duplicated := utils.VerifyIfUrlIsDuplicated(subscriptionsFiltered, subscription.SubscriptionUrl, subscription.SubscriptionEvent); !duplicated {
					subscriptionsFiltered = append(subscriptionsFiltered, dto.SubscriptionDTO{
						AssociationId:     subscription.AssociationId,
						ClientId:          subscription.ClientId,
						AuthProvider:      "",
						SubscriptionUrl:   subscription.SubscriptionUrl,
						SubscriptionEvent: subscription.SubscriptionEvent,
						CreatedAt:         notif.CreatedAt,
					})
				}
			}

			for _, subscription := range subscriptionsFiltered {
				notif.Data["topic"] = subscription.SubscriptionEvent
				notif.Data["created_at"] = subscription.CreatedAt
				queueMessage := dto.QueueMessage{
					ClientId:      subscription.ClientId,
					AssociationId: subscription.AssociationId,
					Url:           subscription.SubscriptionUrl,
					AuthProvider:  subscription.AuthProvider,
					Body:          notif.Data,
				}
				notifyLog.Infof("Sending to qeueue: %+v", queueMessage)
				err = c.rabbitMqClient.Producer(queueMessage)
				if err != nil {
					notifyLog.Errorf("failed to fetch queue message %v in a queue %s", err, c.queueUrl)
					continue
				}

			}
			<-*semaphore
		}(msg, wg, &semaphore)

	}
}

func (c *SQSConsumer) handleMessage(queueMessage *dto.QueueMessage, wg *sync.WaitGroup) {
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
}

func (c *SQSConsumer) handleError(errorMessage *ErrorMessage, wg *sync.WaitGroup) {
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
}

func adaptQueueMessage(message amqp091.Delivery) (dto.QueueMessage, error) {
	queueMessage := dto.QueueMessage{}
	err := json.Unmarshal(message.Body, &queueMessage)

	if err != nil {
		return dto.QueueMessage{}, err
	}

	return queueMessage, nil
}

func adaptQueueMessageFromNotifyQueue(message amqp091.Delivery) (dto.NotifierDTO, error) {
	notif := dto.NotifierDTO{}
	err := json.Unmarshal(message.Body, &notif)

	if err != nil {
		return dto.NotifierDTO{}, err
	}

	if len(notif.Data) == 0 {
		notif.Data = make(map[string]interface{})
	}

	return notif, nil
}
