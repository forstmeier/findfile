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

func handler(
	dbClient db.Databaser,
	httpSecurityHeader, httpSecurityKey string,
) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		httpSecurityKeyReceived, ok := request.Headers[httpSecurityHeader]
		if !ok {
			util.Log("SECURITY_KEY_HEADER_ERROR", fmt.Sprintf("security key header %q not provided", httpSecurityHeader))
			return sendResponse(
				http.StatusBadRequest,
				`{"error": "security key header not provided"}`,
			)
		}

		if httpSecurityKeyReceived != httpSecurityKey {
			util.Log("SECURITY_KEY_VALUE_ERROR", fmt.Sprintf("security key '%s' incorrect", httpSecurityKeyReceived))
			return sendResponse(
				http.StatusBadRequest,
				fmt.Sprintf(`{"error": "security key '%s' incorrect"}`, httpSecurityKeyReceived),
			)
		}

		requestJSON := db.Query{}
		if err := json.Unmarshal([]byte(request.Body), &requestJSON); err != nil {
			util.Log("UNMARSHAL_REQUEST_PAYLOAD_ERROR", err.Error())
			return sendResponse(
				http.StatusBadRequest,
				fmt.Sprintf(`{"error": %q}`, err),
			)
		}

		documents, err := dbClient.QueryDocuments(ctx, requestJSON)
		if err != nil {
			util.Log("QUERY_DOCUMENTS_ERROR", err.Error())
			return sendResponse(
				http.StatusInternalServerError,
				fmt.Sprintf(`{"error": %q}`, err),
			)
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
			return sendResponse(
				http.StatusInternalServerError,
				fmt.Sprintf(`{"error": %q}`, err),
			)
		}

		util.Log("RESPONSE_BODY", string(responseBodyBytes))
		return sendResponse(http.StatusOK, string(responseBodyBytes))
	}
}

func sendResponse(statusCode int, body string) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode:      statusCode,
		Body:            body,
		IsBase64Encoded: false,
	}, nil
}
