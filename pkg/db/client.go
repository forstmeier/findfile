package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"

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
func New(newSession *session.Session) (*Client, error) {
	elasticsearchClient, err := elasticsearch.NewDefaultClient()
	if err != nil {
		return nil, &NewClientError{
			err: err,
		}
	}

	return &Client{
		helper: &help{
			elasticsearchClient: elasticsearchClient,
		},
	}, nil
}

// SetupDatabase implements the db.Databaser.SetupDatabase method
// using AWS OpenSearch.
func (c *Client) SetupDatabase(ctx context.Context) error {
	// NOTE: this may be where domain index
	// mapping configuration is applied
	return nil
}

// UpsertDocuments implements the db.Databaser.UpsertDocuments method
// using AWS OpenSearch.
func (c *Client) UpsertDocuments(ctx context.Context, documents []pars.Document) error {
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

		if _, err = body.Write(data); err != nil {
			return &WriteDocumentDataError{
				err: err,
			}
		}

		body.WriteString("\n")
	}

	response, err := c.helper.executeBulk(ctx, &esapi.BulkRequest{
		Body: &body,
	})
	if response.IsError() || err != nil {
		return &ExecuteBulkError{
			err: err,
		}
	}

	return nil
}

// DeleteDocuments implements the db.Databaser.DeleteDocuments method
// using AWS OpenSearch.
func (c *Client) DeleteDocuments(ctx context.Context, documentPaths []string) error {
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

	response, err := c.helper.executeDelete(ctx, &esapi.DeleteByQueryRequest{
		Body: &body,
	})
	if response.IsError() || err != nil {
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
	Hits []pars.Document `json:"hits"`
}

type queryHit struct {
	// Fields map[string][]pars.Document `json:"fields"`
}

// QueryDocuments implements the db.Databaser.QueryDocuments method
// using AWS OpenSearch.
func (c *Client) QueryDocuments(ctx context.Context, query Query) ([]pars.Document, error) {
	queryString := fmt.Sprintf(`{ "query": { "nested": { "path": "pages.lines", "query": { "match": { "lines.text": "%s" } } } } }`, query.Text)

	response, err := c.helper.executeQuery(ctx, &esapi.SearchRequest{
		Body: strings.NewReader(queryString),
	})
	if response.IsError() || err != nil {
		return nil, &ExecuteQueryError{
			err: err,
		}
	}

	body, err := io.ReadAll(response.Body)
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

	return responseBody.Hits.Hits, nil
}
