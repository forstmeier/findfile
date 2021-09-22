package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/findfiledev/api/pkg/pars"
)

var _ helper = &help{}

type helper interface {
	uploadObject(ctx context.Context, body interface{}, key string) error
	deleteDocumentsByKeys(ctx context.Context, keys []string) error
	executeQuery(ctx context.Context, query []byte) (*string, *string, error)
	getQueryResultDocuments(ctx context.Context, state, executionID string) ([]pars.Document, error)
	getQueryResultKeys(ctx context.Context, state, executionID string) ([]string, error)
	addFolder(ctx context.Context, folder string) error
	startCrawler(ctx context.Context) error
}

type athenaClient interface {
	StartQueryExecution(input *athena.StartQueryExecutionInput) (*athena.StartQueryExecutionOutput, error)
	GetQueryExecution(input *athena.GetQueryExecutionInput) (*athena.GetQueryExecutionOutput, error)
	GetQueryResults(input *athena.GetQueryResultsInput) (*athena.GetQueryResultsOutput, error)
}

type s3Client interface {
	PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error)
	DeleteObjects(input *s3.DeleteObjectsInput) (*s3.DeleteObjectsOutput, error)
}

type glueClient interface {
	StartCrawler(input *glue.StartCrawlerInput) (*glue.StartCrawlerOutput, error)
}

type help struct {
	databaseName   string
	databaseBucket string
	crawlerName    string
	athenaClient   athenaClient
	s3Client       s3Client
	glueClient     glueClient
}

func (h *help) uploadObject(ctx context.Context, body interface{}, key string) error {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	_, err = h.s3Client.PutObject(&s3.PutObjectInput{
		Body:   aws.ReadSeekCloser(bytes.NewReader(bodyBytes)),
		Bucket: aws.String(h.databaseBucket),
		Key:    aws.String(key),
	})

	return err
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
		Bucket: aws.String(h.databaseBucket),
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
			OutputLocation: aws.String(fmt.Sprintf("s3://%s/results", h.databaseBucket)),
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

func (h *help) getQueryResultDocuments(ctx context.Context, state, executionID string) ([]pars.Document, error) {
	if state != "SUCCEEDED" {
		return nil, fmt.Errorf("incorrect query state [%s]", state)
	}

	results, err := h.athenaClient.GetQueryResults(&athena.GetQueryResultsInput{
		QueryExecutionId: &executionID,
	})
	if err != nil {
		return nil, err
	}

	documents := []pars.Document{}
	for _, row := range results.ResultSet.Rows {
		document := pars.Document{
			ID:         *row.Data[0].VarCharValue,
			FileKey:    *row.Data[1].VarCharValue,
			FileBucket: *row.Data[2].VarCharValue,
		}

		documents = append(documents, document)
	}

	return documents, nil
}

func (h *help) getQueryResultKeys(ctx context.Context, state, executionID string) ([]string, error) {
	if state != "SUCCEEDED" {
		return nil, fmt.Errorf("incorrect query state [%s]", state)
	}

	results, err := h.athenaClient.GetQueryResults(&athena.GetQueryResultsInput{
		QueryExecutionId: &executionID,
	})
	if err != nil {
		return nil, err
	}

	documents := map[string]struct{}{}
	pages := map[string]struct{}{}
	lines := map[string]struct{}{}
	coordinates := map[string]struct{}{}

	keys := []string{}
	for _, row := range results.ResultSet.Rows {
		documentKey := fmt.Sprintf("%s/%s.json", Paths[0], *row.Data[0].VarCharValue)
		if _, ok := documents[documentKey]; !ok {
			keys = append(keys, documentKey)
		}

		pageKey := fmt.Sprintf("%s/%s.json", Paths[1], *row.Data[1].VarCharValue)
		if _, ok := pages[pageKey]; !ok {
			keys = append(keys, pageKey)
		}

		lineKey := fmt.Sprintf("%s/%s.json", Paths[2], *row.Data[2].VarCharValue)
		if _, ok := lines[lineKey]; !ok {
			keys = append(keys, lineKey)
		}

		coordinatesKey := fmt.Sprintf("%s/%s.json", Paths[3], *row.Data[3].VarCharValue)
		if _, ok := coordinates[coordinatesKey]; !ok {
			keys = append(keys, coordinatesKey)
		}
	}

	return keys, nil
}

func (h *help) addFolder(ctx context.Context, folder string) error {
	_, err := h.s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(h.databaseBucket),
		Key:    aws.String(folder),
	})

	return err
}

func (h *help) startCrawler(ctx context.Context) error {
	_, err := h.glueClient.StartCrawler(&glue.StartCrawlerInput{
		Name: aws.String(h.crawlerName),
	})

	return err
}
