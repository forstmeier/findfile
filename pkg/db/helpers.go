package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/cheesesteakio/api/pkg/docpars"
)

var _ helper = &help{}

type helper interface {
	uploadObject(ctx context.Context, body interface{}, key string) error
	listDocumentKeys(ctx context.Context, bucket, prefix string) ([]string, error)
	deleteDocumentsByKeys(ctx context.Context, keys []string) error
	executeQuery(ctx context.Context, query []byte) (*string, *string, error)
	getQueryResultAccountID(state, executionID string) (*string, error)
	getQueryResultDocuments(state, executionID string) ([]docpars.Document, error)
}

type help struct {
	databaseName string
	bucketName   string
	athenaClient athenaClient
	s3Client     s3Client
}

func (h *help) uploadObject(ctx context.Context, body interface{}, key string) error {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	_, err = h.s3Client.PutObject(&s3.PutObjectInput{
		Body:   aws.ReadSeekCloser(bytes.NewReader(bodyBytes)),
		Bucket: aws.String(h.bucketName),
		Key:    aws.String(key),
	})

	return err
}

func (h *help) listDocumentKeys(ctx context.Context, bucket, prefix string) ([]string, error) {
	var results []string
	var continuationToken string
	for {
		output, err := h.s3Client.ListObjectsV2(&s3.ListObjectsV2Input{
			Bucket:            aws.String(bucket),
			Prefix:            aws.String(prefix),
			ContinuationToken: aws.String(continuationToken),
		})
		if err != nil {
			return nil, err
		}

		for _, content := range output.Contents {
			results = append(results, *content.Key)
		}

		if *output.IsTruncated {
			continuationToken = *output.NextContinuationToken
		} else {
			break
		}
	}

	return results, nil
}

// deleteDocumentsByKeys processes pre-chunked slices of keys according
// to the S3 1000 object limit per invocation.
func (h *help) deleteDocumentsByKeys(ctx context.Context, keys []string) error {
	objects := make([]*s3.ObjectIdentifier, len(keys))
	for i, key := range keys {
		objects[i] = &s3.ObjectIdentifier{
			Key: aws.String(key),
		}
	}

	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(h.bucketName),
		Delete: &s3.Delete{
			Objects: objects,
		},
	}

	_, err := h.s3Client.DeleteObjects(input)
	return err
}

func (h *help) executeQuery(ctx context.Context, query []byte) (*string, *string, error) {
	queryInput := athena.StartQueryExecutionInput{
		QueryString: aws.String(string(query)),
		QueryExecutionContext: &athena.QueryExecutionContext{
			Database: aws.String(h.databaseName),
		},
		ResultConfiguration: &athena.ResultConfiguration{
			OutputLocation: aws.String(fmt.Sprintf("s3://%s/results", h.bucketName)),
		},
	}

	queryOutput, err := h.athenaClient.StartQueryExecution(&queryInput)
	if err != nil {
		return nil, nil, err
	}

	executionInput := &athena.GetQueryExecutionInput{
		QueryExecutionId: queryOutput.QueryExecutionId,
	}
	executionOutput := &athena.GetQueryExecutionOutput{}
	for {
		executionOutput, err = h.athenaClient.GetQueryExecution(executionInput)
		if err != nil {
			return nil, nil, err
		}
		if *executionOutput.QueryExecution.Status.State != "RUNNING" {
			break
		}
		time.Sleep(1 * time.Second)
	}

	return queryOutput.QueryExecutionId, executionOutput.QueryExecution.Status.State, nil
}

func (h *help) getQueryResultAccountID(state, executionID string) (*string, error) {
	if state != "SUCCEEDED" {
		return nil, fmt.Errorf("incorrect query state [%s]", state)
	}

	results, err := h.athenaClient.GetQueryResults(&athena.GetQueryResultsInput{
		QueryExecutionId: &executionID,
	})
	if err != nil {
		return nil, err
	}

	row := results.ResultSet.Rows[0]
	accountID := *row.Data[0].VarCharValue

	return &accountID, nil
}

func (h *help) getQueryResultDocuments(state, executionID string) ([]docpars.Document, error) {
	if state != "SUCCEEDED" {
		return nil, fmt.Errorf("incorrect query state [%s]", state)
	}

	results, err := h.athenaClient.GetQueryResults(&athena.GetQueryResultsInput{
		QueryExecutionId: &executionID,
	})
	if err != nil {
		return nil, err
	}

	documents := []docpars.Document{}
	for _, row := range results.ResultSet.Rows {
		document := docpars.Document{
			AccountID: *row.Data[0].VarCharValue,
			Filepath:  *row.Data[1].VarCharValue,
			Filename:  *row.Data[2].VarCharValue,
		}

		documents = append(documents, document)
	}

	return documents, nil
}
