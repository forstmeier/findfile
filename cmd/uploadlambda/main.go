package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/cheesesteakio/api/pkg/fs"
)

func handler(fsClient fs.Filesystemer) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		if request.HTTPMethod != http.MethodPost {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
				Body:       fmt.Sprintf(`{"error": "http method [%s] not supported"}`, request.HTTPMethod),
			}, nil
		}

		fileNames := []string{}
		if err := json.Unmarshal([]byte(request.Body), &fileNames); err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       `{"error": "error unmarshalling filenames array"}`,
			}, nil
		}

		presignedURLs := make([]string, len(fileNames))
		for i, fileName := range fileNames {
			presignedURL, err := fsClient.GenerateUploadURL(fileName)
			if err != nil {
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       `{"error": "error generating presigned urls"}`,
				}, nil
			}

			presignedURLs[i] = presignedURL
		}

		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       `{"message": "success"}`,
		}, nil
	}
}

func main() {
	fsClient, err := fs.New()
	if err != nil {
		log.Fatalf("error creating filesystem client: %s", err.Error())
	}

	lambda.Start(handler(fsClient))
}
