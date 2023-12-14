package infra

import (
	"context"
	"fmt"
	"os"
	"strings"
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

// DynamodbClient is struct for client dynamodb
type DynamodbClient struct {
	logger                     interfaces.ServiceLogger
	client                     *dynamodb.DynamoDB
	tableName                  string
	INDEX_AUXILIAR_ASSOCIATION string
}

// NewDynamodbClient return new client dynamodb
func NewDynamodbClient() *DynamodbClient {
	return &DynamodbClient{
		logger:                     utils.NewLogger(os.Stdout),
		tableName:                  os.Getenv("DYNAMO_TABLE_NAME"),
		INDEX_AUXILIAR_ASSOCIATION: "INDEX_AUXILIAR_ASSOCIATION",
	}
}

// Setup execute configuration the session of client dynamodb
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

// CreateSubscription execute creation the subscription in dynamo table
func (d *DynamodbClient) CreateSubscription(subs *dto.SubscriptionDTO) (dto.SubscriptionDTO, error) {
	subscriptions := []entity.Subscription{}
	id := uuid.New()
	subscriptionResult := dto.SubscriptionDTO{SubscriptionId: id.String()}
	if len(subs.Events) > 0 {
		for _, event := range subs.Events {
			subscription := entity.Subscription{}
			subscription.PK = fmt.Sprintf("SUBSCRIPTION#%s", subs.ClientId)
			subscription.SK = fmt.Sprintf("SUBSCRIPTION_EVENT#%s#%s", event, subs.Url)
			subscription.INDEX_AUXILIAR_PK = fmt.Sprintf("ASSOCIATION#%s", subs.AssociationId)
			subscription.INDEX_AUXILIAR_SK = fmt.Sprintf("SUBSCRIPTION_EVENT#%s", event)
			subscription.ClientId = subs.ClientId
			subscription.SubscriptionEvent = event
			subscription.SubscriptionId = id.String()
			subscription.SubscriptionUrl = subs.Url
			subscription.AssociationId = subs.AssociationId
			subscription.Description = subs.Description
			subscription.CreatedAt = time.Now().Format("2006-01-02 15:04:05")
			subscription.UpdatedAt = time.Now().Format("2006-01-02 15:04:05")

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
			TableName: aws.String(d.tableName),
			Item:      putItem,
		}
		_, err = d.client.PutItemWithContext(ctx, input)
		if err != nil {
			d.logger.Error(err.Error())
			putItensError = append(putItensError, item)
			continue
		}
	}
	d.logger.Error(fmt.Sprintf("ITENS SUBSCRIPTIONS ERRORS: %+v", putItensError))
	return subscriptionResult, nil
}

// ListSubscriptions return all subscriptions the client
func (d *DynamodbClient) ListSubscriptions(associationId string) ([]dto.SubscriptionDTO, error) {
	filt := expression.Key("INDEX_AUXILIAR_PK").Equal(expression.Value(fmt.Sprintf("ASSOCIATION#%s", associationId))).And(expression.Key("INDEX_AUXILIAR_SK").BeginsWith("SUBSCRIPTION_EVENT#"))
	expr, err := expression.NewBuilder().WithKeyCondition(filt).Build()
	if err != nil {
		return []dto.SubscriptionDTO{}, err
	}

	input := &dynamodb.QueryInput{
		IndexName:                 &d.INDEX_AUXILIAR_ASSOCIATION,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		TableName:                 aws.String(d.tableName),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	output, err := d.client.QueryWithContext(ctx, input)
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
			subscription.SubscriptionId = sliceSubs[0].SubscriptionId
			subscription.SubscriptionUrl = sliceSubs[0].SubscriptionUrl
			subscription.ClientId = sliceSubs[0].ClientId
			subscription.AssociationId = sliceSubs[0].AssociationId
			subscription.Description = sliceSubs[0].Description
			for _, events := range sliceSubs {
				subscription.Events = append(subscription.Events, events.SubscriptionEvent)
			}
			resultSubs = append(resultSubs, subscription)
		}

	}
	return resultSubs, nil
}

