package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"

	"github.com/forstmeier/findfile/pkg/db"
	"github.com/forstmeier/findfile/pkg/fql"
	"github.com/forstmeier/findfile/util"
)

func handler(fqlClient fql.FQLer, dbClient db.Databaser, httpSecurityHeader, httpSecurityKey string) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		util.Log("REQUEST_BODY", request.Body)
		util.Log("REQUEST_METHOD", request.HTTPMethod)

		httpSecurityKeyReceived, ok := request.Headers[httpSecurityHeader]
		if !ok {
			util.Log("SECURITY_KEY_HEADER_ERROR", fmt.Sprintf("security key header %s not provided", httpSecurityHeader))
			return events.APIGatewayProxyResponse{
				StatusCode:      http.StatusBadRequest,
				Body:            `{"error": "security key header not provided"}`,
				IsBase64Encoded: false,
			}, nil
		}

		if httpSecurityKeyReceived != httpSecurityKey {
			util.Log("SECURITY_KEY_VALUE_ERROR", fmt.Sprintf("security key %s incorrect", httpSecurityKeyReceived))
			return events.APIGatewayProxyResponse{
				StatusCode:      http.StatusBadRequest,
				Body:            fmt.Sprintf(`{"error": "security key %s incorrect"}`, httpSecurityKeyReceived),
				IsBase64Encoded: false,
			}, nil
		}

		if request.HTTPMethod != http.MethodPost {
			util.Log("HTTP_METHOD_ERROR", fmt.Sprintf("http method %s not supported", request.HTTPMethod))
			return events.APIGatewayProxyResponse{
				StatusCode:      http.StatusBadRequest,
				Body:            fmt.Sprintf(`{"error": "http method %s not supported"}`, request.HTTPMethod),
				IsBase64Encoded: false,
			}, nil
		}

		query, err := fqlClient.ConvertFQL(ctx, []byte(request.Body))
		if err != nil {
			util.Log("CONVERT_FQL_ERROR", err.Error())
			return events.APIGatewayProxyResponse{
				StatusCode:      http.StatusInternalServerError,
				Body:            fmt.Sprintf(`{"error": "%s"}`, err.Error()),
				IsBase64Encoded: false,
			}, nil
		}

		documents, err := dbClient.QueryDocumentsByFQL(ctx, query)
		if err != nil {
			util.Log("QUERY_DOCUMENTS_ERROR", err.Error())
			return events.APIGatewayProxyResponse{
				StatusCode:      http.StatusInternalServerError,
				Body:            fmt.Sprintf(`{"error": "%s"}`, err.Error()),
				IsBase64Encoded: false,
			}, nil
		}

		data := map[string][]string{}
		for _, document := range documents {
			if _, ok := data[document.FileBucket]; ok {
				data[document.FileBucket] = append(data[document.FileBucket], document.FileKey)
			} else {
				data[document.FileBucket] = []string{document.FileKey}
			}
		}

		output := struct {
			Message string              `json:"message"`
			Data    map[string][]string `json:"data"`
		}{
			Message: "success",
			Data:    data,
		}

		outputBytes, err := json.Marshal(output)
		if err != nil {
			util.Log("MARSHAL_OUTPUT_ERROR", err.Error())
			return events.APIGatewayProxyResponse{
				StatusCode:      http.StatusInternalServerError,
				Body:            fmt.Sprintf(`{"error": "%s"}`, err.Error()),
				IsBase64Encoded: false,
			}, nil
		}

		util.Log("RESPONSE_BODY", string(outputBytes))
		return events.APIGatewayProxyResponse{
			StatusCode:      http.StatusOK,
			Body:            string(outputBytes),
			IsBase64Encoded: false,
		}, nil
	}
}
