package subscr

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stripe/stripe-go/v72"
)

func TestNew(t *testing.T) {
	client := New("stripe_api_key")
	if client == nil {
		t.Error("error creating subscr client")
	}
}
func TestCreateSubscription(t *testing.T) {
	tests := []struct {
		description            string
		subscriberInfo         SubscriberInfo
		newPaymentMethodOutput *stripe.PaymentMethod
		newPaymentMethodError  error
		newCustomerOutput      *stripe.Customer
		newCustomerError       error
		newSubscriptionOutput  *stripe.Subscription
		newSubscriptionError   error
		subscription           *Subscription
		error                  error
	}{
		{
			description: "missing subscriber info fields",
			subscriberInfo: SubscriberInfo{
				ID:    "subscriber_info_id",
				Email: "subscriber@email.com",
			},
			newPaymentMethodOutput: nil,
			newPaymentMethodError:  nil,
			newCustomerOutput:      nil,
			newCustomerError:       nil,
			newSubscriptionOutput:  nil,
			newSubscriptionError:   nil,
			subscription:           nil,
			error:                  &ErrorMissingFields{},
		},
		{
			description: "error creating payment method",
			subscriberInfo: SubscriberInfo{
				ID:               "subscriber_info_id",
				Email:            "subscriber@email.com",
				ZIP:              "12345",
				ExpirationMonth:  "05",
				ExpirationYear:   "1977",
				CardNumber:       "123443211234",
				CardSecurityCode: "123",
			},
			newPaymentMethodOutput: nil,
			newPaymentMethodError:  errors.New("mock new payment method error"),
			newCustomerOutput:      nil,
			newCustomerError:       nil,
			newSubscriptionOutput:  nil,
			newSubscriptionError:   nil,
			subscription:           nil,
			error:                  &ErrorNewPaymentMethod{},
		},
		{
			description: "error creating customer",
			subscriberInfo: SubscriberInfo{
				ID:               "subscriber_info_id",
				Email:            "subscriber@email.com",
				ZIP:              "12345",
				ExpirationMonth:  "05",
				ExpirationYear:   "1977",
				CardNumber:       "123443211234",
				CardSecurityCode: "123",
			},
			newPaymentMethodOutput: &stripe.PaymentMethod{
				ID: "test_payment_method_id",
			},
			newPaymentMethodError: nil,
			newCustomerOutput:     nil,
			newCustomerError:      errors.New("mock new customer error"),
			newSubscriptionOutput: nil,
			newSubscriptionError:  nil,
			subscription:          nil,
			error:                 &ErrorNewCustomer{},
		},
		{
			description: "error creating subscription",
			subscriberInfo: SubscriberInfo{
				ID:               "subscriber_info_id",
				Email:            "subscriber@email.com",
				ZIP:              "12345",
				ExpirationMonth:  "05",
				ExpirationYear:   "1977",
				CardNumber:       "123443211234",
				CardSecurityCode: "123",
			},
			newPaymentMethodOutput: &stripe.PaymentMethod{
				ID: "test_payment_method_id",
			},
			newPaymentMethodError: nil,
			newCustomerOutput: &stripe.Customer{
				ID: "test_customer_id",
			},
			newCustomerError:      nil,
			newSubscriptionOutput: nil,
			newSubscriptionError:  errors.New("mock new subscription error"),
			subscription:          nil,
			error:                 &ErrorNewSubscription{},
		},
		{
			description: "successful create subscription",
			subscriberInfo: SubscriberInfo{
				ID:               "subscriber_info_id",
				Email:            "subscriber@email.com",
				ZIP:              "12345",
				ExpirationMonth:  "05",
				ExpirationYear:   "1977",
				CardNumber:       "123443211234",
				CardSecurityCode: "123",
			},
			newPaymentMethodOutput: &stripe.PaymentMethod{
				ID: "test_payment_method_id",
			},
			newPaymentMethodError: nil,
			newCustomerOutput: &stripe.Customer{
				ID: "test_customer_id",
			},
			newCustomerError: nil,
			newSubscriptionOutput: &stripe.Subscription{
				ID: "test_subscription_id",
				Items: &stripe.SubscriptionItemList{
					Data: []*stripe.SubscriptionItem{
						{
							ID: "test_subscription_item_id",
							Price: &stripe.Price{
								ID: os.Getenv("STRIPE_VARIABLE_PRICE_ID"),
							},
						},
					},
				},
			},
			newSubscriptionError: nil,
			subscription: &Subscription{
				StripePaymentMethodID: "test_payment_method_id",
				StripeCustomerID:      "test_customer_id",
				StripeSubscriptionID:  "test_subscription_id",
			},
			error: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				newPaymentMethod: func(params *stripe.PaymentMethodParams) (*stripe.PaymentMethod, error) {
					return test.newPaymentMethodOutput, test.newPaymentMethodError
				},
				newCustomer: func(params *stripe.CustomerParams) (*stripe.Customer, error) {
					return test.newCustomerOutput, test.newCustomerError
				},
				newSubscription: func(params *stripe.SubscriptionParams) (*stripe.Subscription, error) {
					return test.newSubscriptionOutput, test.newSubscriptionError
				},
			}

			subscription, err := client.CreateSubscription(context.Background(), "account_id", test.subscriberInfo)

			if err != nil {
				switch test.error.(type) {
				case *ErrorMissingFields:
					var testError *ErrorMissingFields
					if !errors.As(err, &testError) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, testError)
					}
				case *ErrorNewPaymentMethod:
					var testError *ErrorNewPaymentMethod
					if !errors.As(err, &testError) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, testError)
					}
				case *ErrorNewCustomer:
					var testError *ErrorNewCustomer
					if !errors.As(err, &testError) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, testError)
					}
				case *ErrorNewSubscription:
					var testError *ErrorNewSubscription
					if !errors.As(err, &testError) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, testError)
					}
				default:
					t.Fatalf("unexpected error type: %v", err)
				}
			} else {
				if err != test.error {
					t.Errorf("incorrect nil error, received: %v, expected: %v", err, test.error)
				}
			}

			if subscription != nil {
				checkSubscriptions(t, subscription, test.subscription)
			} else {
				if subscription != test.subscription {
					t.Errorf("incorrect nil check, received: %+v, expected: %+v", subscription, test.subscription)
				}
			}
		})
	}
}

