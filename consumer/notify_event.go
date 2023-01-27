package consumer

import (
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
)

type NotifyEventHandler struct {
	// logManager *producer.LoggerManager
}

type NotifyEventMessageBody struct {
	URL           string                 `json:"url"`
	ClientID      string                 `json:"client_id"`
	AssociationId string                 `json:"association_id"`
	Retries       int                    `json:"retries"`
	AuthProvider  string                 `json:"auth_provider"`
	Body          map[string]interface{} `json:"body"`
}

func NewNotifyEventHandler() *NotifyEventHandler {

	return &NotifyEventHandler{}
}

func (h *NotifyEventHandler) Handle(message ConsumerMessage) (map[string]interface{}, error) {
	// handleLog := h.logManager.NewLogger("logger handle function- ", os.Getenv("MACHINE_IP"))
	// handleLog.AddTraceRef(fmt.Sprintf("MessageId: %s", message.ReceiptHandle))
	// handleLog.Infoln("START")
	retriesNumber := 3
	defer func() {
		// handleLog.Infoln("END")
	}()

	request := resty.New().R()

	if message.QueueMessage.Retries > retriesNumber {
		return nil, errors.New("To many retries")
	}

	// handleLog.AddTraceRef(fmt.Sprintf("ClientID = %s", notifyEventMessageBody.ClientID))
	// handleLog.AddTraceRef(fmt.Sprintf("URL = %s", notifyEventMessageBody.URL))
	// handleLog.AddTraceRef(fmt.Sprintf("EventName = %s", notifyEventMessageBody.Body["topic"]))
	// handleLog.AddTraceRef(fmt.Sprintf("CreatedAt = %s", notifyEventMessageBody.Body["created_at"]))

	if message.QueueMessage.AuthProvider != "" {
		// authProvider, _ := h.getAuthProvider(notifyEventMessageBody.AuthProvider)

		// token, ok := authProvider.Authenticate(notifyEventMessageBody.ClientID)

		// if !ok {
		// 	return nil, errors.New("token not found")
		// }

		// TODO: SetHeader should be custom with the authProvider name
		request.SetHeader("token", "token")
	}

	// handleLog.Infoln("Making Request...")
	fmt.Printf("QueueMessage: %+v\n", *message.QueueMessage)
	resp, err := request.SetBody(message.QueueMessage.Body).Post(message.QueueMessage.Url)

	if err != nil || resp.StatusCode() < 200 || resp.StatusCode() > 299 {
		fmt.Printf("ClientId: %+v, Retries: %+v\n", message.QueueMessage.ClientId, message.QueueMessage.Retries)
		message.QueueMessage.Retries++
		message.handleChannel <- message.QueueMessage
		return nil, nil
	}

	// handleLog.Infof("StatusCode = %s", resp.StatusCode())
	fmt.Println("StatusCode = ", resp.StatusCode())

	return nil, nil
}
