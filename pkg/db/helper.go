package db

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

const (
	index        = "files"
	documentType = "file"
)

var _ helper = &help{}

type helper interface {
	executeCreate(ctx context.Context) error
	executeBulk(ctx context.Context, body io.Reader) error
	executeDelete(ctx context.Context, body io.Reader) error
	executeQuery(ctx context.Context, body io.Reader) (io.ReadCloser, error)
}

type help struct {
	opensearchClient *opensearch.Client
}

func (h *help) executeCreate(ctx context.Context) error {
	request := opensearchapi.IndicesCreateRequest{
		Index: index,
	}

	response, err := request.Do(ctx, h.opensearchClient)
	if err != nil {
		return nil
	}

	if err := checkResponse(response); err != nil {
		return err
	}

	return nil
}

func (h *help) executeBulk(ctx context.Context, body io.Reader) error {
	request := opensearchapi.BulkRequest{
		Index:        index,
		DocumentType: documentType,
		Body:         body,
		Timeout:      time.Duration(1 * time.Minute),
	}

	response, err := request.Do(ctx, h.opensearchClient)
	if err != nil {
		return err
	}

	if err := checkResponse(response); err != nil {
		return err
	}

	return nil
}

func (h *help) executeDelete(ctx context.Context, body io.Reader) error {
	request := opensearchapi.DeleteByQueryRequest{
		Index:        []string{index},
		DocumentType: []string{documentType},
		Body:         body,
	}

	response, err := request.Do(ctx, h.opensearchClient)
	if err != nil {
		return err
	}

	if err := checkResponse(response); err != nil {
		return err
	}

	return nil
}

func (h *help) executeQuery(ctx context.Context, body io.Reader) (io.ReadCloser, error) {
	request := opensearchapi.SearchRequest{
		Index:        []string{index},
		DocumentType: []string{documentType},
		Body:         body,
	}

	response, err := request.Do(ctx, h.opensearchClient)
	if err != nil {
		return nil, err
	}

	if err := checkResponse(response); err != nil {
		return nil, err
	}

	return response.Body, nil
}

func checkResponse(response *opensearchapi.Response) error {
	if response.IsError() {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}

		return errors.New(string(body))
	}

	return nil
}
