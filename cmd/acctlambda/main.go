package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/uuid"

	"github.com/cheesesteakio/api/pkg/acct"
	"github.com/cheesesteakio/api/pkg/subscr"
)

const accountIDHeader = "x-cheesesteakstorage-account-id"

func handler(acctClient acct.Accounter, subscrClient subscr.Subscriber) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		body := ""

		switch request.HTTPMethod {
		case http.MethodPost:
			subscriberInfo := subscr.SubscriberInfo{}
			if err := json.Unmarshal([]byte(request.Body), &subscriberInfo); err != nil {
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       `{"error": "error unmarshalling subscriber info"}`,
				}, nil
			}
			subscriberInfo.ID = uuid.NewString()

			accountID := uuid.NewString()

			if err := acctClient.CreateAccount(ctx, accountID); err != nil {
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       `{"error": "error creating user account"}`,
				}, nil
			}

			_, err := subscrClient.CreateSubscription(ctx, subscriberInfo)
			if err != nil {
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       `{"error": "error creating user subscription"}`,
				}, nil
			}

			body = fmt.Sprintf(`{"message": "success", "account_id": "%s"}`, accountID)

		case http.MethodDelete:
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

			if err := subscrClient.RemoveSubscription(ctx, account.StripeCustomerID); err != nil {
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       `{"error": "error removing user subscription"}`,
				}, nil
			}

			if err := acctClient.DeleteAccount(ctx, accountID); err != nil {
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       `{"error": "error removing user account"}`,
				}, nil
			}

			body = fmt.Sprintf(`{"message": "success", "account_id": "%s"}`, accountID)

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
	acctClient := acct.New()

	subscrClient := subscr.New()

	lambda.Start(handler(acctClient, subscrClient))
}
