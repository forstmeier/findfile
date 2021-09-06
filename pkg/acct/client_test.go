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
	client := New(session.New(), "table_name")

	if client == nil {
		t.Error("error creating accounter client")
	}
}

type mockDynamoDBClient struct {
	mockPutItemError    error
	mockGetItemOutput   *dynamodb.GetItemOutput
	mockGetItemError    error
	mockUpdateItemError error
	mockDeleteItemError error
}

func (m *mockDynamoDBClient) PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	return nil, m.mockPutItemError
}

func (m *mockDynamoDBClient) GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	return m.mockGetItemOutput, m.mockGetItemError
}

func (m *mockDynamoDBClient) UpdateItem(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
	return nil, m.mockUpdateItemError
}

func (m *mockDynamoDBClient) DeleteItem(input *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {
	return nil, m.mockDeleteItemError
}

func TestCreateAccount(t *testing.T) {
	tests := []struct {
		description      string
		mockPutItemError error
		error            error
	}{
		{
			description:      "dynamodb client put item error",
			mockPutItemError: errors.New("mock put item error"),
			error:            &ErrorPutItem{},
		},
		{
			description:      "successful create account invocation",
			mockPutItemError: nil,
			error:            nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				dynamoDBClient: &mockDynamoDBClient{
					mockPutItemError: test.mockPutItemError,
				},
			}

			err := client.CreateAccount(context.Background(), "account_id", "bucket_name")

			if err != nil {
				switch e := test.error.(type) {
				case *ErrorPutItem:
					if !errors.As(err, &e) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, e)
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

func TestGetAccountByID(t *testing.T) {
	tests := []struct {
		description       string
		mockGetItemOutput *dynamodb.GetItemOutput
		mockGetItemError  error
		account           *Account
		error             error
	}{
		{
			description:       "dynamodb client get item error",
			mockGetItemOutput: nil,
			mockGetItemError:  errors.New("mock get item error"),
			account:           nil,
			error:             &ErrorGetItem{},
		},
		{
			description: "successful read account invocation",
			mockGetItemOutput: &dynamodb.GetItemOutput{
				Item: map[string]*dynamodb.AttributeValue{
					AccountIDKey: {
						S: aws.String("account_id"),
					},
				},
			},
			mockGetItemError: nil,
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
					mockGetItemOutput: test.mockGetItemOutput,
					mockGetItemError:  test.mockGetItemError,
				},
			}

			account, err := client.GetAccountByID(context.Background(), "account_id")

			if err != nil {
				switch e := test.error.(type) {
				case *ErrorGetItem:
					if !errors.As(err, &e) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, e)
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

func TestGetAccountBySecondaryID(t *testing.T) {
	tests := []struct {
		description       string
		mockGetItemOutput *dynamodb.GetItemOutput
		mockGetItemError  error
		account           *Account
		error             error
	}{
		{
			description:       "dynamodb client get item error",
			mockGetItemOutput: nil,
			mockGetItemError:  errors.New("mock get item error"),
			account:           nil,
			error:             &ErrorGetItem{},
		},
		{
			description: "successful read account invocation",
			mockGetItemOutput: &dynamodb.GetItemOutput{
				Item: map[string]*dynamodb.AttributeValue{
					AccountIDKey: {
						S: aws.String("account_id"),
					},
				},
			},
			mockGetItemError: nil,
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
					mockGetItemOutput: test.mockGetItemOutput,
					mockGetItemError:  test.mockGetItemError,
				},
			}

			account, err := client.GetAccountBySecondaryID(context.Background(), "bucket_name")

			if err != nil {
				switch e := test.error.(type) {
				case *ErrorGetItem:
					if !errors.As(err, &e) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, e)
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
		description         string
		values              map[string]string
		mockUpdateItemError error
		error               error
	}{
		{
			description: "incorrect update value key received",
			values: map[string]string{
				"not_supported_key": "value",
			},
			mockUpdateItemError: nil,
			error:               &ErrorIncorrectValue{},
		},
		{
			description: "dynamodb client update item error",
			values: map[string]string{
				AccountIDKey: "account_id",
			},
			mockUpdateItemError: errors.New("mock update item error"),
			error:               &ErrorUpdateItem{},
		},
		{
			description: "successful update account invocation",
			values: map[string]string{
				AccountIDKey: "account_id",
			},
			mockUpdateItemError: nil,
			error:               nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				dynamoDBClient: &mockDynamoDBClient{
					mockUpdateItemError: test.mockUpdateItemError,
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
		description         string
		mockDeleteItemError error
		error               error
	}{
		{
			description:         "dynamodb client delete item error",
			mockDeleteItemError: errors.New("mock delete item error"),
			error:               &ErrorDeleteItem{},
		},
		{
			description:         "successful delete account invocation",
			mockDeleteItemError: nil,
			error:               nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				dynamoDBClient: &mockDynamoDBClient{
					mockDeleteItemError: test.mockDeleteItemError,
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
