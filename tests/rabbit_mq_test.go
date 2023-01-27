package tests

import (
	"fmt"
	"log"
	"testing"

	"github.com/badico-cloud-hub/pubsub/dto"
	"github.com/badico-cloud-hub/pubsub/infra"
	"github.com/joho/godotenv"
)

func TestNewRabbitMQ(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestNewRabbitMQ: expect(nil) - got(%s)\n", err.Error())
	}
	rabbit := infra.NewRabbitMQ()
	if rabbit == nil {
		t.Errorf("TestNewRabbitMQ: expect(!nil) - got(nil)\n")
	}
	log.Printf("rabbit: %+v\n", rabbit)
}
func TestSetup(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestSetup: expect(nil) - got(%s)\n", err.Error())
	}
	rabbit := infra.NewRabbitMQ()
	if rabbit == nil {
		t.Errorf("TestSetup: expect(!nil) - got(nil)\n")
	}
	if err := rabbit.Setup(); err != nil {
		t.Errorf("TestSetup: expect(nil) - got(%s)\n", err.Error())
	}
	rabbit.Release()
}

func TestNumberOfMessagesQueue(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestNumberOfMessagesQueue: expect(nil) - got(%s)\n", err.Error())
	}
	rabbit := infra.NewRabbitMQ()
	if rabbit == nil {
		t.Errorf("TestNumberOfMessagesQueue: expect(!nil) - got(nil)\n")
	}
	if err := rabbit.Setup(); err != nil {
		t.Errorf("TestNumberOfMessagesQueue: expect(nil) - got(%s)\n", err.Error())
	}

	if err := rabbit.NumberOfMessagesQueue(); err != nil {
		t.Errorf("TestNumberOfMessagesQueue: expect(nil) - got(%s)\n", err.Error())
	}
	rabbit.Release()
}

func TestConsumer(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestConsumer: expect(nil) - got(%s)\n", err.Error())
	}
	rabbit := infra.NewRabbitMQ()
	if rabbit == nil {
		t.Errorf("TestConsumer: expect(!nil) - got(nil)\n")
	}
	if err := rabbit.Setup(); err != nil {
		t.Errorf("TestConsumer: expect(nil) - got(%s)\n", err.Error())
	}

	ch, err := rabbit.Consumer()
	if err != nil {
		t.Errorf("TestConsumer: expect(nil) - got(%s)\n", err.Error())
	}
	var forever chan struct{}
	go func() {
		for msg := range ch {
			fmt.Printf("Body: %+v\n", string(msg.Body))
			if err := rabbit.Ack(msg.DeliveryTag); err != nil {
				t.Errorf("TestConsumer: expect(nil) - got(%s)\n", err.Error())

			}
		}
	}()
	<-forever
}

func TestConsumerDlq(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestConsumerDlq: expect(nil) - got(%s)\n", err.Error())
	}
	rabbit := infra.NewRabbitMQ()
	if rabbit == nil {
		t.Errorf("TestConsumerDlq: expect(!nil) - got(nil)\n")
	}
	if err := rabbit.Setup(); err != nil {
		t.Errorf("TestConsumerDlq: expect(nil) - got(%s)\n", err.Error())
	}

	ch, err := rabbit.ConsumerDlq()
	if err != nil {
		t.Errorf("TestConsumerDlq: expect(nil) - got(%s)\n", err.Error())
	}
	var forever chan struct{}
	go func() {
		for msg := range ch {
			fmt.Printf("Body: %+v\n", string(msg.Body))
			if err := rabbit.Ack(msg.DeliveryTag); err != nil {
				t.Errorf("TestConsumerDlq: expect(nil) - got(%s)\n", err.Error())

			}
		}
	}()
	<-forever
}

func TestProducer(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestProducer: expect(nil) - got(%s)\n", err.Error())
	}
	rabbit := infra.NewRabbitMQ()
	if rabbit == nil {
		t.Errorf("TestProducer: expect(!nil) - got(nil)\n")
	}
	if err := rabbit.Setup(); err != nil {
		t.Errorf("TestProducer: expect(nil) - got(%s)\n", err.Error())
	}
	for i := 0; i < 100; i++ {
		fmt.Printf("send: %v\n", i)
		queueMessage := dto.QueueMessage{
			ClientId:      "id@ed",
			Url:           "https://eo4ym5xeg1n1yqu.m.pipedream.net",
			AuthProvider:  "",
			AssociationId: "myassociation@association",
			Body:          map[string]interface{}{},
		}

		if err := rabbit.Producer(queueMessage); err != nil {
			t.Errorf("TestProducer: expect(nil) - got(%s)\n", err.Error())
		}
	}

	rabbit.Release()
}

func TestProducerDlq(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestProducerDlq: expect(nil) - got(%s)\n", err.Error())
	}
	rabbit := infra.NewRabbitMQ()
	if rabbit == nil {
		t.Errorf("TestProducerDlq: expect(!nil) - got(nil)\n")
	}
	if err := rabbit.Setup(); err != nil {
		t.Errorf("TestProducerDlq: expect(nil) - got(%s)\n", err.Error())
	}
	for i := 0; i < 10; i++ {
		fmt.Printf("send: %v\n", i)
		// cashin := dto.CashinDTO{
		// 	Event:         "cashin.pix.create",
		// 	ApiKey:        "0890188fed26c536b6acfa76a0ef552a5e7419e62f7f0e494ff3e533b8ee294f",
		// 	ApiKeyType:    "pix",
		// 	AssociationId: "orion@zemo",
		// 	CashinPayload: dto.CashinPayload{
		// 		Key:         "71900d60-0ec6-4f48-a966-92c8a031787a",
		// 		Value:       "1.50",
		// 		Expiration:  360,
		// 		Description: "test of cashin create",
		// 		Beneficiaries: []dto.CashinPayloaBeneficiaries{
		// 			{
		// 				Document: "30290199000109",
		// 				Value:    1.00,
		// 				Type:     2,
		// 			},
		// 			{
		// 				Document: "00163847339",
		// 				Value:    0.50,
		// 				Type:     2,
		// 			},
		// 		},
		// 	},
		// }

		if err := rabbit.Dlq(dto.QueueMessage{}); err != nil {
			t.Errorf("TestProducerDlq: expect(nil) - got(%s)\n", err.Error())
		}
	}

	rabbit.Release()
}
