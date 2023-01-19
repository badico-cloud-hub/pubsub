package infra

import (
	"encoding/json"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/badico-cloud-hub/pubsub/dto"
)

//SqsClient is struct of client sqs
type SqsClient struct {
	Client   *sqs.SQS
	queueUrl string
}

//NewSqsClient return new client the sqs
func NewSqsClient() *SqsClient {
	return &SqsClient{
		queueUrl: os.Getenv("NOTIFY_SUBSCRIBERS_QUEUE_URL"),
	}
}

//Setup execute configuration for session the sqs
func (s *SqsClient) Setup() error {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String(os.Getenv("AWS_DEFAULT_REGION")),
		},
	})
	if err != nil {
		return err
	}
	cvc := sqs.New(sess)
	s.Client = cvc
	return nil
}

//Send is function for execute send message to queue
func (s *SqsClient) Send(subs dto.SubscriptionDTO, notification dto.NotifierDTO) (interface{}, error) {
	var body = notification.Data
	body["topic"] = subs.SubscriptionEvent
	body["created_at"] = notification.CreatedAt

	sqsMessage := dto.SqsMessage{
		ClientId:      subs.ClientId,
		AssociationId: subs.AssociationId,
		Url:           subs.SubscriptionUrl,
		AuthProvider:  subs.AuthProvider,
		Body:          body,
	}
	notifyBytes, err := json.Marshal(sqsMessage)
	if err != nil {
		return nil, err
	}

	message := string(notifyBytes)
	input := sqs.SendMessageInput{
		QueueUrl:    &s.queueUrl,
		MessageBody: &message,
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"ClientId": {
				DataType:    aws.String("String"),
				StringValue: &subs.ClientId,
			},
			"Event": {
				DataType:    aws.String("String"),
				StringValue: &subs.SubscriptionEvent,
			},
		},
	}

	output, err := s.Client.SendMessage(&input)
	if err != nil {
		return nil, err
	}
	return output, nil
}
