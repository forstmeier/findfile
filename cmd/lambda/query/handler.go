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
	"github.com/cheesesteakio/api/pkg/fs"
	"github.com/cheesesteakio/api/util"
)

var (
	accountIDHeader = os.Getenv("ACCOUNT_ID_HTTP_HEADER")
	demoAccountID   = os.Getenv("DEMO_ACCOUNT_ID")
)

func handler(acctClient acct.Accounter, cqlClient cql.CQLer, dbClient db.Databaser, fsClient fs.Filesystemer) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

		isDemo := accountID == demoAccountID

		bucketName := fs.DemoBucket
		if !isDemo {
			bucketName = fs.MainBucket
		}

		if !isDemo {
			account, err := acctClient.ReadAccount(ctx, accountID)
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
		}

		query := map[string]interface{}{}
		if err := json.Unmarshal([]byte(request.Body), &query); err != nil {
			util.Log("UNMARSHAL_REQUEST_BODY_ERROR", err)
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       `{"error": "error unmarshalling query"}`,
			}, nil
		}

		cqlQuery, err := cqlClient.ConvertCQL(ctx, accountID, query)
		if err != nil {
			util.Log("CONVERT_CQL_ERROR", err)
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       `{"error": "error converting query to cql"}`,
			}, nil
		}

		documents, err := dbClient.QueryDocuments(ctx, cqlQuery)
		if err != nil {
			util.Log("QUERY_DOCUMENTS_ERROR", err)
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       fmt.Sprintf(`{"error": "error runing [%s] query"}`, cqlQuery),
			}, nil
		}

		filenames := make([]string, len(documents))
		presignedURLs := make([]string, len(documents))
		for i, document := range documents {
			filenames[i] = document.Filename

			fileInfo := fs.FileInfo{
				Filepath: bucketName,
				Filename: document.Filename,
			}

			presignedURL, err := fsClient.GenerateDownloadURL(ctx, accountID, fileInfo)
			if err != nil {
				util.Log("GENERATE_DOWNLOAD_URL_ERROR", err)
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       fmt.Sprintf(`{"error": "error generating [%s] presigned url"}`, document.Filename),
				}, nil
			}
			presignedURLs[i] = presignedURL
		}

		output := struct {
			Message       string   `json:"message"`
			Filenames     []string `json:"filenames"`
			PresignedURLs []string `json:"presigned_urls"`
		}{
			Message:       "success",
			Filenames:     filenames,
			PresignedURLs: presignedURLs,
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
