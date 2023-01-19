package consumer

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/go-resty/resty/v2"
	"github.com/mitchellh/mapstructure"
)

type NotifyEventHandler struct {
}

type NotifyEventMessageBody struct {
	URL           string                 `json:"url"`
	ClientID      string                 `json:"client_id"`
	AssociationId string                 `json:"association_id"`
	Retries       int64                  `json:"retries"`
	AuthProvider  string                 `json:"auth_provider"`
	Body          map[string]interface{} `json:"body"`
}

func NewNotifyEventHandler() *NotifyEventHandler {

	return &NotifyEventHandler{}
}

func (h *NotifyEventHandler) Handle(message ConsumerMessage) (map[string]interface{}, error) {
	// logger := infra.NewLogAggregator("Notify Event")
	fmt.Println("START")
	// logger.Aggregate("START")
	// defer func() {
	// 	logger.Aggregate("END")
	// 	logger.Log()
	// }()

	request := resty.New().R()

	notifyEventMessageBody := h.decodeMessageBody(message)
	notifyBytes, _ := json.Marshal(notifyEventMessageBody)
	notifyMessage := string(notifyBytes)
	sqsMessage := &sqs.Message{
		Body:          &notifyMessage,
		ReceiptHandle: &message.ReceiptHandle,
	}

	if notifyEventMessageBody.Retries > 3 {

		message.removeChannel <- sqsMessage

		return nil, nil
	}

	// logger.Aggregate("ClientID = ", notifyEventMessageBody.ClientID)
	// logger.Aggregate("URL = ", notifyEventMessageBody.URL)
	// logger.Aggregate("EventName = ", notifyEventMessageBody.Body["topic"])
	// logger.Aggregate("CreatedAt = ", notifyEventMessageBody.Body["created_at"])

	if notifyEventMessageBody.AuthProvider != "" {
		// authProvider, _ := h.getAuthProvider(notifyEventMessageBody.AuthProvider)

		// token, ok := authProvider.Authenticate(notifyEventMessageBody.ClientID)

		// if !ok {
		// 	return nil, errors.New("token not found")
		// }

		// TODO: SetHeader should be custom with the authProvider name
		request.SetHeader("token", "token")
	}

	// logger.Aggregate("Making Request...")
	// _, err := request.SetBody(notifyEventMessageBody.Body).Post(notifyEventMessageBody.URL)

	// if err != nil {
	// 	fmt.Printf("MessageId: %+v, Retries: %+v\n", message.ReceiptHandle, notifyEventMessageBody.Retries)
	// 	notifyEventMessageBody.Retries++
	// 	notifyBytes, _ = json.Marshal(notifyEventMessageBody)
	// 	notifyMessage = string(notifyBytes)
	// 	sqsMessage = &sqs.Message{
	// 		Body:          &notifyMessage,
	// 		ReceiptHandle: &message.ReceiptHandle,
	// 	}
	// 	message.handleChannel <- sqsMessage
	// 	return nil, nil
	// }

	// if err != nil {
	// 	// logger.Aggregate("Error = ", err.Error())
	// 	return nil, nil
	// }

	// logger.Aggregate("StatusCode = ", resp.StatusCode())

	return nil, nil
}

func (h *NotifyEventHandler) decodeMessageBody(message ConsumerMessage) NotifyEventMessageBody {
	var notifyEventMessageBody NotifyEventMessageBody

	decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata: nil,
		Result:   &notifyEventMessageBody,
		TagName:  "json",
	})

	decoder.Decode(message.Body)

	return notifyEventMessageBody
}
