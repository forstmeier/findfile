package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"

	"github.com/forstmeier/findfile/pkg/db"
	"github.com/forstmeier/findfile/util"
)

func handler(dbClient db.Databaser, httpSecurityHeader, httpSecurityKey string) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

		requestJSON := db.Query{}
		if err := json.Unmarshal([]byte(request.Body), &requestJSON); err != nil {
			util.Log("UNMARSHAL_REQUEST_PAYLOAD_ERROR", err.Error())
			return events.APIGatewayProxyResponse{
				StatusCode:      http.StatusBadRequest,
				Body:            fmt.Sprintf(`{"error": %q}`, err),
				IsBase64Encoded: false,
			}, nil
		}

		documents, err := dbClient.QueryDocuments(ctx, requestJSON)
		if err != nil {
			util.Log("QUERY_DOCUMENTS_ERROR", err.Error())
			return events.APIGatewayProxyResponse{
				StatusCode:      http.StatusInternalServerError,
				Body:            fmt.Sprintf(`{"error": %q}`, err),
				IsBase64Encoded: false,
			}, nil
		}

		filePaths := []string{}
		for _, document := range documents {
			filePaths = append(filePaths, fmt.Sprintf("%s/%s", document.FileBucket, document.FileKey))
		}

		responseBody := struct {
			Message   string   `json:"message"`
			FilePaths []string `json:"file_paths"`
		}{
			Message:   "success",
			FilePaths: filePaths,
		}

		responseBodyBytes, err := json.Marshal(responseBody)
		if err != nil {
			util.Log("MARSHAL_RESPONSE_PAYLOAD_ERROR", err.Error())
			return events.APIGatewayProxyResponse{
				StatusCode:      http.StatusInternalServerError,
				Body:            fmt.Sprintf(`{"error": %q}`, err),
				IsBase64Encoded: false,
			}, nil
		}

		util.Log("RESPONSE_BODY", string(responseBodyBytes))
		return events.APIGatewayProxyResponse{
			StatusCode:      http.StatusOK,
			Body:            string(responseBodyBytes),
			IsBase64Encoded: false,
		}, nil
	}
}