// GetSubscription return subscription from event and associationId
func (d *DynamodbClient) GetSubscription(associationId, event string) (entity.Subscription, error) {
	filt := expression.Key("INDEX_AUXILIAR_PK").Equal(expression.Value(fmt.Sprintf("ASSOCIATION#%s", associationId))).And(expression.Key("INDEX_AUXILIAR_SK").BeginsWith(fmt.Sprintf("SUBSCRIPTION_EVENT#%s", event)))
	expr, err := expression.NewBuilder().WithKeyCondition(filt).Build()
	if err != nil {
		return entity.Subscription{}, err
	}

	input := &dynamodb.QueryInput{
		IndexName:                 &d.INDEX_AUXILIAR_ASSOCIATION,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		TableName:                 aws.String(d.tableName),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	output, err := d.client.QueryWithContext(ctx, input)
	if err != nil {
		return entity.Subscription{}, err
	}
	subscriptions := []entity.Subscription{}
	if err := dynamodbattribute.UnmarshalListOfMaps(output.Items, &subscriptions); err != nil {
		return entity.Subscription{}, err
	}
	if len(subscriptions) == 0 {
		return entity.Subscription{}, ErrorSubscriptinEventNotFound
	}
	return subscriptions[0], nil
}

// GetSubscription return subscription from part of a event and associationId
func (d *DynamodbClient) GetSubscriptionByAssociationIdAndEvent(associationId, event string) ([]entity.Subscription, error) {
	filt := expression.Key("INDEX_AUXILIAR_PK").Equal(expression.Value(fmt.Sprintf("ASSOCIATION#%s", associationId))).And(expression.Key("INDEX_AUXILIAR_SK").BeginsWith(fmt.Sprintf("SUBSCRIPTION_EVENT#%s", event)))
	expr, err := expression.NewBuilder().WithKeyCondition(filt).Build()
	if err != nil {
		return []entity.Subscription{}, err
	}

	input := &dynamodb.QueryInput{
		IndexName:                 &d.INDEX_AUXILIAR_ASSOCIATION,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		TableName:                 aws.String(d.tableName),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	output, err := d.client.QueryWithContext(ctx, input)
	if err != nil {
		return []entity.Subscription{}, err
	}
	subscriptions := []entity.Subscription{}
	if err := dynamodbattribute.UnmarshalListOfMaps(output.Items, &subscriptions); err != nil {
		return []entity.Subscription{}, err
	}
	if len(subscriptions) == 0 {
		return []entity.Subscription{}, ErrorSubscriptinEventNotFound
	}
	return subscriptions, nil
}

// DeleteSubscription execute remove the event subscription
func (d *DynamodbClient) DeleteSubscription(clientId, event, url string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	sub := entity.Subscription{
		PK: fmt.Sprintf("SUBSCRIPTION#%s", clientId),
		SK: fmt.Sprintf("SUBSCRIPTION_EVENT#%s#%s", event, url),
	}
	item, err := dynamodbattribute.MarshalMap(sub)
	if err != nil {
		return err
	}
	inputDelete := &dynamodb.DeleteItemInput{
		TableName: aws.String(d.tableName),
		Key:       item,
	}
	_, err = d.client.DeleteItemWithContext(ctx, inputDelete)
	if err != nil {
		return err
	}
	return nil
}

// CreateClients execute creation the clients in dynamo table
func (d *DynamodbClient) CreateClients(client dto.ClientDTO) (string, error) {
	secret := os.Getenv("SECRET")
	plainKey := fmt.Sprintf("%s:%s:%s", client.Identifier, client.Service, secret)
	newApiKey, err := utils.GenerateApiKey(plainKey)
	if err != nil {
		return "", err
	}
	newClient := entity.Clients{
		PK:                fmt.Sprintf("CLIENT#%s", newApiKey),
		SK:                fmt.Sprintf("CLIENT_SERVICE#%s", client.Service),
		INDEX_AUXILIAR_PK: fmt.Sprintf("ASSOCIATION#%s", client.AssociationId),
		INDEX_AUXILIAR_SK: fmt.Sprintf("CLIENT_SERVICE#%s", client.Service),
		Identifier:        client.Identifier,
		Service:           client.Service,
		AssociationId:     client.AssociationId,
		Description:       client.Description,
		Provider:          client.Provider,
		CreatedAt:         time.Now().Format("2006-01-02 15:04:05"),
		UpdatedAt:         time.Now().Format("2006-01-02 15:04:05"),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	putItem, err := dynamodbattribute.MarshalMap(newClient)
	if err != nil {
		return "", err
	}
	input := &dynamodb.PutItemInput{
		TableName: aws.String(d.tableName),
		Item:      putItem,
	}
	_, err = d.client.PutItemWithContext(ctx, input)
	if err != nil {
		return "", err
	}
	return newApiKey, nil
}

// CreateScope execute creation the scopes in dynamo table
func (d *DynamodbClient) CreateScope(scope dto.ScopeDTO) error {
	id := uuid.New()
	newScope := entity.Scopes{
		PK:                fmt.Sprintf("API_KEY#%s", scope.ApiKey),
		SK:                fmt.Sprintf("SCOPE#%s", scope.Scope),
		INDEX_AUXILIAR_PK: fmt.Sprintf("CLIENT#%s", scope.Identifier),
		INDEX_AUXILIAR_SK: fmt.Sprintf("SCOPE#%s", scope.Scope),
		Identifier:        scope.Identifier,
		Scope:             scope.Scope,
		ApiKey:            scope.ApiKey,
		Provider:          scope.Provider,
		Service:           scope.Service,
		AssociationId:     scope.AssociationId,
		ScopeId:           id.String(),
		CreatedAt:         time.Now().Format("2006-01-02 15:04:05"),
		UpdatedAt:         time.Now().Format("2006-01-02 15:04:05"),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	putItem, err := dynamodbattribute.MarshalMap(newScope)
	if err != nil {
		return err
	}
	input := &dynamodb.PutItemInput{
		TableName: aws.String(d.tableName),
		Item:      putItem,
	}
	_, err = d.client.PutItemWithContext(ctx, input)
	if err != nil {
		return err
	}
	return nil
}

// ListClients return all clients the table
func (d *DynamodbClient) ListClients() ([]entity.Clients, error) {
	filt := expression.Name("PK").BeginsWith("CLIENT#")
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		return []entity.Clients{}, err
	}

	input := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		TableName:                 aws.String(d.tableName),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	output, err := d.client.ScanWithContext(ctx, input)
	if err != nil {
		return []entity.Clients{}, err
	}
	clients := []entity.Clients{}
	if err := dynamodbattribute.UnmarshalListOfMaps(output.Items, &clients); err != nil {
		return []entity.Clients{}, err
	}

	return clients, nil
}

// GetClientsByClientId return client with service by clientId
func (d *DynamodbClient) GetClientsByClientId(clientId string) (entity.Clients, error) {
	filt := expression.Name("identifier").Equal(expression.Value(clientId)).And(expression.Name("PK").BeginsWith("CLIENT#"))
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		return entity.Clients{}, err
	}
	input := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression: expr.Filter(),
		TableName:        aws.String(d.tableName),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	output, err := d.client.ScanWithContext(ctx, input)
	if err != nil {
		return entity.Clients{}, err
	}
	clients := []entity.Clients{}
	if err := dynamodbattribute.UnmarshalListOfMaps(output.Items, &clients); err != nil {
		return entity.Clients{}, err
	}
	if *output.Count == 0 {
		return entity.Clients{}, ErrorClientNotFound
	}

	return clients[0], nil
}

// GetClients return client with service by apiKey hash
func (d *DynamodbClient) GetClients(apiKey, service string) (entity.Clients, error) {
	filt := expression.And(expression.Name("PK").Equal(expression.Value(fmt.Sprintf("CLIENT#%s", apiKey))), expression.Name("SK").Equal(expression.Value(fmt.Sprintf("CLIENT_SERVICE#%s", service))))
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		return entity.Clients{}, err
	}

	input := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		TableName:                 aws.String(d.tableName),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	output, err := d.client.ScanWithContext(ctx, input)
	if err != nil {
		return entity.Clients{}, err
	}
	clients := []entity.Clients{}
	if err := dynamodbattribute.UnmarshalListOfMaps(output.Items, &clients); err != nil {
		return entity.Clients{}, err
	}
	if *output.Count == 0 {
		return entity.Clients{}, ErrorClientNotFound
	}

	return clients[0], nil
}

// ExisteClient return client by identifier in table
func (d *DynamodbClient) ExistClient(association_id, service string) (entity.Clients, bool, error) {
	filt := expression.Key("INDEX_AUXILIAR_PK").Equal(expression.Value(fmt.Sprintf("ASSOCIATION#%s", association_id))).And(expression.Key("INDEX_AUXILIAR_SK").Equal(expression.Value(fmt.Sprintf("CLIENT_SERVICE#%s", service))))
	expr, err := expression.NewBuilder().WithKeyCondition(filt).Build()
	if err != nil {
		return entity.Clients{}, false, err
	}

	input := &dynamodb.QueryInput{
		IndexName:                 &d.INDEX_AUXILIAR_ASSOCIATION,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		TableName:                 aws.String(d.tableName),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	output, err := d.client.QueryWithContext(ctx, input)
	if err != nil {
		return entity.Clients{}, false, err
	}
	clients := []entity.Clients{}
	if err := dynamodbattribute.UnmarshalListOfMaps(output.Items, &clients); err != nil {
		return entity.Clients{}, false, err
	}
	if *output.Count == 0 {
		return entity.Clients{}, false, ErrorClientNotFound
	}

	return clients[0], true, nil
}

// GetClientByApiKey return client by api key
func (d *DynamodbClient) GetClientByApiKey(apiKey string) (entity.Clients, error) {
	filt := expression.Key("PK").Equal(expression.Value(fmt.Sprintf("CLIENT#%s", apiKey)))
	expr, err := expression.NewBuilder().WithKeyCondition(filt).Build()
	if err != nil {
		return entity.Clients{}, err
	}

	input := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		TableName:                 aws.String(d.tableName),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	output, err := d.client.QueryWithContext(ctx, input)
	if err != nil {
		return entity.Clients{}, err
	}
	clients := []entity.Clients{}
	if err := dynamodbattribute.UnmarshalListOfMaps(output.Items, &clients); err != nil {
		return entity.Clients{}, err
	}
	if *output.Count == 0 {
		return entity.Clients{}, ErrorClientNotFound
	}

	return clients[0], nil
}

// DeleteClients remove client in table
func (d *DynamodbClient) DeleteClients(apiKey, service string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	client := entity.Clients{
		PK: fmt.Sprintf("CLIENT#%s", apiKey),
		SK: fmt.Sprintf("CLIENT_SERVICE#%s", service),
	}
	item, err := dynamodbattribute.MarshalMap(client)
	if err != nil {
		return err
	}
	inputDelete := &dynamodb.DeleteItemInput{
		TableName: aws.String(d.tableName),
		Key:       item,
	}
	_, err = d.client.DeleteItemWithContext(ctx, inputDelete)
	if err != nil {
		return err
	}
	return nil
}

// CreateServices execute creation the services in dynamo table
func (d *DynamodbClient) CreateServices(serv dto.ServicesDTO) (string, error) {
	secret := os.Getenv("SECRET")
	services := []entity.Services{}
	id := uuid.New()
	plainKey := fmt.Sprintf("%s:%s:%s", id, serv.Name, secret)
	apiKey, err := utils.GenerateApiKey(plainKey)
	for _, ev := range serv.Events {
		if err != nil {
			return "", err
		}
		serviceEntity := strings.Split(ev, ".")[1]
		service := entity.Services{
			PK:                fmt.Sprintf("SERVICE#%s", serv.Name),
			SK:                fmt.Sprintf("SERVICE_EVENT#%s", ev),
			INDEX_AUXILIAR_PK: fmt.Sprintf("SERVICE#%s", apiKey),
			INDEX_AUXILIAR_SK: fmt.Sprintf("SERVICE_EVENT#%s", ev),
			ApiKey:            apiKey,
			Name:              serv.Name,
			ServiceEvent:      ev,
			Entity:            serviceEntity,
			ServiceId:         id.String(),
			CreatedAt:         time.Now().Format("2006-01-02 15:04:05"),
			UpdatedAt:         time.Now().Format("2006-01-02 15:04:05"),
		}
		services = append(services, service)
	}
	putItensError := []entity.Services{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	for _, item := range services {
		putItem, err := dynamodbattribute.MarshalMap(item)
		if err != nil {
			d.logger.Error(err.Error())
			putItensError = append(putItensError, item)
			continue
		}
		input := &dynamodb.PutItemInput{
			TableName: aws.String(d.tableName),
			Item:      putItem,
		}
		_, err = d.client.PutItemWithContext(ctx, input)
		if err != nil {
			d.logger.Error(err.Error())
			putItensError = append(putItensError, item)
			continue
		}
	}
	d.logger.Error(fmt.Sprintf("ITENS SERVICES ERRORS: %+v", putItensError))
	return apiKey, nil
}

// ListServices return all services with events
func (d *DynamodbClient) ListServices() ([]entity.Services, error) {
	filt := expression.Name("PK").BeginsWith("SERVICE#")
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		return []entity.Services{}, err
	}

	input := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		TableName:                 aws.String(d.tableName),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	output, err := d.client.ScanWithContext(ctx, input)
	if err != nil {
		return []entity.Services{}, err
	}
	services := []entity.Services{}
	if err := dynamodbattribute.UnmarshalListOfMaps(output.Items, &services); err != nil {
		return []entity.Services{}, err
	}

	return services, nil
}

// GetServices return all events from service
func (d *DynamodbClient) GetServices(serviceName string) ([]entity.Services, error) {
	keyCondition := expression.Key("PK").Equal(expression.Value(fmt.Sprintf("SERVICE#%s", serviceName)))
	expr, err := expression.NewBuilder().WithKeyCondition(keyCondition).Build()
	if err != nil {
		return []entity.Services{}, err
	}

	input := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		TableName:                 aws.String(d.tableName),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	output, err := d.client.QueryWithContext(ctx, input)
	if err != nil {
		return []entity.Services{}, err
	}
	services := []entity.Services{}
	if err := dynamodbattribute.UnmarshalListOfMaps(output.Items, &services); err != nil {
		return []entity.Services{}, err
	}

	return services, nil
}

// GetServicesEvents return service event from table
func (d *DynamodbClient) GetServicesEvents(serviceName, event string) (entity.Services, error) {
	filt := expression.Key("PK").Equal(expression.Value(fmt.Sprintf("SERVICE#%s", serviceName))).And(expression.Key("SK").BeginsWith(fmt.Sprintf("SERVICE_EVENT#%s", event)))
	expr, err := expression.NewBuilder().WithKeyCondition(filt).Build()
	if err != nil {
		return entity.Services{}, err
	}

	input := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		TableName:                 aws.String(d.tableName),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	output, err := d.client.QueryWithContext(ctx, input)
	if err != nil {
		return entity.Services{}, err
	}
	fmt.Printf("output: %+v\n", output)
	services := []entity.Services{}
	if err := dynamodbattribute.UnmarshalListOfMaps(output.Items, &services); err != nil {
		return entity.Services{}, err
	}
	if len(services) == 0 {
		return entity.Services{}, ErrorServiceEventNotFound
	}
	return services[0], nil
}

// GetClientByApiKey return service by api key
func (d *DynamodbClient) GetServiceByApiKey(apiKey string) (entity.Services, error) {
	filt := expression.Key("INDEX_AUXILIAR_PK").Equal(expression.Value(fmt.Sprintf("SERVICE#%s", apiKey)))
	expr, err := expression.NewBuilder().WithKeyCondition(filt).Build()
	if err != nil {
		return entity.Services{}, err
	}

	input := &dynamodb.QueryInput{
		IndexName:                 &d.INDEX_AUXILIAR_ASSOCIATION,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		TableName:                 aws.String(d.tableName),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	output, err := d.client.QueryWithContext(ctx, input)
	if err != nil {
		return entity.Services{}, err
	}
	services := []entity.Services{}
	if err := dynamodbattribute.UnmarshalListOfMaps(output.Items, &services); err != nil {
		return entity.Services{}, err
	}
	if *output.Count == 0 {
		return entity.Services{}, ErrorClientNotFound
	}

	return services[0], nil
}

// DeleteServices execute remotion of events the services
func (d *DynamodbClient) DeleteServices(serviceName, event string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	serv := entity.Services{
		PK: fmt.Sprintf("SERVICE#%s", serviceName),
		SK: fmt.Sprintf("SERVICE_EVENT#%s", event),
	}
	item, err := dynamodbattribute.MarshalMap(serv)
	if err != nil {
		return err
	}
	inputDelete := &dynamodb.DeleteItemInput{
		TableName: aws.String(d.tableName),
		Key:       item,
	}
	_, err = d.client.DeleteItemWithContext(ctx, inputDelete)
	if err != nil {
		return err
	}
	return nil
}

// PutEventService add event to service
func (d *DynamodbClient) PutEventService(serviceName, idService, event string) (interface{}, error) {
	secret := os.Getenv("SECRET")
	serviceEntity := strings.Split(event, ".")[1]
	plainKey := fmt.Sprintf("%s:%s:%s", idService, serviceName, secret)
	apiKey, err := utils.GenerateApiKey(plainKey)
	serviceEvent := entity.Services{
		PK:                fmt.Sprintf("SERVICE#%s", serviceName),
		SK:                fmt.Sprintf("SERVICE_EVENT#%s", event),
		INDEX_AUXILIAR_PK: fmt.Sprintf("SERVICE#%s", idService),
		INDEX_AUXILIAR_SK: fmt.Sprintf("SERVICE_EVENT#%s", event),
		Name:              serviceName,
		ServiceEvent:      event,
		ServiceId:         idService,
		Entity:            serviceEntity,
		ApiKey:            apiKey,
		CreatedAt:         time.Now().Format("2006-01-02 15:04:05"),
		UpdatedAt:         time.Now().Format("2006-01-02 15:04:05"),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	putItem, err := dynamodbattribute.MarshalMap(serviceEvent)
	if err != nil {
		return nil, err
	}
	input := &dynamodb.PutItemInput{
		TableName: aws.String(d.tableName),
		Item:      putItem,
	}
	output, err := d.client.PutItemWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	return output, nil
}

// ExistService return boolean if exist service in table
func (d *DynamodbClient) ExistService(serviceName string) (bool, string, error) {
	filt := expression.Key("PK").Equal(expression.Value(fmt.Sprintf("SERVICE#%s", serviceName)))
	expr, err := expression.NewBuilder().WithKeyCondition(filt).Build()
	if err != nil {
		return false, "", err
	}

	input := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		TableName:                 aws.String(d.tableName),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	output, err := d.client.QueryWithContext(ctx, input)
	if err != nil {
		return false, "", err
	}
	service := entity.Services{}
	if *output.Count > 0 {
		if err := dynamodbattribute.UnmarshalMap(output.Items[0], &service); err != nil {
			return false, "", err
		}
		return true, service.ServiceId, nil
	}
	return false, "", nil
}

// DescribeTable return informations from table
func (d *DynamodbClient) DescribeTable() (interface{}, error) {
	input := &dynamodb.DescribeTableInput{
		TableName: aws.String(d.tableName),
	}
	output, err := d.client.DescribeTable(input)
	if err != nil {
		return nil, err
	}
	return output, nil
}
