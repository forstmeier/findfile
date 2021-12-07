package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"

	"github.com/forstmeier/findfile/pkg/db"
	"github.com/forstmeier/findfile/pkg/evt"
	"github.com/forstmeier/findfile/pkg/fs"
	"github.com/forstmeier/findfile/pkg/pars"
	"github.com/forstmeier/findfile/util"
)

type requestsPayload struct {
	Add    []string `json:"add,omitempty"`
	Remove []string `json:"remove,omitempty"`
}

func handler(evtClient evt.Eventer, fsClient fs.Filesystemer, parsClient pars.Parser, dbClient db.Databaser, httpSecurityHeader, httpSecurityKey string) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		httpSecurityKeyReceived, ok := request.Headers[httpSecurityHeader]
		if !ok {
			util.Log("SECURITY_KEY_HEADER_ERROR", fmt.Sprintf("security key header %q not provided", httpSecurityHeader))
			return events.APIGatewayProxyResponse{
				StatusCode:      http.StatusBadRequest,
				Body:            `{"error": "security key header not provided"}`,
				IsBase64Encoded: false,
			}, nil
		}

		if httpSecurityKeyReceived != httpSecurityKey {
			util.Log("SECURITY_KEY_VALUE_ERROR", fmt.Sprintf("security key %q incorrect", httpSecurityKeyReceived))
			return events.APIGatewayProxyResponse{
				StatusCode:      http.StatusBadRequest,
				Body:            fmt.Sprintf(`{"error": "security key %q incorrect"}`, httpSecurityKeyReceived),
				IsBase64Encoded: false,
			}, nil
		}

		requestJSON := requestsPayload{}
		if err := json.Unmarshal([]byte(request.Body), &requestJSON); err != nil {
			util.Log("UNMARSHAL_REQUEST_PAYLOAD_ERROR", err.Error())
			return events.APIGatewayProxyResponse{
				StatusCode:      http.StatusBadRequest,
				Body:            fmt.Sprintf(`{"error": %q}`, err),
				IsBase64Encoded: false,
			}, nil
		}

		if requestJSON.Add != nil {
			if err := evtClient.AddBucketListeners(ctx, requestJSON.Add); err != nil {
				util.Log("ADD_BUCKET_LISTENERS_ERROR", err.Error())
				return events.APIGatewayProxyResponse{
					StatusCode:      http.StatusInternalServerError,
					Body:            fmt.Sprintf(`{"error": %q}`, err),
					IsBase64Encoded: false,
				}, nil
			}

			for _, bucket := range requestJSON.Add {
				fileKeys, err := fsClient.ListFiles(ctx, bucket)
				if err != nil {
					util.Log("LIST_FILES_ERROR", err.Error())
					return events.APIGatewayProxyResponse{
						StatusCode:      http.StatusInternalServerError,
						Body:            fmt.Sprintf(`{"error": %q}`, err),
						IsBase64Encoded: false,
					}, nil
				}

				documents := make([]pars.Document, len(fileKeys))
				for i, fileKey := range fileKeys {
					if !hasSuffix(fileKey) {
						continue
					}

					document, err := parsClient.Parse(ctx, bucket, fileKey)
					if err != nil {
						util.Log("PARSE_ERROR", err.Error())
						return events.APIGatewayProxyResponse{
							StatusCode:      http.StatusInternalServerError,
							Body:            fmt.Sprintf(`{"error": %q}`, err),
							IsBase64Encoded: false,
						}, nil
					}

					documents[i] = *document
				}

				if err := dbClient.UpsertDocuments(ctx, documents); err != nil {
					util.Log("UPSERT_DOCUMENTS_ERROR", err.Error())
					return events.APIGatewayProxyResponse{
						StatusCode:      http.StatusInternalServerError,
						Body:            fmt.Sprintf(`{"error": %q}`, err),
						IsBase64Encoded: false,
					}, nil
				}
			}
		}

		if requestJSON.Remove != nil {
			if err := evtClient.RemoveBucketListeners(ctx, requestJSON.Remove); err != nil {
				util.Log("REMOVE_BUCKET_LISTENERS_ERROR", err.Error())
				return events.APIGatewayProxyResponse{
					StatusCode:      http.StatusInternalServerError,
					Body:            fmt.Sprintf(`{"error": %q}`, err),
					IsBase64Encoded: false,
				}, nil
			}

			if err := dbClient.DeleteDocumentsByBuckets(ctx, requestJSON.Remove); err != nil {
				util.Log("DELETE_DOCUMENTS_BY_BUCKETS_ERROR", err.Error())
				return events.APIGatewayProxyResponse{
					StatusCode:      http.StatusInternalServerError,
					Body:            fmt.Sprintf(`{"error": %q}`, err),
					IsBase64Encoded: false,
				}, nil
			}
		}

		outputBody := fmt.Sprintf(`{"message": "success", "buckets_added": %d, "buckets_removed": %d}`, len(requestJSON.Add), len(requestJSON.Remove))

		util.Log("RESPONSE_BODY", outputBody)
		return events.APIGatewayProxyResponse{
			StatusCode:      http.StatusOK,
			Body:            outputBody,
			IsBase64Encoded: false,
		}, nil
	}
}

func hasSuffix(fileKey string) bool {
	suffixes := []string{"png", "jpg", "jpeg", "tiff", "pdf"}

	for _, suffix := range suffixes {
		if strings.HasSuffix(fileKey, suffix) {
			return true
		}
	}

	return false
}
