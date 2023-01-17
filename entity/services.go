package entity

//Services is entity the services
type Services struct {
	PK                string `dynamodbav:"PK,omitempty"`
	SK                string `dynamodbav:"SK,omitempty"`
	INDEX_AUXILIAR_PK string `dynamodbav:"INDEX_AUXILIAR_PK,omitempty"`
	INDEX_AUXILIAR_SK string `dynamodbav:"INDEX_AUXILIAR_SK,omitempty"`
	ApiKey            string `dynamodbav:"api_key,omitempty"`
	Name              string `dynamodbav:"name,omitempty"`
	ServiceId         string `dynamodbav:"service_id,omitempty"`
	ServiceEvent      string `dynamodbav:"service_event,omitempty"`
	Entity            string `dynamodbav:"entity,omitempty"`
	CreatedAt         string `dynamodbav:"createdAt,omitempty"`
	UpdatedAt         string `dynamodbav:"updatedAt,omitempty"`
}
