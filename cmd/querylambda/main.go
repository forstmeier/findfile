package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/cheesesteakio/api/pkg/acct"
	"github.com/cheesesteakio/api/pkg/csql"
	"github.com/cheesesteakio/api/pkg/db"
	"github.com/cheesesteakio/api/pkg/fs"
)

const accountIDHeader = "x-cheesesteakstorage-account-id"

var demoAccountID = os.Getenv("DEMO_ACCOUNT_ID")

func handler(acctClient acct.Accounter, csqlClient csql.CSQLer, dbClient db.Databaser, fsClient fs.Filesystemer) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		accountID, ok := request.Headers[accountIDHeader]
		if !ok {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
				Body:       `{"error": "account id not provided"}`,
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

			fileInfo := fs.FileInfo{
				Filepath: bucketName,
				Filename: document.Filename,
			}

			presignedURL, err := fsClient.GenerateDownloadURL(ctx, accountID, fileInfo)
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
	newSession := session.New()

	ddb, err := mongo.NewClient(nil)
	if err != nil {
		log.Fatalf("error creating mongo db client: %s", err.Error())
	}

	acctClient := acct.New(newSession)

	csqlClient := csql.New()

	dbClient, err := db.New(ddb, "main", "documents")
	if err != nil {
		log.Fatalf("error creating db client: %s", err.Error())
	}

	fsClient, err := fs.New(newSession)
	if err != nil {
		log.Fatalf("error creating fs client: %s", err.Error())
	}

	lambda.Start(handler(acctClient, csqlClient, dbClient, fsClient))
}
