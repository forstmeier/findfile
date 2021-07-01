package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"

	"github.com/cheesesteakio/api/pkg/acct"
	"github.com/cheesesteakio/api/pkg/fs"
)

var (
	accountIDHeader = os.Getenv("ACCOUNT_ID_HTTP_HEADER")
	demoAccountID   = os.Getenv("DEMO_ACCOUNT_ID")
)

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
			for i, fileame := range filenames {
				fileInfo := fs.FileInfo{
					Filepath: bucketName,
					Filename: fileame,
				}

				presignedURL, err := fsClient.GenerateUploadURL(ctx, accountID, fileInfo)
				if err != nil {
					return events.APIGatewayProxyResponse{
						StatusCode: http.StatusInternalServerError,
						Body:       fmt.Sprintf(`{"error": "error generating [%s] presigned url"}`, fileame),
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
			filesInfo := make([]fs.FileInfo, len(filenames))
			for i, filename := range filenames {
				filesInfo[i] = fs.FileInfo{
					Filepath: bucketName,
					Filename: filename,
				}
			}

			if err := fsClient.DeleteFiles(ctx, accountID, filesInfo); err != nil {
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
