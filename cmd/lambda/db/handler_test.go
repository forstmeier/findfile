package main

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"

	"github.com/cheesesteakio/api/pkg/acct"
	"github.com/cheesesteakio/api/pkg/db"
	"github.com/cheesesteakio/api/pkg/docpars"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

type mockAcctClient struct {
	mockGetAccountBySecondaryIDOutput *acct.Account
	mockGetAccountBySecondaryIDError  error
}

func (m *mockAcctClient) CreateAccount(ctx context.Context, accountID, bucketName string) error {
	return nil
}

func (m *mockAcctClient) GetAccountByID(ctx context.Context, accountID string) (*acct.Account, error) {
	return nil, nil
}

func (m *mockAcctClient) GetAccountBySecondaryID(ctx context.Context, secondaryID string) (*acct.Account, error) {
	return m.mockGetAccountBySecondaryIDOutput, m.mockGetAccountBySecondaryIDError
}

func (m *mockAcctClient) UpdateAccount(ctx context.Context, accountID string, values map[string]string) error {
	return nil
}

func (m *mockAcctClient) DeleteAccount(ctx context.Context, accountID string) error {
	return nil
}

type mockDocParsClient struct {
	mockParseAccountID string
	mockParseFilename  string
	mockParseFilepath  string
	mockParseOutput    *docpars.Document
	mockParseError     error
}

func (m *mockDocParsClient) Parse(ctx context.Context, accountID, filename, filepath string, doc []byte) (*docpars.Document, error) {
	m.mockParseAccountID = accountID
	m.mockParseFilename = filename
	m.mockParseFilepath = filepath

	return m.mockParseOutput, m.mockParseError
}

type mockDBClient struct {
	mockCreateOrUpdateDocumentsError error
	mockDeleteDocumentsError         error
}

func (m *mockDBClient) CreateOrUpdateDocuments(ctx context.Context, documents []docpars.Document) error {
	return m.mockCreateOrUpdateDocumentsError
}

func (m *mockDBClient) DeleteDocuments(ctx context.Context, documentsInfo []db.DocumentInfo) error {
	return m.mockDeleteDocumentsError
}

func (m *mockDBClient) QueryDocuments(ctx context.Context, query []byte) ([]docpars.Document, error) {
	return nil, nil
}

