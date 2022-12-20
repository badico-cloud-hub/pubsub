package infra

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/badico-cloud-hub/pubsub/dto"
	"github.com/badico-cloud-hub/pubsub/entity"
	"github.com/badico-cloud-hub/pubsub/interfaces"
	"github.com/badico-cloud-hub/pubsub/utils"
	"github.com/google/uuid"
)

//DynamodbClient is struct for client dynamodb
type DynamodbClient struct {
	logger interfaces.ServiceLogger
	client *dynamodb.DynamoDB
}

//NewDynamodbClient return new client dynamodb
func NewDynamodbClient() *DynamodbClient {
	return &DynamodbClient{
		logger: utils.NewLogger(os.Stdout),
	}
}

//Setup execute configuration the session of client dynamodb
func (d *DynamodbClient) Setup() error {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String(os.Getenv("AWS_DEFAULT_REGION")),
		},
	})
	if err != nil {
		return err
	}
	cvc := dynamodb.New(sess)
	d.client = cvc
	return nil
}

//CreateSubscription execute creation the subscription in dynamo table
func (d *DynamodbClient) CreateSubscription(subs *dto.SubscriptionDTO) (dto.SubscriptionDTO, error) {
	subscriptions := []entity.Subscription{}
	id := uuid.New()
	subscriptionResult := dto.SubscriptionDTO{SubscriptionId: id.String()}
	if len(subs.Events) > 0 {
		filteredsEvents := utils.FilterEvents(subs.Events)
		for _, event := range filteredsEvents {
			subscription := entity.Subscription{}
			subscription.PK = fmt.Sprintf("SUBSCRIPTION#%s", id.String())
			subscription.SK = fmt.Sprintf("SUBSCRIPTION_EVENT#%s", event)
			subscription.GSIPK = fmt.Sprintf("CLIENT#%s", subs.ClientId)
			subscription.GSISK = fmt.Sprintf("SUBSCRIPTION_EVENT#%s", event)
			subscription.ClientId = subs.ClientId
			subscription.SubscriptionEvent = event
			subscription.SubscriptionId = id.String()
			subscription.SubscriptionUrl = subs.Url
			subscription.CreatedAt = time.Now().UTC().String()
			subscription.UpdatedAt = time.Now().UTC().String()

			subscriptions = append(subscriptions, subscription)
		}

	}
	putItensError := []entity.Subscription{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	for _, item := range subscriptions {
		putItem, err := dynamodbattribute.MarshalMap(item)
		if err != nil {
			d.logger.Error(err.Error())
			putItensError = append(putItensError, item)
			continue
		}
		input := &dynamodb.PutItemInput{
			TableName: aws.String(os.Getenv("BATCH_TABLE")),
			Item:      putItem,
		}
		_, err = d.client.PutItemWithContext(ctx, input)
		if err != nil {
			d.logger.Error(err.Error())
			putItensError = append(putItensError, item)
			continue
		}
	}
	d.logger.Error(fmt.Sprintf("ITENS ERRORS: %+v", putItensError))
	return subscriptionResult, nil
}

//ListSubscriptions return all subscriptions the client
func (d *DynamodbClient) ListSubscriptions(clientId string) ([]dto.SubscriptionDTO, error) {
	filt := expression.And(expression.Name("PK").BeginsWith("SUBSCRIPTION#"), expression.Name("client_id").Equal(expression.Value(clientId)))
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		return []dto.SubscriptionDTO{}, err
	}

	input := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		TableName:                 aws.String(os.Getenv("BATCH_TABLE")),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	output, err := d.client.ScanWithContext(ctx, input)
	if err != nil {
		return []dto.SubscriptionDTO{}, err
	}
	subscriptions := []entity.Subscription{}
	if err := dynamodbattribute.UnmarshalListOfMaps(output.Items, &subscriptions); err != nil {
		return []dto.SubscriptionDTO{}, err
	}
	mapSubscriptions := make(map[string][]entity.Subscription)
	resultSubs := []dto.SubscriptionDTO{}
	if len(subscriptions) > 0 {
		for _, event := range subscriptions {
			mapSubscriptions[event.SubscriptionId] = append(mapSubscriptions[event.SubscriptionId], event)
		}
		for _, sliceSubs := range mapSubscriptions {
			subscription := dto.SubscriptionDTO{}
			subscription.ClientId = sliceSubs[0].ClientId
			subscription.SubscriptionId = sliceSubs[0].SubscriptionId
			subscription.SubscriptionUrl = sliceSubs[0].SubscriptionUrl
			for _, events := range sliceSubs {
				subscription.Events = append(subscription.Events, events.SubscriptionEvent)
			}
			resultSubs = append(resultSubs, subscription)
		}

	}
	return resultSubs, nil
}

//GetSubscription return subscription from event and client
func (d *DynamodbClient) GetSubscription(clientId, event string) (entity.Subscription, error) {
	filt := expression.And(expression.Name("SK").Equal(expression.Value(fmt.Sprintf("SUBSCRIPTION_EVENT#%s", event))), expression.Name("client_id").Equal(expression.Value(clientId)))
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		return entity.Subscription{}, err
	}

	input := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		TableName:                 aws.String(os.Getenv("BATCH_TABLE")),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	output, err := d.client.ScanWithContext(ctx, input)
	if err != nil {
		return entity.Subscription{}, err
	}
	subscriptions := []entity.Subscription{}
	if err := dynamodbattribute.UnmarshalListOfMaps(output.Items, &subscriptions); err != nil {
		return entity.Subscription{}, err
	}
	if len(subscriptions) == 0 {
		return entity.Subscription{}, errors.New("not subscription event in table")
	}
	return subscriptions[0], nil
}

//DeleteSubscription execute remove the event subscription
func (d *DynamodbClient) DeleteSubscription(clientId, idSubscription, event string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	sub := entity.Subscription{
		PK: fmt.Sprintf("SUBSCRIPTION#%s", idSubscription),
		SK: fmt.Sprintf("SUBSCRIPTION_EVENT#%s", event),
	}
	item, err := dynamodbattribute.MarshalMap(sub)
	if err != nil {
		return err
	}
	inputDelete := &dynamodb.DeleteItemInput{
		TableName: aws.String(os.Getenv("BATCH_TABLE")),
		Key:       item,
	}
	_, err = d.client.DeleteItemWithContext(ctx, inputDelete)
	if err != nil {
		return err
	}
	return nil
}

//DescribeTable return informations from table
func (d *DynamodbClient) DescribeTable(tableName string) (interface{}, error) {
	input := &dynamodb.DescribeTableInput{
		TableName: &tableName,
	}
	output, err := d.client.DescribeTable(input)
	if err != nil {
		return nil, err
	}
	return output, nil
}
