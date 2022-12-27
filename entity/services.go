package entity

//Services is entity the services
type Services struct {
	PK           string `dynamodbav:"PK,omitempty"`
	SK           string `dynamodbav:"SK,omitempty"`
	GSIPK        string `dynamodbav:"GSIPK,omitempty"`
	GSISK        string `dynamodbav:"GSISK,omitempty"`
	Name         string `dynamodbav:"name,omitempty"`
	ServiceId    string `dynamodbav:"service_id,omitempty"`
	ServiceEvent string `dynamodbav:"service_event,omitempty"`
	CreatedAt    string `dynamodbav:"createdAt,omitempty"`
	UpdatedAt    string `dynamodbav:"updatedAt,omitempty"`
}
