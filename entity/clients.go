package entity

//Clients is entity the services
type Clients struct {
	PK            string `dynamodbav:"PK,omitempty"`
	SK            string `dynamodbav:"SK,omitempty"`
	GSIPK         string `dynamodbav:"GSIPK,omitempty"`
	GSISK         string `dynamodbav:"GSISK,omitempty"`
	Identifier    string `dynamodbav:"identifier,omitempty"`
	Service       string `dynamodbav:"service,omitempty"`
	ApiKey        string `dynamodbav:"api_key,omitempty"`
	AssociationId string `dynamodbav:"association_id,omitempty"`
	CreatedAt     string `dynamodbav:"createdAt,omitempty"`
	UpdatedAt     string `dynamodbav:"updatedAt,omitempty"`
}
