package entity

//Subscription is entity the subscription
type Subscription struct {
	PK                string `dynamodbav:"PK,omitempty"`
	SK                string `dynamodbav:"SK,omitempty"`
	INDEX_AUXILIAR_PK string `dynamodbav:"INDEX_AUXILIAR_PK,omitempty"`
	INDEX_AUXILIAR_SK string `dynamodbav:"INDEX_AUXILIAR_SK,omitempty"`
	ClientId          string `dynamodbav:"client_id,omitempty"`
	SubscriptionEvent string `dynamodbav:"subscription_event,omitempty"`
	SubscriptionUrl   string `dynamodbav:"subscription_url,omitempty"`
	SubscriptionId    string `dynamodbav:"subscription_id,omitempty"`
	AssociationId     string `dynamodbav:"association_id,omitempty"`
	CreatedAt         string `dynamodbav:"createdAt,omitempty"`
	UpdatedAt         string `dynamodbav:"updatedAt,omitempty"`
}
