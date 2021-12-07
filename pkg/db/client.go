package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/opensearch-project/opensearch-go"

	"github.com/forstmeier/findfile/pkg/pars"
)

var _ Databaser = &Client{}

// Query holds the fields required for building an OpenSearch query
// from the values provided by the user.
type Query struct {
	Text string `json:"text"`
}

// Client implements the db.Databaser methods using AWS OpenSearch.
type Client struct {
	helper helper
}

// New generates a db.Client pointer instance with AWS OpenSearch.
func New(newSession *session.Session, url, username, password string) (*Client, error) {
	opensearchClient, err := opensearch.NewClient(opensearch.Config{
		Addresses: []string{url},
		Username:  username,
		Password:  password,
	})
	if err != nil {
		return nil, &NewClientError{
			err: err,
		}
	}

	return &Client{
		helper: &help{
			opensearchClient: opensearchClient,
		},
	}, nil
}

// SetupDatabase implements the db.Databaser.SetupDatabase method
// using AWS OpenSearch.
func (c *Client) SetupDatabase(ctx context.Context) error {
	if err := c.helper.executeCreate(ctx); err != nil {
		return &ExecuteCreateError{
			err: err,
		}
	}

	return nil
}

// UpsertDocuments implements the db.Databaser.UpsertDocuments method
// using AWS OpenSearch.
func (c *Client) UpsertDocuments(ctx context.Context, documents []pars.Document) error {
	if len(documents) == 0 {
		return nil
	}

	var body bytes.Buffer
	for _, document := range documents {
		metadata := fmt.Sprintf(`{ "index": { "_id": "%s" } }`, document.ID)
		body.WriteString(metadata + "\n")

		data, err := json.Marshal(document)
		if err != nil {
			return &MarshalDocumentError{
				err: err,
			}
		}
		body.Write(data)
		body.WriteString("\n")
	}

	if err := c.helper.executeBulk(ctx, &body); err != nil {
		return &ExecuteBulkError{
			err: err,
		}
	}

	return nil
}

// DeleteDocumentsByIDs implements the db.Databaser.DeleteDocumentsByIDs
// method using AWS OpenSearch.
func (c *Client) DeleteDocumentsByIDs(ctx context.Context, documentPaths []string) error {
	if len(documentPaths) == 0 {
		return nil
	}

	queryString := `{ "query": { "bool": { "minimum_should_match": 1, "should": [ %s ] } } }`

	matches := []string{}
	for _, documentPath := range documentPaths {
		documentInfo := strings.Split(documentPath, "/")
		match := fmt.Sprintf(`{ "match": { "file_bucket": "%s", "file_key": "%s" } }`, documentInfo[0], documentInfo[1])
		matches = append(matches, match)
	}

	queryString = fmt.Sprintf(queryString, strings.Join(matches, ", "))

	var body bytes.Buffer
	body.WriteString(queryString)

	if err := c.helper.executeDelete(ctx, &body); err != nil {
		return &ExecuteDeleteError{
			err: err,
		}
	}

	return nil
}

// DeleteDocumentsByBuckets implements the db.Databaser.DeleteDocumentsByBuckets
// method using AWS OpenSearch.
func (c *Client) DeleteDocumentsByBuckets(ctx context.Context, buckets []string) error {
	if len(buckets) == 0 {
		return nil
	}

	queryString := `{ "query": { "bool": { "minimum_should_match": 1, "should": [ %s ] } } }`

	matches := []string{}
	for _, bucket := range buckets {
		match := fmt.Sprintf(`{ "match": { "file_bucket": "%s" } }`, bucket)
		matches = append(matches, match)
	}

	queryString = fmt.Sprintf(queryString, strings.Join(matches, ", "))

	var body bytes.Buffer
	body.WriteString(queryString)

	if err := c.helper.executeDelete(ctx, &body); err != nil {
		return &ExecuteDeleteError{
			err: err,
		}
	}

	return nil
}

type queryResponseBody struct {
	Hits queryHits `json:"hits"`
}

type queryHits struct {
	Hits []hits `json:"hits"`
}

type hits struct {
	Source pars.Document `json:"_source"`
}

// QueryDocuments implements the db.Databaser.QueryDocuments method
// using AWS OpenSearch.
func (c *Client) QueryDocuments(ctx context.Context, query Query) ([]pars.Document, error) {
	if query.Text == "" {
		return []pars.Document{}, nil
	}

	queryString := fmt.Sprintf(`{ "query": { "match": { "pages.lines.text": { "query": "%s", "fuzziness": "AUTO" } } } }`, query.Text)

	response, err := c.helper.executeQuery(ctx, strings.NewReader(queryString))
	if err != nil {
		return nil, &ExecuteQueryError{
			err: err,
		}
	}

	body, err := io.ReadAll(response)
	if err != nil {
		return nil, &ReadQueryResponseBodyError{
			err: err,
		}
	}

	var responseBody queryResponseBody
	if err := json.Unmarshal(body, &responseBody); err != nil {
		return nil, &UnmarshalQueryResponseBodyError{
			err: err,
		}
	}

	documents := []pars.Document{}
	for _, document := range responseBody.Hits.Hits {
		documents = append(documents, document.Source)
	}

	return documents, nil
}
