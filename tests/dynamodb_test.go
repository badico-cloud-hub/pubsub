package tests

import (
	"log"
	"os"
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
	tableName := os.Getenv("BATCH_TABLE")
	result, err := dynamo.DescribeTable(tableName)
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
	if err := dynamo.DeleteSubscription("id@tom", "cd5a4396-8235-40a1-8203-aae37a285cf9", "batch-item.pix.paid"); err != nil {
		t.Errorf("TestDeleteSubscriptions: expect(nil) - got(%s)\n", err.Error())
	}
}
