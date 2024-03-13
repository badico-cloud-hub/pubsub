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
	url           string
	conn          *amqp.Connection
	ch            *amqp.Channel
	chNotify      *amqp.Channel
	chCallback    *amqp.Channel
	queue         amqp.Queue
	queueDlq      amqp.Queue
	queueNotify   amqp.Queue
	queueCallback amqp.Queue
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
	pubsubChannel, err := connection.Channel()
	if err != nil {
		return err
	}
	notifyChannel, err := connection.Channel()
	if err != nil {
		return err
	}
	callbackChannel, err := connection.Channel()
	if err != nil {
		return err
	}

	args := amqp.Table{
		amqp.QueueTypeArg: amqp.QueueTypeQuorum,
	}

	q, err := pubsubChannel.QueueDeclare(
		os.Getenv("QUEUE_PUBSUB"), // name
		true,                      // durable
		false,                     // delete when unused
		false,                     // exclusive
		false,                     // no-wait
		args,                      // arguments
	)

	if err != nil {
		return err
	}

	qd, err := pubsubChannel.QueueDeclare(
		os.Getenv("QUEUE_PUBSUB_DLQ"), // name
		true,                          // durable
		false,                         // delete when unused
		false,                         // exclusive
		false,                         // no-wait
		args,                          // arguments
	)
	if err != nil {
		return err
	}

	qn, err := notifyChannel.QueueDeclare(
		os.Getenv("QUEUE_NOTIFY"), // name
		true,                      // durable
		false,                     // delete when unused
		false,                     // exclusive
		false,                     // no-wait
		args,                      // arguments
	)
	if err != nil {
		return err
	}

	qc, err := notifyChannel.QueueDeclare(
		os.Getenv("QUEUE_CALLBACK"), // name
		true,                        // durable
		false,                       // delete when unused
		false,                       // exclusive
		false,                       // no-wait
		args,                        // arguments
	)

	r.conn = connection
	r.ch = pubsubChannel
	r.chNotify = notifyChannel
	r.chCallback = callbackChannel
	r.queue = q
	r.queueDlq = qd
	r.queueNotify = qn
	r.queueCallback = qc
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
func (r *RabbitMQ) ConnectionIsClosed() bool {
	return r.conn.IsClosed()
}

func (r *RabbitMQ) ChannelNotifyIsClosed() bool {
	return r.chNotify.IsClosed()
}
func (r *RabbitMQ) ChannelPubSubIsClosed() bool {
	return r.ch.IsClosed()
}
func (r *RabbitMQ) ChannelCallbackIsClosed() bool {
	return r.chCallback.IsClosed()
}

//Producer is send message to broker
func (r *RabbitMQ) Producer(queueMessage dto.NotifierDTO) error {
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
	if err := r.chNotify.PublishWithContext(
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

//ProducerCashinCallback is send message to callback queue
func (r *RabbitMQ) ProducerCashinCallback(callbackCashinMessage dto.CallbackCashinMessage) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	notifyBytes, err := json.Marshal(callbackCashinMessage)
	if err != nil {
		return err
	}
	if err := r.chCallback.PublishWithContext(
		ctx,
		"",
		r.queueCallback.Name,
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

//ProducerCashoutCallback is send message to callback queue
func (r *RabbitMQ) ProducerCashoutCallback(callbackCashoutMessage dto.CallbackCashoutMessage) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	notifyBytes, err := json.Marshal(callbackCashoutMessage)
	if err != nil {
		return err
	}
	if err := r.chCallback.PublishWithContext(
		ctx,
		"",
		r.queueCallback.Name,
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
	msgs, err := r.chNotify.Consume(
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
