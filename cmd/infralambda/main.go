package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/cheesesteakio/api/pkg/infra"
)

const (
	accountIDParameter  = "account_id"
	filesystemParameter = "create_filesystem"
	databaseParameter   = "create_database"
)

func handler(infraClient infra.Infrastructurer) func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		parameters := req.QueryStringParameters
		if req.HTTPMethod == http.MethodPost {
			if output, ok := validateCreateParameters(parameters); !ok {
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusBadRequest,
					Body:       fmt.Sprintf(`{"error": "missing parameters [%s]"}`, output),
				}, nil
			}

			accountID := parameters[accountIDParameter]
			if strings.ToLower(parameters[filesystemParameter]) == "true" {
				if err := infraClient.CreateFilesystem(ctx, accountID); err != nil {
					return events.APIGatewayProxyResponse{
						StatusCode: http.StatusInternalServerError,
						Body:       `{"error": "error creating filesystem"}`,
					}, nil
				}
			}

			if strings.ToLower(parameters[databaseParameter]) == "true" {
				if err := infraClient.CreateDatabase(ctx, accountID); err != nil {
					return events.APIGatewayProxyResponse{
						StatusCode: http.StatusInternalServerError,
						Body:       `{"error": "error creating database"}`,
					}, nil
				}
			}

		} else if req.HTTPMethod == http.MethodDelete {
			accountID, ok := parameters[accountIDParameter]
			if !ok {
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusBadRequest,
					Body:       `{"error": "missing parameter [account id]"}`,
				}, nil
			}

			if err := infraClient.DeleteFilesystem(ctx, accountID); err != nil {
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       `{"error": "error deleting filesystem"}`,
				}, nil
			}

			if err := infraClient.DeleteDatabase(ctx, accountID); err != nil {
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       `{"error": "error deleting database"}`,
				}, nil
			}

		} else {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
				Body:       fmt.Sprintf(`{"error": "method not supported [%s]"}`, req.HTTPMethod),
			}, nil
		}

		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       fmt.Sprintf(`{"message": "[%s] success"}`, req.HTTPMethod),
		}, nil
	}
}

func validateCreateParameters(parameters map[string]string) (string, bool) {
	output := []string{}
	if _, ok := parameters[accountIDParameter]; !ok {
		output = append(output, "account id")
	}
	if _, ok := parameters[filesystemParameter]; !ok {
		output = append(output, "create filesystem")
	}
	if _, ok := parameters[databaseParameter]; !ok {
		output = append(output, "create database")
	}

	if len(output) > 0 {
		return strings.Join(output, ","), false
	}

	return "", true
}

func main() {
	infraClient := infra.New()

	lambda.Start(handler(infraClient))
}