func checkSubscriptions(t *testing.T, received, expected *Subscription) {
	check := false
	if received.StripePaymentMethodID != expected.StripePaymentMethodID {
		check = true
	} else if received.StripeCustomerID != expected.StripeCustomerID {
		check = true
	} else if received.StripeSubscriptionID != expected.StripeSubscriptionID {
		check = true
	}

	if check {
		t.Errorf("incorrect subscription, received: %+v, expected: %+v", received, expected)
	}
}

func Test_checkInfoFields(t *testing.T) {
	tests := []struct {
		description string
		info        SubscriberInfo
		fields      []string
	}{
		{
			description: "missing all fields",
			info:        SubscriberInfo{},
			fields: []string{
				"subscriber info id",
				"email",
				"zip",
				"expiration month",
				"expiration year",
				"card number",
				"card security code",
			},
		},
		{
			description: "missing no fields",
			info: SubscriberInfo{
				ID:               "test_subscriber_info_id",
				Email:            "test_email",
				ZIP:              "test_zip",
				ExpirationMonth:  "test_expiration_month",
				ExpirationYear:   "test_expiration_year",
				CardNumber:       "test_card_number",
				CardSecurityCode: "test_card_security_code",
			},
			fields: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			fields := checkInfoFields(test.info)

			for i, field := range fields {
				if field != test.fields[i] {
					t.Errorf("incorrect fields, received: %v, expected: %v", fields, test.fields)
					break // keep to ensure all test scenarios run
				}
			}
		})
	}
}

func TestRemoveSubscription(t *testing.T) {
	tests := []struct {
		description         string
		deleteCustomerError error
		error               error
	}{
		{
			description:         "error deleting customer",
			deleteCustomerError: errors.New("mock delete customer error"),
			error:               &ErrorDeleteCustomer{},
		},
		{
			description:         "successful remove subscription",
			deleteCustomerError: nil,
			error:               nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				deleteCustomer: func(id string, params *stripe.CustomerParams) (*stripe.Customer, error) {
					return nil, test.deleteCustomerError
				},
			}

			subscription := Subscription{
				StripeCustomerID: "test_customer_id",
			}

			err := client.RemoveSubscription(context.Background(), subscription)

			if err != nil {
				switch test.error.(type) {
				case *ErrorDeleteCustomer:
					var testError *ErrorDeleteCustomer
					if !errors.As(err, &testError) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, testError)
					}
				default:
					t.Fatalf("unexpected error type: %v", err)
				}
			} else {
				if err != test.error {
					t.Errorf("incorrect nil error, received: %v, expected: %v", err, test.error)
				}
			}
		})
	}
}

func TestAddUsage(t *testing.T) {
	tests := []struct {
		description            string
		createUsageRecordError error
		error                  error
	}{
		{
			description:            "error creating usage record",
			createUsageRecordError: errors.New("mock create usage record error"),
			error:                  &ErrorCreateUsageRecord{},
		},
		{
			description:            "successful add usage call",
			createUsageRecordError: nil,
			error:                  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				newUsageRecord: func(params *stripe.UsageRecordParams) (*stripe.UsageRecord, error) {
					return nil, test.createUsageRecordError
				},
			}

			err := client.AddUsage(context.Background(), "test_subscription_item_id")

			if err != nil {
				switch test.error.(type) {
				case *ErrorCreateUsageRecord:
					var testError *ErrorCreateUsageRecord
					if !errors.As(err, &testError) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, testError)
					}
				default:
					t.Fatalf("unexpected error type: %v", err)
				}
			} else {
				if err != test.error {
					t.Errorf("incorrect nil error, received: %v, expected: %v", err, test.error)
				}
			}
		})
	}
}
