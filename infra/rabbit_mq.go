package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/badico-cloud-hub/pubsub/dto"
	amqp "github.com/rabbitmq/amqp091-go"
)

//RabbitMQ is struct for broker in amazon mq
type RabbitMQ struct {
	url         string
	conn        *amqp.Connection
	ch          *amqp.Channel
	queue       amqp.Queue
	queueDlq    amqp.Queue
	queueNotify amqp.Queue
}

//NewRabbitMQ return new instance of RabbitMQ
func NewRabbitMQ() *RabbitMQ {

	amzMqUrl := os.Getenv("AMAZON_MQ_URL")
	return &RabbitMQ{
		url: amzMqUrl,
	}
}

//Setup is configure broker
func (r *RabbitMQ) Setup() error {
	connection, err := amqp.Dial(r.url)
	if err != nil {
		return err
	}
	pusubChannel, err := connection.Channel()
	notifyChannel, err := connection.Channel()
	if err != nil {
		return err
	}
	if err := pusubChannel.Qos(500, 0, false); err != nil {
		return err
	}
	if err := notifyChannel.Qos(500, 0, false); err != nil {
		return err
	}
	q, err := pusubChannel.QueueDeclare(
		"pubsub", // name
		true,     // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)

	if err != nil {
		return err
	}

	qd, err := pusubChannel.QueueDeclare(
		"pubsub_dlq", // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)

	qn, err := notifyChannel.QueueDeclare(
		"pubsub_service_notify", // name
		true,                    // durable
		false,                   // delete when unused
		false,                   // exclusive
		false,                   // no-wait
		nil,                     // arguments
	)

	if err != nil {
		return err
	}

	r.conn = connection
	r.ch = pusubChannel
	r.queue = q
	r.queueDlq = qd
	r.queueNotify = qn
	return nil
}

func (r *RabbitMQ) Ack(deliveredTag uint64) error {
	err := r.ch.Ack(deliveredTag, false) // <-- Difference
	if err != nil {
		return err
	}
	return nil
}

func (r *RabbitMQ) NumberOfMessagesQueue() error {
	fmt.Printf("%s: contains -> %v\n", r.queue.Name, r.queue.Messages)
	fmt.Printf("%s: contains -> %v\n", r.queueDlq.Name, r.queueDlq.Messages)
	return nil
}

//Producer is send message to broker
func (r *RabbitMQ) Producer(queueMessage dto.QueueMessage) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	queueMessageBytes, err := json.Marshal(queueMessage)
	if err != nil {
		return err
	}
	if err := r.ch.PublishWithContext(
		ctx,
		"",
		r.queue.Name,
		false,
		false,
		amqp.Publishing{
			Timestamp: time.Now(),
			Type:      "text/plain",
			Body:      queueMessageBytes,
		}); err != nil {
		return err
	}

	return nil
}

//Dlq is send message to dlq broker
func (r *RabbitMQ) Dlq(queueMessage dto.QueueMessage) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	queueMessageBytes, err := json.Marshal(queueMessage)
	if err != nil {
		return err
	}
	if err := r.ch.PublishWithContext(
		ctx,
		"",
		r.queueDlq.Name,
		false,
		false,
		amqp.Publishing{
			Timestamp: time.Now(),
			Type:      "text/plain",
			Body:      queueMessageBytes,
		}); err != nil {
		return err
	}

	return nil
}

//ProducerNotify is send message to notify queue
func (r *RabbitMQ) ProducerNotify(notify dto.NotifierDTO) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	notifyBytes, err := json.Marshal(notify)
	if err != nil {
		return err
	}
	if err := r.ch.PublishWithContext(
		ctx,
		"",
		r.queueNotify.Name,
		false,
		false,
		amqp.Publishing{
			Timestamp: time.Now(),
			Type:      "text/plain",
			Body:      notifyBytes,
		}); err != nil {
		return err
	}

	return nil
}

//Consumer is return channel for consume from broker
func (r *RabbitMQ) Consumer() (<-chan amqp.Delivery, error) {
	msgs, err := r.ch.Consume(
		r.queue.Name, // queue
		"",           // consumer
		true,         // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

func (r *RabbitMQ) ConsumerNotifyQueue() (<-chan amqp.Delivery, error) {
	msgs, err := r.ch.Consume(
		r.queueNotify.Name, // queue
		"",                 // consumer
		true,               // auto-ack
		false,              // exclusive
		false,              // no-local
		false,              // no-wait
		nil,                // args
	)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

//ConsumerDlq is return channel for consume dlq from broker
func (r *RabbitMQ) ConsumerDlq() (<-chan amqp.Delivery, error) {
	msgs, err := r.ch.Consume(
		r.queueDlq.Name, // queue
		"",              // consumer
		true,            // auto-ack
		false,           // exclusive
		false,           // no-local
		false,           // no-wait
		nil,             // args
	)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

//Release is close connection
func (r *RabbitMQ) Release() {
	r.conn.Close()
}
