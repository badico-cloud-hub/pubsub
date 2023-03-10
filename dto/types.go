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

//QueueMessage is struct of message to sqs
type QueueMessage struct {
	ClientId      string                 `json:"client_id"`
	Url           string                 `json:"url"`
	AuthProvider  string                 `json:"auth_provider,omitempty"`
	AssociationId string                 `json:"association_id,omitempty"`
	Retries       int                    `json:"retries,omitempty"`
	Callback      map[string]interface{} `json:"callback,omitempty"`
	Body          map[string]interface{} `json:"body"`
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
	ClientId       string                 `json:"client_id,omitempty"`
	Event          string                 `json:"event,omitempty"`
	Url            string                 `json:"url,omitempty"`
	Data           map[string]interface{} `json:"data,omitempty"`
	AssociationsId []string               `json:"associations_id,omitempty"`
	Callback       map[string]interface{} `json:"callback,omitempty"`
	CreatedAt      string                 `json:"createdAt,omitempty"`
}

//CallbackMessage is struct for callback message
type CallbackMessage struct {
	Event           string                 `json:"event,omitempty"`
	Payload         map[string]interface{} `json:"payload,omitempty"`
	ClientId        string                 `json:"client_id,omitempty"`
	CashinId        string                 `json:"cashin_id,omitempty"`
	DeliveredStatus string                 `json:"delivered_status,omitempty"`
	DeliveredAt     string                 `json:"delivered_at,omitempty"`
	DeliveredUrl    string                 `json:"delivered_url,omitempty"`
	ErrorMessage    string                 `json:"error_message,omitempty"`
	StatusCode      int                    `json:"status_code,omitempty"`
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

type ErrorMessage struct {
	Source        string                 `json:"source,omitempty"`
	Input         map[string]interface{} `json:"input,omitempty"`
	Reason        string                 `json:"reason"`
	Output        map[string]interface{} `json:"output,omitempty"`
	SourceMessage *QueueMessage          `json:"-"`
}
