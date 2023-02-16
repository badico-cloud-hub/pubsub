package consumer

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sync"

	"github.com/badico-cloud-hub/log-driver/producer"
	"github.com/badico-cloud-hub/pubsub/dto"
	"github.com/badico-cloud-hub/pubsub/entity"
	"github.com/badico-cloud-hub/pubsub/infra"
	"github.com/badico-cloud-hub/pubsub/utils"
	"github.com/rabbitmq/amqp091-go"
)

type PubsubConsumer struct {
	consumer   ConsumerHandler
	subs       chan dto.NotifierDTO
	err        chan *dto.ErrorMessage
	handle     chan *dto.QueueMessage
	rabbit     *infra.RabbitMQ
	logManager *producer.LoggerManager
	dynamo     *infra.DynamodbClient
	cache      map[string][]entity.Subscription
}

func NewPubsubConsumer(consumer ConsumerHandler, logManager *producer.LoggerManager, dynamoClient *infra.DynamodbClient) (*PubsubConsumer, error) {
	err := make(chan *dto.ErrorMessage)
	handle := make(chan *dto.QueueMessage)
	subs := make(chan dto.NotifierDTO, 5)
	rabbitMqClient := infra.NewRabbitMQ()
	cacheClient := make(map[string][]entity.Subscription)
	if err := rabbitMqClient.Setup(); err != nil {
		return &PubsubConsumer{}, err
	}

	return &PubsubConsumer{
		consumer:   consumer,
		rabbit:     rabbitMqClient,
		logManager: logManager,
		dynamo:     dynamoClient,
		err:        err,
		handle:     handle,
		subs:       subs,
		cache:      cacheClient,
	}, nil
}

func (p *PubsubConsumer) managerChannels(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		wg.Add(1)
		fmt.Printf("LENGTH GO ROUTINES: %+v\n", runtime.NumGoroutine())
		select {
		case notif := <-p.subs:
			go p.getSubscriptions(notif, wg)
		case queueMessage := <-p.handle:
			go p.handleMessage(queueMessage, wg)

		case err := <-p.err:
			go p.handleError(err, wg)

		}
	}
}

func (p *PubsubConsumer) handleMessage(queueMessage *dto.QueueMessage, wg *sync.WaitGroup) {
	handleMessageLog := p.logManager.NewLogger("logger handle message - ", os.Getenv("MACHINE_IP"))
	consumeMessage := ConsumerMessage{
		handleChannel: p.handle,
		QueueMessage:  queueMessage,
	}
	handleMessageLog.AddTraceRef(fmt.Sprintf("ClientId: %s", queueMessage.ClientId))
	handleMessageLog.AddTraceRef(fmt.Sprintf("AssociationId: %s", queueMessage.AssociationId))
	// handleMessageLog.AddTraceRef(fmt.Sprintf("QueueUrl: %s", c.queueUrl))

	handleMessageLog.Infoln("Handling message...")
	output, err := p.consumer.Handle(consumeMessage)

	if err != nil {
		p.err <- &dto.ErrorMessage{
			Reason:        err.Error(),
			SourceMessage: consumeMessage.QueueMessage,
			Output:        output,
		}

		return
	}

	wg.Done()
}

func (p *PubsubConsumer) handleError(errorMessage *dto.ErrorMessage, wg *sync.WaitGroup) {
	handleErrorLog := p.logManager.NewLogger("logger handle error - ", os.Getenv("MACHINE_IP"))
	handleErrorLog.AddTraceRef(fmt.Sprintf("ClientId: %s", errorMessage.SourceMessage.ClientId))
	handleErrorLog.AddTraceRef(fmt.Sprintf("AssociationId: %s", errorMessage.SourceMessage.AssociationId))
	// handleErrorLog.AddTraceRef(fmt.Sprintf("QueueUrl: %s", p.queueUrl))

	err := p.rabbit.Dlq(*errorMessage.SourceMessage)

	if err != nil {
		handleErrorLog.Errorln(err.Error())
	}

	handleErrorLog.Errorf("Error: %s", errorMessage.Reason)
	handleErrorLog.Infoln("Sending to dlq...")

	fmt.Printf("Error to process message: %+v\n", errorMessage)
	wg.Done()
}

func (p *PubsubConsumer) adaptQueueMessageFromNotifyQueue(message amqp091.Delivery) (dto.NotifierDTO, error) {
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

func (p *PubsubConsumer) getSubscriptions(notif dto.NotifierDTO, wg *sync.WaitGroup) {
	defer wg.Done()
	subscriptions := []entity.Subscription{}
	subscriptionsFiltered := []dto.SubscriptionDTO{}
	for _, associationId := range notif.AssociationsId {
		fmt.Printf("CACHE INITIAL: %+v\n", p.cache)
		newSubsciptions := p.cache[associationId]
		if newSubsciptions != nil {
			fmt.Printf("COM CACHE\n")
			subscriptions = append(subscriptions, newSubsciptions...)
		} else {
			fmt.Printf("SEM CACHE\n")
			newSubsciptions, err := p.dynamo.GetSubscriptionByAssociationIdAndEvent(associationId, notif.Event)
			p.cache[associationId] = newSubsciptions
			subscriptions = append(subscriptions, newSubsciptions...)
			if err != nil && err == infra.ErrorSubscriptinEventNotFound {
				fmt.Printf("Subscription with AssociationId %s and Event %s not found\n", associationId, notif.Event)
				break
			}
		}
		fmt.Printf("CACHE FINAL: %+v\n", p.cache)

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
		fmt.Printf("Sending to qeueue: %+v\n", queueMessage)

		p.handle <- &queueMessage

	}

}
func (p *PubsubConsumer) consumeServiceNotifyQueue(wg *sync.WaitGroup) {
	notifyLog := p.logManager.NewLogger("logger consumer queue - ", os.Getenv("MACHINE_IP"))
	notifyLog.Infoln("Start consume notify queue...")
	msgs, err := p.rabbit.ConsumerNotifyQueue()
	if err != nil {
		notifyLog.Errorf("failed to fetch queue message %v in a queue pubsub service notify", err)
	}
	defer wg.Done()
	var forever chan struct{}
	wg.Add(1)

	go func() {
		for msg := range msgs {
			notif, err := p.adaptQueueMessageFromNotifyQueue(msg)
			if err != nil {
				fmt.Printf("failed to fetch queue message %v in a queue pubsub service notify\n", err)
				continue
			}
			p.subs <- notif
		}
	}()

	<-forever
}

func (p *PubsubConsumer) Consume(wg *sync.WaitGroup) {
	wg.Add(3)
	go p.managerChannels(wg)
	go p.consumeServiceNotifyQueue(wg)
}
