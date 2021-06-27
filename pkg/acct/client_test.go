package acct

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func TestNew(t *testing.T) {
	client := New(session.New())

	if client == nil {
		t.Error("error creating accounter client")
	}
}

type mockDynamoDBClient struct {
	putItemError    error
	getItemOutput   *dynamodb.GetItemOutput
	getItemError    error
	updateItemError error
	deleteItemError error
}

func (m *mockDynamoDBClient) PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	return nil, m.putItemError
}

func (m *mockDynamoDBClient) GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	return m.getItemOutput, m.getItemError
}

func (m *mockDynamoDBClient) UpdateItem(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
	return nil, m.updateItemError
}

func (m *mockDynamoDBClient) DeleteItem(input *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {
	return nil, m.deleteItemError
}

func TestCreateAccount(t *testing.T) {
	tests := []struct {
		description  string
		putItemError error
		error        error
	}{
		{
			description:  "dynamodb client put item error",
			putItemError: errors.New("mock put item error"),
			error:        &ErrorPutItem{},
		},
		{
			description:  "successful create account invocation",
			putItemError: nil,
			error:        nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				dynamoDBClient: &mockDynamoDBClient{
					putItemError: test.putItemError,
				},
			}

			err := client.CreateAccount(context.Background(), "account_id")

			if err != nil {
				switch test.error.(type) {
				case *ErrorPutItem:
					var testError *ErrorPutItem
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

func TestReadAccount(t *testing.T) {
	tests := []struct {
		description   string
		getItemOutput *dynamodb.GetItemOutput
		getItemError  error
		account       *Account
		error         error
	}{
		{
			description:   "dynamodb client get item error",
			getItemOutput: nil,
			getItemError:  errors.New("mock get item error"),
			account:       nil,
			error:         &ErrorGetItem{},
		},
		{
			description: "successful read account invocation",
			getItemOutput: &dynamodb.GetItemOutput{
				Item: map[string]*dynamodb.AttributeValue{
					AccountIDKey: {
						S: aws.String("account_id"),
					},
				},
			},
			getItemError: nil,
			account: &Account{
				ID: "account_id",
			},
			error: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				dynamoDBClient: &mockDynamoDBClient{
					getItemOutput: test.getItemOutput,
					getItemError:  test.getItemError,
				},
			}

			account, err := client.ReadAccount(context.Background(), "account_id")

			if err != nil {
				switch test.error.(type) {
				case *ErrorGetItem:
					var testError *ErrorGetItem
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

			if !reflect.DeepEqual(account, test.account) {
				t.Errorf("incorrect account, received: %+v, expected: %+v", account, test.account)
			}
		})
	}
}

func TestUpdateAccount(t *testing.T) {
	tests := []struct {
		description     string
		values          map[string]string
		updateItemError error
		error           error
	}{
		{
			description: "incorrect update value key received",
			values: map[string]string{
				"not_supported_key": "value",
			},
			updateItemError: nil,
			error:           &ErrorIncorrectValue{},
		},
		{
			description: "dynamodb client update item error",
			values: map[string]string{
				AccountIDKey: "account_id",
			},
			updateItemError: errors.New("mock update item error"),
			error:           &ErrorUpdateItem{},
		},
		{
			description: "successful update account invocation",
			values: map[string]string{
				AccountIDKey: "account_id",
			},
			updateItemError: nil,
			error:           nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				dynamoDBClient: &mockDynamoDBClient{
					updateItemError: test.updateItemError,
				},
			}

			err := client.UpdateAccount(context.Background(), "account_id", test.values)

			if err != nil {
				switch test.error.(type) {
				case *ErrorIncorrectValue:
					var testError *ErrorIncorrectValue
					if !errors.As(err, &testError) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, testError)
					}
				case *ErrorUpdateItem:
					var testError *ErrorUpdateItem
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

func TestDeleteAccount(t *testing.T) {
	tests := []struct {
		description     string
		deleteItemError error
		error           error
	}{
		{
			description:     "dynamodb client delete item error",
			deleteItemError: errors.New("mock delete item error"),
			error:           &ErrorDeleteItem{},
		},
		{
			description:     "successful delete account invocation",
			deleteItemError: nil,
			error:           nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				dynamoDBClient: &mockDynamoDBClient{
					deleteItemError: test.deleteItemError,
				},
			}

			err := client.DeleteAccount(context.Background(), "account_id")

			if err != nil {
				switch test.error.(type) {
				case *ErrorDeleteItem:
					var testError *ErrorDeleteItem
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
