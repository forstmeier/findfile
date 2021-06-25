package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/cheesesteakio/api/pkg/acct"
	"github.com/cheesesteakio/api/pkg/csql"
	"github.com/cheesesteakio/api/pkg/db"
	"github.com/cheesesteakio/api/pkg/fs"
)

const accountIDHeader = "x-cheesesteakstorage-account-id"

func handler(acctClient acct.Accounter, csqlClient csql.CSQLer, dbClient db.Databaser, fsClient fs.Filesystemer) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		accountID, ok := request.Headers[accountIDHeader]
		if !ok {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
				Body:       `{"error": "account id not provided"}`,
			}, nil
		}

		account, err := acctClient.ReadAccount(ctx, accountID)

		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       `{"error": "error getting account values"}`,
			}, nil
		}

		if account == nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       fmt.Sprintf(`{"error": "account [%s] not found}`, accountID),
			}, nil
		}

		query := map[string]interface{}{}
		if err := json.Unmarshal([]byte(request.Body), &query); err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       `{"error": "error unmarshalling query"}`,
			}, nil
		}

		query["account_id"] = accountID

		csqlQuery, err := csqlClient.ConvertCSQL(query)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       `{"error": "error converting query to csql"}`,
			}, nil
		}

		documents, err := dbClient.QueryDocuments(ctx, csqlQuery)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       fmt.Sprintf(`{"error": "error runing [%s] query"}`, csqlQuery),
			}, nil
		}

		filenames := make([]string, len(documents))
		presignedURLs := make([]string, len(documents))
		for i, document := range documents {
			filenames[i] = document.Filename

			presignedURL, err := fsClient.GenerateDownloadURL(ctx, accountID, document.Filename)
			if err != nil {
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
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       `{"error": "error marshalling presigned urls"}`,
			}, nil
		}

		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       string(outputBytes),
		}, nil
	}
}

func main() {
	acctClient := acct.New()

	csqlClient := csql.New()

	dbClient, err := db.New("main", "documents")
	if err != nil {
		log.Fatalf("error creating db client: %s", err.Error())
	}

	fsClient, err := fs.New()
	if err != nil {
		log.Fatalf("error creating fs client: %s", err.Error())
	}

	lambda.Start(handler(acctClient, csqlClient, dbClient, fsClient))
}