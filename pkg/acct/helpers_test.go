package acct

import (
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func Test_itemToAccountObject(t *testing.T) {
	accountID := "test_account_id"
	subscriptionID := "test_subscription_id"
	stripePaymentMethodID := "test_stripe_payment_method_id"
	stripeCustomerID := "test_stripe_customer_id"
	stripeSubscriptionID := "test_stripe_subscription_id"
	stripeSubscriptionItemID := "test_stripe_subscription_item_id"

	input := map[string]*dynamodb.AttributeValue{
		accountIDKey: {
			S: aws.String(accountID),
		},
		subscriptionIDKey: {
			S: aws.String(subscriptionID),
		},
		stripePaymentMethodIDKey: {
			S: aws.String(stripePaymentMethodID),
		},
		stripeCustomerIDKey: {
			S: aws.String(stripeCustomerID),
		},
		stripeSubscriptionIDKey: {
			S: aws.String(stripeSubscriptionID),
		},
		stripeSubscriptionItemIDKey: {
			S: aws.String(stripeSubscriptionItemID),
		},
	}

	account := itemToAccountObject(input)

	if account.ID != accountID {
		t.Errorf("incorrect account id, received: %s, expected: %s", account.ID, accountID)
	}

	if account.SubscriptionID != subscriptionID {
		t.Errorf("incorrect subscription id, received: %s, expected: %s", account.SubscriptionID, subscriptionID)
	}

	if account.StripePaymentMethodID != stripePaymentMethodID {
		t.Errorf("incorrect stripe payment method id, received: %s, expected: %s", account.StripePaymentMethodID, stripePaymentMethodID)
	}

	if account.StripeCustomerID != stripeCustomerID {
		t.Errorf("incorrect stripe customer id, received: %s, expected: %s", account.StripeCustomerID, stripeCustomerID)
	}

	if account.StripeSubscriptionID != stripeSubscriptionID {
		t.Errorf("incorrect stripe subscription id, received: %s, expected: %s", account.StripeSubscriptionID, stripeSubscriptionID)
	}

	if account.StripeSubscriptionItemID != stripeSubscriptionItemID {
		t.Errorf("incorrect stripe subscription item id, received: %s, expected: %s", account.StripeSubscriptionItemID, stripeSubscriptionItemID)
	}
}

func Test_checkValues(t *testing.T) {
	tests := []struct {
		description string
		values      map[string]string
		key         string
		ok          bool
	}{
		{
			description: "values key not supported",
			values: map[string]string{
				"not_supported_key": "not_supported_value",
			},
			key: "not_supported_key",
			ok:  false,
		},
		{
			description: "all values keys supported",
			values: map[string]string{
				accountIDKey: "account_id",
			},
			key: "",
			ok:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			ok, key := checkValues(test.values)

			if key != test.key {
				t.Errorf("incorrect key, received: %s, expected: %s", key, test.key)
			}

			if ok != test.ok {
				t.Errorf("incorrect ok, received: %t, expected: %t", ok, test.ok)
			}
		})
	}
}

func Test_generateExpressionAndAttributes(t *testing.T) {
	values := map[string]string{
		subscriptionIDKey:       "subscription_id_value",
		stripeSubscriptionIDKey: "stripe_subscription_id_value",
	}

	expression, attributes := generateExpressionAndAttributes(values)

	if !strings.Contains(expression, "SET ") ||
		!strings.Contains(expression, "stripe_subscription_id_value = :stripe_subscription_id") ||
		!strings.Contains(expression, "subscription_id_value = :subscription_id") {
		t.Errorf("incorrect expression, received: %s", expression)
	}

	expectedAttributes := map[string]*dynamodb.AttributeValue{
		":" + subscriptionIDKey: {
			S: aws.String("subscription_id_value"),
		},
		":" + stripeSubscriptionIDKey: {
			S: aws.String(("stripe_subscription_id_value")),
		},
	}
	if !reflect.DeepEqual(attributes, expectedAttributes) {
		t.Errorf("incorrect attributes, received: %+v, expected: %+v", attributes, expectedAttributes)
	}
}