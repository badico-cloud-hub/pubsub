package tests

import (
	"log"
	"testing"

	"github.com/badico-cloud-hub/pubsub/dto"
	"github.com/badico-cloud-hub/pubsub/infra"
	"github.com/joho/godotenv"
)

func TestSetupDynamodb(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestSetupDynamodb: expect(nil) - got(%s)\n", err.Error())
	}
	dynamo := infra.NewDynamodbClient()
	if err := dynamo.Setup(); err != nil {
		t.Errorf("TestSetupDynamodb: expect(nil) - got(%s)\n", err.Error())
	}
}

func TestDescribeTable(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestDescribeTable: expect(nil) - got(%s)\n", err.Error())
	}
	dynamo := infra.NewDynamodbClient()
	if err := dynamo.Setup(); err != nil {
		t.Errorf("TestDescribeTable: expect(nil) - got(%s)\n", err.Error())
	}
	result, err := dynamo.DescribeTable()
	if err != nil {
		t.Errorf("TestDescribeTable: expect(nil) - got(%s)\n", err.Error())
	}
	log.Println(result)
}

func TestGetSubscription(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestGetSubscription: expect(nil) - got(%s)\n", err.Error())
	}
	dynamo := infra.NewDynamodbClient()
	if err := dynamo.Setup(); err != nil {
		t.Errorf("TestGetSubscription: expect(nil) - got(%s)\n", err.Error())
	}
	subscriptionEvent, err := dynamo.GetSubscription("id@tom", "batch-item.pix.paid")
	if err != nil {
		t.Errorf("TestGetSubscription: expect(nil) - got(%s)\n", err.Error())
	}
	log.Printf("subcription: %+v\n", subscriptionEvent)
}

func TestListSubscriptions(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestListSubscriptions: expect(nil) - got(%s)\n", err.Error())
	}
	dynamo := infra.NewDynamodbClient()
	if err := dynamo.Setup(); err != nil {
		t.Errorf("TestListSubscriptions: expect(nil) - got(%s)\n", err.Error())
	}
	subs, err := dynamo.ListSubscriptions("id@tom")
	if err != nil {
		t.Errorf("TestListSubscriptions: expect(nil) - got(%s)\n", err.Error())
	}
	log.Printf("%+v\n", subs)
}

func TestCreateSubscriptions(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestCreateSubscriptions: expect(nil) - got(%s)\n", err.Error())
	}
	dynamo := infra.NewDynamodbClient()
	if err := dynamo.Setup(); err != nil {
		t.Errorf("TestCreateSubscriptions: expect(nil) - got(%s)\n", err.Error())
	}
	sub := dto.SubscriptionDTO{
		ClientId: "id@tom",
		Url:      "https://eoudkrtxlcsevnd.m.pipedream.net",
		Events: []string{"batch.items-creating",
			"batch.creating",
			"batch-item.pix.qrcodecreated",
			"batch-item.pix.updated",
			"batch-item.pix.paid"},
	}
	result, err := dynamo.CreateSubscription(&sub)
	if err != nil {
		t.Errorf("TestCreateSubscriptions: expect(nil) - got(%s)\n", err.Error())
	}
	log.Printf("result: %+v\n", result)
}

func TestDeleteSubscriptions(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestDeleteSubscriptions: expect(nil) - got(%s)\n", err.Error())
	}
	dynamo := infra.NewDynamodbClient()
	if err := dynamo.Setup(); err != nil {
		t.Errorf("TestDeleteSubscriptions: expect(nil) - got(%s)\n", err.Error())
	}
	if err := dynamo.DeleteSubscription("id@tom", "batch-item.pix.paid", "google.com"); err != nil {
		t.Errorf("TestDeleteSubscriptions: expect(nil) - got(%s)\n", err.Error())
	}
}

func TestCreateServices(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestCreateServices: expect(nil) - got(%s)\n", err.Error())
	}
	dynamo := infra.NewDynamodbClient()
	if err := dynamo.Setup(); err != nil {
		t.Errorf("TestCreateServices: expect(nil) - got(%s)\n", err.Error())
	}
	service := dto.ServicesDTO{
		Name:   "contractor",
		Events: []string{"contractor.billets.created", "contractor.billets.paid"},
	}
	output, err := dynamo.CreateServices(service)
	if err != nil {
		t.Errorf("TestCreateServices: expect(nil) - got(%s)\n", err.Error())
	}
	log.Printf("output: %+v\n", output)
}

func TestGetServices(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestGetServices: expect(nil) - got(%s)\n", err.Error())
	}
	dynamo := infra.NewDynamodbClient()
	if err := dynamo.Setup(); err != nil {
		t.Errorf("TestGetServices: expect(nil) - got(%s)\n", err.Error())
	}
	output, err := dynamo.GetServices("contractor")
	if err != nil {
		t.Errorf("TestGetServices: expect(nil) - got(%s)\n", err.Error())
	}
	log.Printf("output: %+v\n", output)
}

func TestGetServicesEvents(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestGetServicesEvents: expect(nil) - got(%s)\n", err.Error())
	}
	dynamo := infra.NewDynamodbClient()
	if err := dynamo.Setup(); err != nil {
		t.Errorf("TestGetServicesEvents: expect(nil) - got(%s)\n", err.Error())
	}
	output, err := dynamo.GetServicesEvents("contractor", "contractor.billets.created")
	if err != nil {
		t.Errorf("TestGetServicesEvents: expect(nil) - got(%s)\n", err.Error())
	}
	log.Printf("output: %+v\n", output)
}

