package tests

import (
	"log"
	"testing"
	"time"

	"github.com/badico-cloud-hub/pubsub/dto"
	"github.com/badico-cloud-hub/pubsub/infra"
	"github.com/joho/godotenv"
)

func TestNewSqsCliente(t *testing.T) {
	sqs := infra.NewSqsClient()
	if sqs == nil {
		t.Errorf("TestNewSqsClient: expect(!=nil) - got(nil)\n")
	}
}

func TestSetupSqs(t *testing.T) {
	sqs := infra.NewSqsClient()
	if err := sqs.Setup(); err != nil {
		t.Errorf("TestSetupSqs: expect(nil) - got(%s)\n", err.Error())
	}
	log.Printf("sqs client: %+v\n", sqs)
}

func TestSendSqs(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestSendSqs: expect(nil) - got(%s)\n", err.Error())
	}
	sqs := infra.NewSqsClient()
	if err := sqs.Setup(); err != nil {
		t.Errorf("TestSendSqs: expect(nil) - got(%s)\n", err.Error())
	}
	if err := sqs.Setup(); err != nil {
		t.Errorf("TestSendSqs: expect(nil) - got(%s)\n", err.Error())
	}
	subs := dto.SubscriptionDTO{
		ClientId:          "id@tom",
		AuthProvider:      "",
		SubscriptionUrl:   "https://eoudkrtxlcsevnd.m.pipedream.net",
		SubscriptionEvent: "batch-item.pix.paid",
	}
	notif := dto.NotifierDTO{
		Payload: map[string]interface{}{
			"invoice": "invoice.paid",
		},
		CreatedAt: time.Now().UTC().String(),
	}
	output, err := sqs.Send(subs, notif)
	if err != nil {
		t.Errorf("TestSendSqs: expect(nil) - got(%s)\n", err.Error())
	}
	log.Printf("output: %+v\n", output)
}
