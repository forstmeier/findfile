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
			return util.SendResponse(
				http.StatusBadRequest,
				fmt.Errorf("security key header '%s' not provided", httpSecurityHeader),
				"SECURITY_KEY_HEADER_ERROR",
			)
		}

		if httpSecurityKeyReceived != httpSecurityKey {
			return util.SendResponse(
				http.StatusBadRequest,
				fmt.Errorf("security key '%s' incorrect", httpSecurityKeyReceived),
				"SECURITY_KEY_VALUE_ERROR",
			)
		}

		requestJSON := db.Query{}
		if err := json.Unmarshal([]byte(request.Body), &requestJSON); err != nil {
			return util.SendResponse(
				http.StatusBadRequest,
				err,
				"UNMARSHAL_REQUEST_PAYLOAD_ERROR",
			)
		}

		documents, err := dbClient.QueryDocuments(ctx, requestJSON)
		if err != nil {
			return util.SendResponse(
				http.StatusInternalServerError,
				err,
				"QUERY_DOCUMENTS_ERROR",
			)
		}

		filePaths := []string{}
		for _, document := range documents {
			filePaths = append(filePaths, fmt.Sprintf("%s/%s", document.FileBucket, document.FileKey))
		}

		return util.SendResponse(
			http.StatusOK,
			filePaths,
			"RESPONSE_BODY",
		)
	}
}