func TestListServices(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestListServices: expect(nil) - got(%s)\n", err.Error())
	}
	dynamo := infra.NewDynamodbClient()
	if err := dynamo.Setup(); err != nil {
		t.Errorf("TestListServices: expect(nil) - got(%s)\n", err.Error())
	}
	output, err := dynamo.ListServices()
	if err != nil {
		t.Errorf("TestListServices: expect(nil) - got(%s)\n", err.Error())
	}
	log.Printf("output: %+v\n", output)
}

func TestDeleteServices(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestDeleteServices: expect(nil) - got(%s)\n", err.Error())
	}
	dynamo := infra.NewDynamodbClient()
	if err := dynamo.Setup(); err != nil {
		t.Errorf("TestDeleteServices: expect(nil) - got(%s)\n", err.Error())
	}
	if err := dynamo.DeleteServices("693bd480-b40b-47ef-9e39-289d223df2dc", "contractor.billets.created"); err != nil {
		t.Errorf("TestDeleteServices: expect(nil) - got(%s)\n", err.Error())
	}
}

func TestExistServices(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestExistServices: expect(nil) - got(%s)\n", err.Error())
	}
	dynamo := infra.NewDynamodbClient()
	if err := dynamo.Setup(); err != nil {
		t.Errorf("TestExistServices: expect(nil) - got(%s)\n", err.Error())
	}
	exists, _, err := dynamo.ExistService("contractor")
	if err != nil {
		t.Errorf("TestExistServices: expect(nil) - got(%s)\n", err.Error())
	}
	log.Printf("exists: %+v\n", exists)
}

func TestPutEventService(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestPutEventService: expect(nil) - got(%s)\n", err.Error())
	}
	dynamo := infra.NewDynamodbClient()
	if err := dynamo.Setup(); err != nil {
		t.Errorf("TestPutEventService: expect(nil) - got(%s)\n", err.Error())
	}
	output, err := dynamo.PutEventService("contractor", "fc48b531-1216-4bf3-a2fc-24b4d1a28ffe", "contractor.billets.created")
	if err != nil {
		t.Errorf("TestPutEventService: expect(nil) - got(%s)\n", err.Error())
	}
	log.Printf("output: %+v\n", output)
}

func TestCreateClients(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestCreateClients: expect(nil) - got(%s)\n", err.Error())
	}
	dynamo := infra.NewDynamodbClient()
	if err := dynamo.Setup(); err != nil {
		t.Errorf("TestCreateClients: expect(nil) - got(%s)\n", err.Error())
	}
	client := dto.ClientDTO{
		Identifier: "id@tom",
		Service:    "contractor",
	}
	output, err := dynamo.CreateClients(client)
	if err != nil {
		t.Errorf("TestCreateClients: expect(nil) - got(%s)\n", err.Error())
	}
	log.Printf("output: %+v\n", output)
}

func TestListClients(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestListClients: expect(nil) - got(%s)\n", err.Error())
	}
	dynamo := infra.NewDynamodbClient()
	if err := dynamo.Setup(); err != nil {
		t.Errorf("TestListClients: expect(nil) - got(%s)\n", err.Error())
	}
	clients, err := dynamo.ListClients()
	if err != nil {
		t.Errorf("TestListClients: expect(nil) - got(%s)\n", err.Error())
	}
	log.Printf("clients: %+v\n", clients)
}

func TestGetClients(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestGetClients: expect(nil) - got(%s)\n", err.Error())
	}
	dynamo := infra.NewDynamodbClient()
	if err := dynamo.Setup(); err != nil {
		t.Errorf("TestGetClients: expect(nil) - got(%s)\n", err.Error())
	}
	output, err := dynamo.GetClients("b20d74b1d5a8420c0d696de0df5f49836e709d8ab00d778109a52c0c75546ddb", "contractor")
	if err != nil {
		t.Errorf("TestGetClients: expect(nil) - got(%s)\n", err.Error())
	}
	log.Printf("output: %+v\n", output)
}

func TestExistClient(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestExistClient: expect(nil) - got(%s)\n", err.Error())
	}
	dynamo := infra.NewDynamodbClient()
	if err := dynamo.Setup(); err != nil {
		t.Errorf("TestExistClient: expect(nil) - got(%s)\n", err.Error())
	}
	output, exists, err := dynamo.ExistClient("id@tom", "contractor")
	if err != nil {
		t.Errorf("TestExistClient: expect(nil) - got(%s)\n", err.Error())
	}
	log.Printf("output: %+v\n", output)
	log.Printf("exists: %+v\n", exists)
}

func TestDeleteClients(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestDeleteClients: expect(nil) - got(%s)\n", err.Error())
	}
	dynamo := infra.NewDynamodbClient()
	if err := dynamo.Setup(); err != nil {
		t.Errorf("TestDeleteClients: expect(nil) - got(%s)\n", err.Error())
	}
	if err := dynamo.DeleteClients("5f5387781bdf49ea0afdbc3277fc83fb3eff5b86d518e062e94d5107ec0eb95f", "pagadoria"); err != nil {
		t.Errorf("TestDeleteClients: expect(nil) - got(%s)\n", err.Error())
	}
}

func TestGetClientByApiKey(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestGetClientByApiKey: expect(nil) - got(%s)\n", err.Error())
	}
	dynamo := infra.NewDynamodbClient()
	if err := dynamo.Setup(); err != nil {
		t.Errorf("TestGetClientByApiKey: expect(nil) - got(%s)\n", err.Error())
	}
	client, err := dynamo.GetClientByApiKey("26b997a2428ae17b4a5c5f502910f41e52cac237f9dbff93fac4dff626ed1985")
	if err != nil {
		t.Errorf("TestGetClientByApiKey: expect(nil) - got(%s)\n", err.Error())
	}
	log.Printf("Client: %+v\n", client)
}
