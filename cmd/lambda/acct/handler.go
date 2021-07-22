package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"

	"github.com/cheesesteakio/api/pkg/acct"
	"github.com/cheesesteakio/api/pkg/fs"
	"github.com/cheesesteakio/api/pkg/subscr"
	"github.com/cheesesteakio/api/util"
)

var accountIDHeader = os.Getenv("ACCOUNT_ID_HTTP_HEADER")

func handler(acctClient acct.Accounter, subscrClient subscr.Subscriber, fsClient fs.Filesystemer) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		util.Log("REQUEST_BODY", request.Body)
		util.Log("REQUEST_METHOD", request.HTTPMethod)

		body := ""

		switch request.HTTPMethod {
		case http.MethodPost:
			subscriberInfo := subscr.SubscriberInfo{}
			if err := json.Unmarshal([]byte(request.Body), &subscriberInfo); err != nil {
				util.Log("UNMARSHAL_ERROR", err)
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       `{"error": "error unmarshalling subscriber info"}`,
				}, nil
			}
			subscriberInfo.ID = uuid.NewString()

			accountID := uuid.NewString()

			if err := acctClient.CreateAccount(ctx, accountID); err != nil {
				util.Log("CREATE_ACCOUNT_ERROR", err)
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       `{"error": "error creating user account"}`,
				}, nil
			}

			subscription, err := subscrClient.CreateSubscription(ctx, accountID, subscriberInfo)
			if err != nil {
				util.Log("CREATE_SUBSCRIPTION_ERROR", err)
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       `{"error": "error creating user subscription"}`,
				}, nil
			}

			subscriptionValues := map[string]string{
				acct.SubscriptionIDKey:        subscription.ID,
				acct.StripePaymentMethodIDKey: subscription.StripePaymentMethodID,
				acct.StripeCustomerIDKey:      subscription.StripeCustomerID,
				acct.StripeSubscriptionIDKey:  subscription.StripeSubscriptionID,
			}

			if err := acctClient.UpdateAccount(ctx, accountID, subscriptionValues); err != nil {
				util.Log("UPDATE_ACCOUNT_ERROR", err)
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       `{"error": "error adding subscription to user account"}`,
				}, nil
			}

			body = fmt.Sprintf(`{"message": "success", "account_id": "%s"}`, accountID)

		case http.MethodDelete:
			accountID, ok := request.Headers[accountIDHeader]
			if !ok {
				util.Log("ACCOUNT_ID_ERROR", "account id not provided")
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusBadRequest,
					Body:       `{"error": "account id not provided"}`,
				}, nil
			}

			account, err := acctClient.ReadAccount(ctx, accountID)
			if err != nil {
				util.Log("READ_ACCOUNT_ERROR", err)
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       `{"error": "error getting account values"}`,
				}, nil
			}

			subscription := subscr.Subscription{
				ID:                    account.SubscriptionID,
				StripePaymentMethodID: account.StripePaymentMethodID,
				StripeCustomerID:      account.StripeCustomerID,
				StripeSubscriptionID:  account.StripeSubscriptionID,
			}

			if err := subscrClient.RemoveSubscription(ctx, subscription); err != nil {
				util.Log("REMOVE_SUBSCRIPTION_ERROR", err)
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       `{"error": "error removing user subscription"}`,
				}, nil
			}

			if err := acctClient.DeleteAccount(ctx, accountID); err != nil {
				util.Log("DELETE_ACCOUNT_ERROR", err)
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       `{"error": "error removing user account"}`,
				}, nil
			}

			filesInfo, err := fsClient.ListFilesByAccountID(ctx, fs.MainBucket, accountID)
			if err != nil {
				util.Log("LIST_FILES_BY_ACCOUNT_ID_ERROR", err)
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       `{"error": "error listing user files"}`,
				}, nil
			}

			if err := fsClient.DeleteFiles(ctx, accountID, filesInfo); err != nil {
				util.Log("DELETE_FILES_ERROR", err)
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       `{"error": "error removing user files"}`,
				}, nil
			}

			body = fmt.Sprintf(`{"message": "success", "account_id": "%s"}`, accountID)

		default:
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
				Body:       fmt.Sprintf(`{"error": "http method [%s] not supported"}`, request.HTTPMethod),
			}, nil
		}

		util.Log("RESPONSE_BODY", body)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       body,
		}, nil
	}
}
