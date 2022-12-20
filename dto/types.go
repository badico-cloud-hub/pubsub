package dto

//ResponseDTO is struct for response default
type ResponseDTO struct {
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
}

//ResponseNotifyDTO is struct for response the notify
type ResponseNotifyDTO struct {
	Ok     bool              `json:"ok"`
	Topic  string            `json:"topic"`
	SentTo []SubscriptionDTO `json:"sent_to"`
}

//SqsMessage is struct of message to sqs
type SqsMessage struct {
	ClientId     string                 `json:"client_id"`
	Url          string                 `json:"url"`
	AuthProvider string                 `json:"auth_provider,omitempty"`
	Body         map[string]interface{} `json:"body"`
}

//SubscriptionDTO is struct for dto the subscription
type SubscriptionDTO struct {
	ClientId          string   `json:"client_id,omitempty"`
	Events            []string `json:"events,omitempty"`
	Url               string   `json:"url,omitempty"`
	AuthProvider      string   `json:"authProvider,omitempty"`
	SubscriptionUrl   string   `json:"subscription_url,omitempty"`
	SubscriptionId    string   `json:"subscription_id,omitempty"`
	SubscriptionEvent string   `json:"subscription_event,omitempty"`
	CreatedAt         string   `json:"createdAt,omitempty"`
}

//NotifierDTO is struct for dto the notify
type NotifierDTO struct {
	ClientId  string                 `json:"client_id,omitempty"`
	Event     string                 `json:"event,omitempty"`
	Url       string                 `json:"url,omitempty"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	CreatedAt string                 `json:"createdAt,omitempty"`
}
