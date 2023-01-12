package entity

//Subscription is entity the subscription
type Subscription struct {
	PK                string `dynamodbav:"PK,omitempty"`
	SK                string `dynamodbav:"SK,omitempty"`
	GSIPK             string `dynamodbav:"GSIPK,omitempty"`
	GSISK             string `dynamodbav:"GSISK,omitempty"`
	ClientId          string `dynamodbav:"client_id,omitempty"`
	SubscriptionEvent string `dynamodbav:"subscription_event,omitempty"`
	SubscriptionUrl   string `dynamodbav:"subscription_url,omitempty"`
	SubscriptionId    string `dynamodbav:"subscription_id,omitempty"`
	AssociationId     string `dynamodbav:"association_id,omitempty"`
	CreatedAt         string `dynamodbav:"createdAt,omitempty"`
	UpdatedAt         string `dynamodbav:"updatedAt,omitempty"`
}
