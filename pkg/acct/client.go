package acct

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// DynamoDB-specific attribute keys for subscription values.
const (
	TableName                   = "accounts"
	AccountIDKey                = "id"
	SubscriptionIDKey           = "subscription_id"
	StripePaymentMethodIDKey    = "stripe_payment_method_id"
	StripeCustomerIDKey         = "stripe_customer_id"
	StripeSubscriptionIDKey     = "stripe_subscription_id"
	StripeSubscriptionItemIDKey = "stripe_subscription_item_id"
)

var _ Accounter = &Client{}

// Client implements the acct.Accounter methods using DynamoDB.
type Client struct {
	dynamoDBClient dynamoDBClient
}

type dynamoDBClient interface {
	PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error)
	GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error)
	UpdateItem(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error)
	DeleteItem(input *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error)
}

// New generates a acct.Client pointer instance with a DynamoDB client.
func New(newSession *session.Session) *Client {
	dynamoDBClient := dynamodb.New(newSession)

	return &Client{
		dynamoDBClient: dynamoDBClient,
	}
}

// CreateAccount implements the acct.Accounter.CreateAccount method.
func (c *Client) CreateAccount(ctx context.Context, accountID string) error {
	input := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			AccountIDKey: {
				S: aws.String(accountID),
			},
		},
		TableName: aws.String(TableName),
	}

	_, err := c.dynamoDBClient.PutItem(input)
	if err != nil {
		return &ErrorPutItem{err: err}
	}

	return nil
}

// ReadAccount implements the acct.Accounter.ReadAccount method.
func (c *Client) ReadAccount(ctx context.Context, accountID string) (*Account, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			AccountIDKey: {
				S: aws.String(accountID),
			},
		},
		TableName: aws.String(TableName),
	}

	output, err := c.dynamoDBClient.GetItem(input)
	if err != nil {
		return nil, &ErrorGetItem{err: err}
	}

	return itemToAccountObject(output.Item), nil
}

// UpdateAccount implements the acct.Accounter.UpdateAccount method.
func (c *Client) UpdateAccount(ctx context.Context, accountID string, values map[string]string) error {
	ok, key := checkValues(values)
	if !ok {
		return &ErrorIncorrectValue{key: key}
	}

	expression, attributes := generateExpressionAndAttributes(values)

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(TableName),
		Key: map[string]*dynamodb.AttributeValue{
			AccountIDKey: {
				S: aws.String(accountID),
			},
		},
		ExpressionAttributeValues: attributes,
		UpdateExpression:          aws.String(expression),
		ConditionExpression:       aws.String("attribute_exists(id)"),
	}

	_, err := c.dynamoDBClient.UpdateItem(input)
	if err != nil {
		return &ErrorUpdateItem{err: err}
	}

	return nil
}

// DeleteAccount implements the acct.Accounter.DeleteAccount method.
func (c *Client) DeleteAccount(ctx context.Context, accountID string) error {
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			AccountIDKey: {
				S: aws.String(accountID),
			},
		},
		TableName: aws.String(TableName),
	}

	_, err := c.dynamoDBClient.DeleteItem(input)
	if err != nil {
		return &ErrorDeleteItem{err: err}
	}

	return nil
}
