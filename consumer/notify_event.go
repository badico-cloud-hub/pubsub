package consumer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/badico-cloud-hub/pubsub/dto"
	"github.com/badico-cloud-hub/pubsub/infra"

	"github.com/badico-cloud-hub/log-driver/producer"
	"github.com/go-resty/resty/v2"
)

type NotifyEventHandler struct {
	logManager     *producer.LoggerManager
	rabbitMqClient *infra.RabbitMQ
}

type NotifyEventMessageBody struct {
	URL           string                 `json:"url"`
	ClientID      string                 `json:"client_id"`
	AssociationId string                 `json:"association_id"`
	Retries       int                    `json:"retries"`
	AuthProvider  string                 `json:"auth_provider"`
	Body          map[string]interface{} `json:"body"`
}

func NewNotifyEventHandler(logManager *producer.LoggerManager, rabbitMqClient *infra.RabbitMQ) *NotifyEventHandler {

	return &NotifyEventHandler{
		logManager,
		rabbitMqClient,
	}
}

func (h *NotifyEventHandler) Handle(message ConsumerMessage) (map[string]interface{}, error) {
	handleLog := h.logManager.NewLogger("logger handle function- ", os.Getenv("MACHINE_IP"))
	handleLog.Infoln("=======================================")
	handleLog.Infoln("START HANDLE MESSAGE")
	handleLog.Infoln("=======================================")
	retriesNumber := 3
	defer func() {
		handleLog.Infoln("END")
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	request := resty.New().R()
	request.SetContext(ctx)
	if message.QueueMessage.Retries > retriesNumber {
		return nil, errors.New("To many retries")
	}

	handleLog.AddTraceRef(fmt.Sprintf("ClientID = %s", message.QueueMessage.ClientId))
	handleLog.AddTraceRef(fmt.Sprintf("URL = %s", message.QueueMessage.Url))
	handleLog.AddTraceRef(fmt.Sprintf("EventName = %s", message.QueueMessage.Body["topic"]))
	handleLog.AddTraceRef(fmt.Sprintf("CashinId = %s", message.QueueMessage.Body["cashin_id"]))
	handleLog.AddTraceRef(fmt.Sprintf("CreatedAt = %s", message.QueueMessage.Body["created_at"]))

	if message.QueueMessage.AuthProvider != "" {
		// authProvider, _ := h.getAuthProvider(notifyEventMessageBody.AuthProvider)

		// token, ok := authProvider.Authenticate(notifyEventMessageBody.ClientID)

		// if !ok {
		// 	return nil, errors.New("token not found")
		// }

		// TODO: SetHeader should be custom with the authProvider name
		request.SetHeader("token", "token")
	}

	handleLog.Infoln("=======================================")
	handleLog.Infof("Making Request in %s with the message queue: %+v\n", time.Now().Format("2006-01-02T15:04:05.000"), *message.QueueMessage)
	handleLog.Infoln("=======================================")
	fmt.Printf("QueueMessage: %+v\n", *message.QueueMessage)
	resp, err := request.SetBody(message.QueueMessage.Body).Post(message.QueueMessage.Url)

	callbackType := message.QueueMessage.Callback["type"].(string)

	callbackMessage := dto.CallbackMessage{
		Event:           message.QueueMessage.Body["topic"].(string),
		Payload:         message.QueueMessage.Body,
		ClientId:        message.QueueMessage.ClientId,
		CashinId:        message.QueueMessage.Body["cashin_id"].(string),
		DeliveredStatus: "SUCCESS",
		DeliveredAt:     time.Now().Format("2006-01-02T15:04:05.000"),
		DeliveredUrl:    message.QueueMessage.Url,
		ErrorMessage:    "",
		StatusCode:      resp.StatusCode(),
	}

	if err != nil || resp.StatusCode() < 200 || resp.StatusCode() > 299 {
		callbackMessage.DeliveredStatus = "ERROR"
		callbackMessage.ErrorMessage = err.Error()
		callbackMessage.StatusCode = resp.StatusCode()
		handleLog.Infoln("=======================================")
		handleLog.Infof("ClientId: %+v, Retries: %+v\n", message.QueueMessage.ClientId, message.QueueMessage.Retries)
		handleLog.Infoln("=======================================")
		if callbackType == "queue.rabbitmq" {
			handleLog.Infoln("Notifing callback queue...")
			err = h.rabbitMqClient.ProducerCallback(callbackMessage)
			if err != nil {
				handleLog.Infof("Error sending callback message: %+v\n", err.Error())
				return nil, err
			}
		}
		message.QueueMessage.Retries++
		message.handleChannel <- message.QueueMessage
		return nil, nil
	}

	handleLog.Infoln("=======================================")
	handleLog.Infof("StatusCode = %s\n", resp.StatusCode())
	handleLog.Infoln("=======================================")
	if callbackType == "queue.rabbitmq" {
		handleLog.Infoln("Notifing callback queue...")
		err = h.rabbitMqClient.ProducerCallback(callbackMessage)
		if err != nil {
			handleLog.Infof("Error sending callback message: %+v\n", err.Error())
			return nil, err
		}
	}
	fmt.Println("StatusCode = ", resp.StatusCode())

	return nil, nil
}
