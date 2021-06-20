package acct

import "context"

// Accounter defines methods for interacting with user account
// values stored in the application.
type Accounter interface {
	CreateAccount(ctx context.Context, accountID string) error
	ReadAccount(ctx context.Context, accountID string) (*Account, error)
	UpdateAccount(ctx context.Context, accountID string, values map[string]string) error
	DeleteAccount(ctx context.Context, accountID string) error
}

// Account represents the user information stored in the application.
type Account struct {
	ID                       string
	SubscriptionID           string
	StripePaymentMethodID    string
	StripeCustomerID         string
	StripeSubscriptionID     string
	StripeSubscriptionItemID string
}
