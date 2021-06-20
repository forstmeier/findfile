package subscr

import "context"

// Subscriber defines the methods for interacting with the
// subscription management system.
type Subscriber interface {
	CreateSubscription(ctx context.Context, info SubscriberInfo) (*Subscription, error)
	RemoveSubscription(ctx context.Context, id string) error
	AddUsage(ctx context.Context, id string) error
}

// SubscriberInfo contains the user information required to
// configure the subscription.
type SubscriberInfo struct {
	ID               string
	Email            string `json:"email"`
	ZIP              string `json:"zip"`
	ExpirationMonth  string `json:"expiration_month"`
	ExpirationYear   string `json:"expiration_year"`
	CardNumber       string `json:"card_number"`
	CardSecurityCode string `json:"card_security_code"`
}

// Subscription contains the output of the CreateSubscription
// method.
type Subscription struct {
	ID                       string
	StripePaymentMethodID    string
	StripeCustomerID         string
	StripeSubscriptionID     string
	StripeSubscriptionItemID string
}
