package acct

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func itemToAccountObject(items map[string]*dynamodb.AttributeValue) *Account {
	account := &Account{}

	if attribute, ok := items[accountIDKey]; ok {
		account.ID = *attribute.S
	}

	if attribute, ok := items[subscriptionIDKey]; ok {
		account.SubscriptionID = *attribute.S
	}

	if attribute, ok := items[stripePaymentMethodIDKey]; ok {
		account.StripePaymentMethodID = *attribute.S
	}

	if attribute, ok := items[stripeCustomerIDKey]; ok {
		account.StripeCustomerID = *attribute.S
	}

	if attribute, ok := items[stripeSubscriptionIDKey]; ok {
		account.StripeSubscriptionID = *attribute.S
	}

	if attribute, ok := items[stripeSubscriptionItemIDKey]; ok {
		account.StripeSubscriptionItemID = *attribute.S
	}

	return account
}

func checkValues(values map[string]string) (bool, string) {
	supportedKeys := map[string]struct{}{
		accountIDKey:                {},
		subscriptionIDKey:           {},
		stripePaymentMethodIDKey:    {},
		stripeCustomerIDKey:         {},
		stripeSubscriptionIDKey:     {},
		stripeSubscriptionItemIDKey: {},
	}

	for key := range values {
		if _, ok := supportedKeys[key]; !ok {
			return ok, key
		}
	}

	return true, ""
}

func generateExpressionAndAttributes(values map[string]string) (string, map[string]*dynamodb.AttributeValue) {
	expressionElements := make([]string, len(values))
	attributes := make(map[string]*dynamodb.AttributeValue, len(values))

	i := 0
	for key, value := range values {
		expressionElements[i] = fmt.Sprintf("%s = :%s", value, key)

		attributes[":"+key] = &dynamodb.AttributeValue{
			S: aws.String(value),
		}

		i++
	}

	expression := "SET " + strings.Join(expressionElements, ", ")

	return expression, attributes
}
