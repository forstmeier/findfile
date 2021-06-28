package subscr

import (
	"context"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/paymentmethod"
	"github.com/stripe/stripe-go/v72/sub"
	"github.com/stripe/stripe-go/v72/usagerecord"
)

var _ Subscriber = &Client{}

// Client implements the Subscriber methods using Stripe.
type Client struct {
	newPaymentMethod func(params *stripe.PaymentMethodParams) (*stripe.PaymentMethod, error)
	newCustomer      func(params *stripe.CustomerParams) (*stripe.Customer, error)
	newSubscription  func(params *stripe.SubscriptionParams) (*stripe.Subscription, error)
	deleteCustomer   func(id string, params *stripe.CustomerParams) (*stripe.Customer, error)
	newUsageRecord   func(params *stripe.UsageRecordParams) (*stripe.UsageRecord, error)
}

// New generates a Client pointer instance with a Stripe
// session client.
func New(stripeAPIKey string) *Client {
	stripe.Key = stripeAPIKey

	return &Client{
		newPaymentMethod: paymentmethod.New,
		newCustomer:      customer.New,
		newSubscription:  sub.New,
		deleteCustomer:   customer.Del,
		newUsageRecord:   usagerecord.New,
	}
}

// CreateSubscription implements the Subscriber.CreateSubscription
// method and adds the required customer, subscription, and
// payment information to Stripe.
func (c *Client) CreateSubscription(ctx context.Context, accountID string, info SubscriberInfo) (*Subscription, error) {
	fields := checkInfoFields(info)
	if len(fields) > 0 {
		return nil, &ErrorMissingFields{fields: fields}
	}

	paymentParams := &stripe.PaymentMethodParams{
		Card: &stripe.PaymentMethodCardParams{
			Number:   &info.CardNumber,
			ExpMonth: &info.ExpirationMonth,
			ExpYear:  &info.ExpirationYear,
			CVC:      &info.CardSecurityCode,
		},
		Type: stripe.String("card"),
		Params: stripe.Params{
			Metadata: map[string]string{
				"account_id": accountID,
			},
		},
	}
	newPaymentMethod, err := c.newPaymentMethod(paymentParams)
	if err != nil {
		return nil, &ErrorNewPaymentMethod{err: err}
	}

	customerParams := &stripe.CustomerParams{
		PaymentMethod: &newPaymentMethod.ID,
		Email:         &info.Email,
		Address: &stripe.AddressParams{
			PostalCode: &info.ZIP,
		},
		Params: stripe.Params{
			Metadata: map[string]string{
				"account_id": accountID,
			},
		},
	}
	newCustomer, err := c.newCustomer(customerParams)
	if err != nil {
		return nil, &ErrorNewCustomer{err: err}
	}

	subscriptionParams := &stripe.SubscriptionParams{
		Customer: &newCustomer.ID,
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(os.Getenv("STRIPE_FIXED_PRICE_ID")),
			},
			{
				Price: stripe.String(os.Getenv("STRIPE_VARIABLE_PRICE_ID")),
			},
		},
		Params: stripe.Params{
			Metadata: map[string]string{
				"account_id": accountID,
			},
		},
	}
	newSubscription, err := c.newSubscription(subscriptionParams)
	if err != nil {
		return nil, &ErrorNewSubscription{err: err}
	}

	subscription := &Subscription{
		ID:                    uuid.NewString(),
		StripePaymentMethodID: newPaymentMethod.ID,
		StripeCustomerID:      newCustomer.ID,
		StripeSubscriptionID:  newSubscription.ID,
	}

	return subscription, nil
}

func checkInfoFields(info SubscriberInfo) []string {
	fields := []string{}

	if info.ID == "" {
		fields = append(fields, "subscriber info id")
	}
	if info.Email == "" {
		fields = append(fields, "email")
	}
	if info.ZIP == "" {
		fields = append(fields, "zip")
	}
	if info.ExpirationMonth == "" {
		fields = append(fields, "expiration month")
	}
	if info.ExpirationYear == "" {
		fields = append(fields, "expiration year")
	}
	if info.CardNumber == "" {
		fields = append(fields, "card number")
	}
	if info.CardSecurityCode == "" {
		fields = append(fields, "card security code")
	}

	return fields
}

// RemoveSubscription implements the Subscriber.RemoveSubscription
// and removes the customer and cancels the subscription in Stripe.
func (c *Client) RemoveSubscription(ctx context.Context, subscription Subscription) error {
	_, err := c.deleteCustomer(subscription.StripeCustomerID, nil)
	if err != nil {
		return &ErrorDeleteCustomer{err: err}
	}

	return nil
}

// AddUsage implements the Subscriber.AddUsage method and adds a
// usage record for metered billing in Stripe for the provided
// subscription item id.
func (c *Client) AddUsage(ctx context.Context, id string) error {
	params := &stripe.UsageRecordParams{
		Quantity:         stripe.Int64(1),
		SubscriptionItem: stripe.String(id),
		Timestamp:        stripe.Int64(time.Now().Unix()),
	}
	_, err := c.newUsageRecord(params)
	if err != nil {
		return &ErrorCreateUsageRecord{err: err}
	}

	return nil
}
