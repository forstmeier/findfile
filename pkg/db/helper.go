package db

import (
	"context"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

var _ helper = &help{}

type helper interface {
	executeIndexMapping(ctx context.Context, request *esapi.IndicesPutMappingRequest) (*esapi.Response, error)
	executeBulk(ctx context.Context, request *esapi.BulkRequest) (*esapi.Response, error)
	executeDelete(ctx context.Context, request *esapi.DeleteByQueryRequest) (*esapi.Response, error)
	executeQuery(ctx context.Context, request *esapi.SearchRequest) (*esapi.Response, error)
}

type help struct {
	elasticsearchClient *elasticsearch.Client
}

func (h *help) executeIndexMapping(ctx context.Context, request *esapi.IndicesPutMappingRequest) (*esapi.Response, error) {
	return request.Do(ctx, h.elasticsearchClient)
}

func (h *help) executeBulk(ctx context.Context, request *esapi.BulkRequest) (*esapi.Response, error) {
	return request.Do(ctx, h.elasticsearchClient)
}

func (h *help) executeDelete(ctx context.Context, request *esapi.DeleteByQueryRequest) (*esapi.Response, error) {
	return request.Do(ctx, h.elasticsearchClient)
}

func (h *help) executeQuery(ctx context.Context, request *esapi.SearchRequest) (*esapi.Response, error) {
	return request.Do(ctx, h.elasticsearchClient)
}
