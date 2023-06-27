package entity

//Scope is entity the clients
type Scopes struct {
	PK                string `dynamodbav:"PK,omitempty"`
	SK                string `dynamodbav:"SK,omitempty"`
	INDEX_AUXILIAR_PK string `dynamodbav:"INDEX_AUXILIAR_PK,omitempty"`
	INDEX_AUXILIAR_SK string `dynamodbav:"INDEX_AUXILIAR_SK,omitempty"`
	Identifier        string `dynamodbav:"identifier,omitempty"`
	Scope             string `dynamodbav:"scope,omitempty"`
	ApiKey            string `dynamodbav:"api_key,omitempty"`
	ScopeId           string `dynamodbav:"scope_id,omitempty"`
	Provider          string `dynamodbav:"provider,omitempty"`
	AssociationId     string `dynamodbav:"association_id,omitempty"`
	CreatedAt         string `dynamodbav:"createdAt,omitempty"`
	UpdatedAt         string `dynamodbav:"updatedAt,omitempty"`
}
