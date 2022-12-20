package entity

//Notify is entity the notification
type Notify struct {
	ClientId  string      `json:"clientId,omitempty"`
	Name      string      `json:"name,omitempty"`
	Payload   interface{} `json:"payload,omitempty"`
	CreatedAt string      `json:"createdAt,omitempty"`
}
