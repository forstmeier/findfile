package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"

	"github.com/cheesesteakio/api/pkg/acct"
	"github.com/cheesesteakio/api/pkg/cql"
	"github.com/cheesesteakio/api/pkg/db"
	"github.com/cheesesteakio/api/util"
)

var accountIDHeader = os.Getenv("ACCOUNT_ID_HTTP_HEADER")

func handler(acctClient acct.Accounter, cqlClient cql.CQLer, dbClient db.Databaser) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		util.Log("REQUEST_BODY", request.Body)
		util.Log("REQUEST_METHOD", request.HTTPMethod)

		accountID, ok := request.Headers[accountIDHeader]
		if !ok {
			util.Log("ACCOUNT_ID_ERROR", "account id not provided")
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
				Body:       `{"error": "account id not provided"}`,
			}, nil
		}

		if request.HTTPMethod != http.MethodPost {
			util.Log("HTTP_METHOD_ERROR", "http method not supported")
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
				Body:       fmt.Sprintf(`{"error": "http method [%s] not supported"}`, request.HTTPMethod),
			}, nil
		}

		account, err := acctClient.GetAccountByID(ctx, accountID)
		if err != nil {
			util.Log("READ_ACCOUNT_ERROR", err)
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       `{"error": "error getting account values"}`,
			}, nil
		}

		if account == nil {
			util.Log("ACCOUNT_ERROR", "nil account value")
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       fmt.Sprintf(`{"error": "account [%s] not found}`, accountID),
			}, nil
		}

		cqlJSON := map[string]interface{}{}
		if err := json.Unmarshal([]byte(request.Body), &cqlJSON); err != nil {
			util.Log("UNMARSHAL_REQUEST_BODY_ERROR", err)
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       `{"error": "error unmarshalling query"}`,
			}, nil
		}

		query, err := cqlClient.ConvertCQL(ctx, accountID, cqlJSON)
		if err != nil {
			util.Log("CONVERT_CQL_ERROR", err)
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       `{"error": "error converting cql to query"}`,
			}, nil
		}

		documents, err := dbClient.QueryDocuments(ctx, query)
		if err != nil {
			util.Log("QUERY_DOCUMENTS_ERROR", err)
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       `{"error": "error running query"}`,
			}, nil
		}

		filenames := make([]string, len(documents))
		for i, document := range documents {
			filenames[i] = document.Filename
		}

		output := struct {
			Message   string   `json:"message"`
			Filenames []string `json:"filenames"`
		}{
			Message:   "success",
			Filenames: filenames,
		}

		outputBytes, err := json.Marshal(output)
		if err != nil {
			util.Log("MARSHAL_OUTPUT_ERROR", err)
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       `{"error": "error marshalling presigned urls"}`,
			}, nil
		}

		util.Log("RESPONSE_BODY", string(outputBytes))
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       string(outputBytes),
		}, nil
	}
}
