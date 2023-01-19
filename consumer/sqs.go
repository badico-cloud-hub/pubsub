package consumer

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type ConsumerMessage struct {
	Body          map[string]interface{}
	ReceiptHandle string
	handleChannel chan *sqs.Message
	removeChannel chan *sqs.Message
}

type ErrorMessage struct {
	Source        string                 `json:"source,omitempty"`
	Input         map[string]interface{} `json:"input,omitempty"`
	Reason        string                 `json:"reason"`
	Output        map[string]interface{} `json:"output,omitempty"`
	SourceMessage *sqs.Message           `json:"-"`
}

type ConsumerHandler interface {
	Handle(ConsumerMessage) (map[string]interface{}, error)
}

type SQSConsumer struct {
	consumer            ConsumerHandler
	sqsClient           *sqs.SQS
	queueUrl            string
	err                 chan *ErrorMessage
	handle              chan *sqs.Message
	remove              chan *sqs.Message
	maxNumberOfMessages int64
	pollInterval        time.Duration
}

func NewSQSConsumer(queueUrl string, sqsClient *sqs.SQS, consumer ConsumerHandler, maxNumberOfMessages int64, pollInterval time.Duration) (*SQSConsumer, error) {

	err := make(chan *ErrorMessage)
	handle := make(chan *sqs.Message)
	remove := make(chan *sqs.Message)

	newConsumer := &SQSConsumer{
		consumer,
		sqsClient,
		queueUrl,
		err,
		handle,
		remove,
		maxNumberOfMessages,
		pollInterval,
	}

	return newConsumer, nil
}

func (c *SQSConsumer) Init() {
	go c.initChannels()
	go c.consume()
}

func (c *SQSConsumer) initChannels() {
	for {
		fmt.Printf("LENGTH GO ROUTINES: %+v\n", runtime.NumGoroutine())
		select {
		case sqsMessage := <-c.remove:
			go c.removeMessage(sqsMessage)

		case sqsMessage := <-c.handle:
			go c.handleMessage(sqsMessage)

		case err := <-c.err:
			go c.handleError(err)

		}
	}

}

func (c *SQSConsumer) consume() {
	for {
		output, err := c.sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(c.queueUrl),
			MaxNumberOfMessages: aws.Int64(c.maxNumberOfMessages),
		})

		if err != nil {
			log.Printf("failed to fetch sqs message %v", err)
		}

		for _, SQSMessage := range output.Messages {
			c.handle <- SQSMessage
		}

		time.Sleep(c.pollInterval)
	}
}

func (c *SQSConsumer) removeMessage(sqsMessage *sqs.Message) {
	_, err := c.sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.queueUrl),
		ReceiptHandle: sqsMessage.ReceiptHandle,
	})

	if err != nil {
		log.Print("[Consumer][Error] DeleteMessage from Queue: ", c.queueUrl)
		log.Print("[Consumer][Error] MessageId: ", sqsMessage.MessageId)
		log.Print("[Consumer][Error] Reason: ", err.Error())
		c.err <- &ErrorMessage{
			Reason:        err.Error(),
			SourceMessage: sqsMessage,
		}
	}
}

func (c *SQSConsumer) handleMessage(sqsMessage *sqs.Message) {
	message, err := adaptSQSMessage(sqsMessage)
	message.handleChannel = c.handle
	message.removeChannel = c.remove
	message.ReceiptHandle = *sqsMessage.ReceiptHandle

	if err != nil {
		c.err <- &ErrorMessage{
			Reason:        err.Error(),
			SourceMessage: sqsMessage,
			Input:         message.Body,
		}
	}

	output, err := c.consumer.Handle(message)

	if err != nil {
		c.err <- &ErrorMessage{
			Reason:        err.Error(),
			SourceMessage: sqsMessage,
			Input:         message.Body,
			Output:        output,
		}

		return
	}

	c.remove <- sqsMessage
}

func (c *SQSConsumer) handleError(errorMessage *ErrorMessage) {

	fmt.Printf("Error to process message: %+v\n", errorMessage)
	c.remove <- errorMessage.SourceMessage
}

func adaptSQSMessage(message *sqs.Message) (ConsumerMessage, error) {
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(*message.Body), &jsonMap)

	if err != nil {
		return ConsumerMessage{}, err
	}

	return ConsumerMessage{
		Body: jsonMap,
	}, nil
}