func Test_handler(t *testing.T) {
	tests := []struct {
		description                       string
		s3Event                           events.S3Event
		mockGetAccountBySecondaryIDOutput *acct.Account
		mockGetAccountBySecondaryIDError  error
		mockParseOutput                   *docpars.Document
		mockParseError                    error
		mockCreateOrUpdateDocumentsError  error
		mockDeleteDocumentsError          error
		parseAccountID                    string
		parseFilename                     string
		parseFilepath                     string
		error                             error
	}{
		{
			description: "unsupported event type error",
			s3Event: events.S3Event{
				Records: []events.S3EventRecord{
					{
						EventName: "not_supported",
					},
				},
			},
			mockGetAccountBySecondaryIDOutput: nil,
			mockGetAccountBySecondaryIDError:  nil,
			mockParseOutput:                   nil,
			mockParseError:                    nil,
			mockCreateOrUpdateDocumentsError:  nil,
			mockDeleteDocumentsError:          nil,
			parseAccountID:                    "",
			parseFilename:                     "",
			parseFilepath:                     "",
			error:                             errorUnsupportedEvent,
		},
		{
			description: "error getting account by secondary id",
			s3Event: events.S3Event{
				Records: []events.S3EventRecord{
					{
						EventName: "ObjectCreated:Put",
						S3: events.S3Entity{
							Bucket: events.S3Bucket{
								Name: "test_bucket",
							},
						},
					},
				},
			},
			mockGetAccountBySecondaryIDOutput: nil,
			mockGetAccountBySecondaryIDError:  errors.New("mock get account by secondary id error"),
			mockParseOutput:                   nil,
			mockParseError:                    nil,
			mockCreateOrUpdateDocumentsError:  nil,
			mockDeleteDocumentsError:          nil,
			parseAccountID:                    "",
			parseFilename:                     "",
			parseFilepath:                     "",
			error:                             errorGetAccount,
		},
		{
			description: "error parsing request file",
			s3Event: events.S3Event{
				Records: []events.S3EventRecord{
					{
						EventName: "ObjectCreated:Put",
						S3: events.S3Entity{
							Bucket: events.S3Bucket{
								Name: "test_bucket",
							},
							Object: events.S3Object{
								Key: "test_file.jpg",
							},
						},
					},
				},
			},
			mockGetAccountBySecondaryIDOutput: &acct.Account{
				ID: "account_id",
			},
			mockGetAccountBySecondaryIDError: nil,
			mockParseOutput:                  nil,
			mockParseError:                   errors.New("mock parse error"),
			mockCreateOrUpdateDocumentsError: nil,
			mockDeleteDocumentsError:         nil,
			parseAccountID:                   "account_id",
			parseFilename:                    "test_file.jpg",
			parseFilepath:                    "test_bucket",
			error:                            errorParseFile,
		},
		{
			description: "create or update documents method error",
			s3Event: events.S3Event{
				Records: []events.S3EventRecord{
					{
						EventName: "ObjectCreated:Put",
						S3: events.S3Entity{
							Bucket: events.S3Bucket{
								Name: "test_bucket",
							},
							Object: events.S3Object{
								Key: "test_file.jpg",
							},
						},
					},
				},
			},
			mockGetAccountBySecondaryIDOutput: &acct.Account{
				ID: "account_id",
			},
			mockGetAccountBySecondaryIDError: nil,
			mockParseOutput:                  &docpars.Document{},
			mockParseError:                   nil,
			mockCreateOrUpdateDocumentsError: errors.New("create or update documents mock error"),
			mockDeleteDocumentsError:         nil,
			parseAccountID:                   "account_id",
			parseFilename:                    "test_file.jpg",
			parseFilepath:                    "test_bucket",
			error:                            errorCreateOrUpdateDocuments,
		},
		{
			description: "delete documents method error",
			s3Event: events.S3Event{
				Records: []events.S3EventRecord{
					{
						EventName: "ObjectRemoved:Delete",
						S3: events.S3Entity{
							Bucket: events.S3Bucket{
								Name: "test_bucket",
							},
							Object: events.S3Object{
								Key: "test_file.jpg",
							},
						},
					},
				},
			},
			mockGetAccountBySecondaryIDOutput: &acct.Account{
				ID: "account_id",
			},
			mockGetAccountBySecondaryIDError: nil,
			mockParseOutput:                  &docpars.Document{},
			mockParseError:                   nil,
			mockCreateOrUpdateDocumentsError: nil,
			mockDeleteDocumentsError:         errors.New("delete documents mock error"),
			parseAccountID:                   "",
			parseFilename:                    "",
			parseFilepath:                    "",
			error:                            errorDeleteDocuments,
		},
		{
			description: "successful database handler invocation",
			s3Event: events.S3Event{
				Records: []events.S3EventRecord{
					{
						EventName: "ObjectCreated:Put",
						S3: events.S3Entity{
							Bucket: events.S3Bucket{
								Name: "test_bucket",
							},
							Object: events.S3Object{
								Key: "test_file_1.jpg",
							},
						},
					},
					{
						EventName: "ObjectRemoved:Delete",
						S3: events.S3Entity{
							Bucket: events.S3Bucket{
								Name: "test_bucket",
							},
							Object: events.S3Object{
								Key: "test_file_2.jpg",
							},
						},
					},
				},
			},
			mockGetAccountBySecondaryIDOutput: &acct.Account{
				ID: "account_id",
			},
			mockGetAccountBySecondaryIDError: nil,
			mockParseOutput:                  &docpars.Document{},
			mockParseError:                   nil,
			mockCreateOrUpdateDocumentsError: nil,
			mockDeleteDocumentsError:         nil,
			parseAccountID:                   "account_id",
			parseFilename:                    "test_file_1.jpg",
			parseFilepath:                    "test_bucket",
			error:                            nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			acctClient := &mockAcctClient{
				mockGetAccountBySecondaryIDOutput: test.mockGetAccountBySecondaryIDOutput,
				mockGetAccountBySecondaryIDError:  test.mockGetAccountBySecondaryIDError,
			}

			docparseClient := &mockDocParsClient{
				mockParseOutput: test.mockParseOutput,
				mockParseError:  test.mockParseError,
			}

			dbClient := &mockDBClient{
				mockCreateOrUpdateDocumentsError: test.mockCreateOrUpdateDocumentsError,
				mockDeleteDocumentsError:         test.mockDeleteDocumentsError,
			}

			handlerFunc := handler(acctClient, docparseClient, dbClient)

			s3Event, err := json.Marshal(test.s3Event)
			if err != nil {
				t.Fatalf("error marshalling s3 event body: %s", err.Error())
			}

			err = handlerFunc(context.Background(), events.SNSEvent{
				Records: []events.SNSEventRecord{
					{
						SNS: events.SNSEntity{
							Message: string(s3Event),
						},
					},
				},
			})

			if err != test.error {
				t.Errorf("incorrect error, received: %v, expected: %v", err, test.error)
			}

			if docparseClient.mockParseAccountID != test.parseAccountID {
				t.Errorf("incorrect parse account id, received: %s, expected: %s", docparseClient.mockParseAccountID, test.parseAccountID)
			}

			if docparseClient.mockParseFilename != test.parseFilename {
				t.Errorf("incorrect parse filename, received: %s, expected: %s", docparseClient.mockParseFilename, test.parseFilename)
			}

			if docparseClient.mockParseFilepath != test.parseFilepath {
				t.Errorf("incorrect parse filepath, received: %s, expected: %s", docparseClient.mockParseFilepath, test.parseFilepath)
			}
		})
	}
}
