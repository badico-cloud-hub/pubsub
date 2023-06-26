package entity

//Clients is entity the services
type Clients struct {
	PK                string `dynamodbav:"PK,omitempty"`
	SK                string `dynamodbav:"SK,omitempty"`
	INDEX_AUXILIAR_PK string `dynamodbav:"INDEX_AUXILIAR_PK,omitempty"`
	INDEX_AUXILIAR_SK string `dynamodbav:"INDEX_AUXILIAR_SK,omitempty"`
	Identifier        string `dynamodbav:"identifier,omitempty"`
	Service           string `dynamodbav:"service,omitempty"`
	ApiKey            string `dynamodbav:"api_key,omitempty"`
	AssociationId     string `dynamodbav:"association_id,omitempty"`
	Description       string `dynamodbav:"description,omitempty"`
	Provider          string `dynamodbav:"provider,omitempty"`
	CreatedAt         string `dynamodbav:"createdAt,omitempty"`
	UpdatedAt         string `dynamodbav:"updatedAt,omitempty"`
}
