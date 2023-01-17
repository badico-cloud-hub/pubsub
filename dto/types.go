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
	AssociationId     string   `json:"association_id,omitempty"`
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

//ServiceDTO is struct for dto the service
type ServicesDTO struct {
	Name      string   `json:"name,omitempty"`
	Events    []string `json:"events,omitempty"`
	ServiceId string   `json:"service_id,omitempty"`
	ApiKey    string   `json:"api_key,omitempty"`
	Entity    string   `json:"entity,omitempty"`
	CreatedAt string   `json:"created_at,omitempty"`
}

//ServiceEventsDTO is struct for dto the service events
type ServiceEventsDTO struct {
	Name      string `json:"name,omitempty"`
	Service   string `json:"service,omitempty"`
	ServiceId string `json:"service_id,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

//ClientDTO is struct for dto the clients
type ClientDTO struct {
	ApiKey        string `json:"api_key,omitempty"`
	Identifier    string `json:"identifier,omitempty"`
	Service       string `json:"service,omitempty"`
	AssociationId string `json:"association_id,omitempty"`
	CreatedAt     string `json:"created_at,omitempty"`
	UpdatedAt     string `json:"updated_at,omitempty"`
}

//AdminObject is struct the api key of admin
type AdminObject struct {
	ClientId string `json:"client_id"`
	ApiKey   string `json:"api_key"`
}
