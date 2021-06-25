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

	"github.com/cheesesteakio/api/pkg/acct"
	"github.com/cheesesteakio/api/pkg/fs"
)

const accountIDHeader = "x-cheesesteakstorage-account-id"

var demoAccountID = os.Getenv("DEMO_ACCOUNT_ID")

func handler(acctClient acct.Accounter, fsClient fs.Filesystemer) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

		filenames := []string{}
		if err := json.Unmarshal([]byte(request.Body), &filenames); err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       `{"error": "error unmarshalling filenames array"}`,
			}, nil
		}

		body := ""

		switch request.HTTPMethod {
		case http.MethodPost:
			presignedURLs := make([]string, len(filenames))
			for i, fileName := range filenames {
				presignedURL, err := fsClient.GenerateUploadURL(ctx, bucketName, accountID, fileName)
				if err != nil {
					return events.APIGatewayProxyResponse{
						StatusCode: http.StatusInternalServerError,
						Body:       fmt.Sprintf(`{"error": "error generating [%s] presigned url"}`, fileName),
					}, nil
				}

				presignedURLs[i] = presignedURL
			}

			presignedURLsBytes, err := json.Marshal(presignedURLs)
			if err != nil {
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       `{"error": "error marshalling presigned urls"}`,
				}, nil
			}

			body = fmt.Sprintf(`{"message": "success", "urls": %s}`, presignedURLsBytes)
		case http.MethodDelete:
			if err := fsClient.DeleteFiles(ctx, bucketName, accountID, filenames); err != nil {
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       `{"error": "error deleting files"}`,
				}, nil
			}

			body = `{"message": "success"}`
		default:
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
				Body:       fmt.Sprintf(`{"error": "http method [%s] not supported"}`, request.HTTPMethod),
			}, nil
		}

		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       body,
		}, nil
	}
}

func main() {
	newSession := session.New()

	acctClient := acct.New(newSession)

	fsClient, err := fs.New(newSession)
	if err != nil {
		log.Fatalf("error creating fs client: %s", err.Error())
	}

	lambda.Start(handler(acctClient, fsClient))
}
