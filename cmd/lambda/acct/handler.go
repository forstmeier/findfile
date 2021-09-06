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
	"github.com/cheesesteakio/api/pkg/db"
	"github.com/cheesesteakio/api/pkg/fs"
	"github.com/cheesesteakio/api/util"
)

var accountIDHeader = os.Getenv("ACCOUNT_ID_HTTP_HEADER")

func handler(acctClient acct.Accounter, partitionerClient db.Partitioner, fsClient fs.Filesystemer) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		util.Log("REQUEST_BODY", request.Body)
		util.Log("REQUEST_METHOD", request.HTTPMethod)

		body := ""

		switch request.HTTPMethod {
		case http.MethodPost:
			accountID := uuid.NewString()

			requestJSON := map[string]string{}
			if err := json.Unmarshal([]byte(request.Body), &requestJSON); err != nil {
				util.Log("UNMARSHAL_REQUEST_BODY_ERROR", err)
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       `{"error": "error unmarshalling request"}`,
				}, nil
			}

			bucketName := requestJSON["bucket_name"]
			if err := acctClient.CreateAccount(ctx, accountID, bucketName); err != nil {
				util.Log("CREATE_ACCOUNT_ERROR", err)
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       `{"error": "error creating user account"}`,
				}, nil
			}

			if err := partitionerClient.AddPartition(ctx, accountID); err != nil {
				util.Log("ADD_PARTITION_ERROR", err)
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       `{"error": "error adding account partition"}`,
				}, nil
			}

			if err := fsClient.CreateFileWatcher(ctx, bucketName); err != nil {
				util.Log("CREATE_FILE_WATCHER_ERROR", err)
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       `{"error": "error creating bucket notification"}`,
				}, nil
			}

			// NOTE: for the subscription logic, there would also be a check
			// for the existence of the required Stripe values and if the user
			// sent them the subscription would be created with a subscrClient
			// and the values would be added via the acctClient.UpdateAccount method.

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

			account, err := acctClient.GetAccountByID(ctx, accountID)
			if err != nil {
				util.Log("GET_ACCOUNT_BY_ID_ERROR", err)
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       `{"error": "error getting user account"}`,
				}, nil
			}

			if err := acctClient.DeleteAccount(ctx, accountID); err != nil {
				util.Log("DELETE_ACCOUNT_ERROR", err)
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       `{"error": "error removing user account"}`,
				}, nil
			}

			if err := partitionerClient.RemovePartition(ctx, accountID); err != nil {
				util.Log("REMOVE_PARTITION_ERROR", err)
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       `{"error": "error removing account partition"}`,
				}, nil
			}

			if err := fsClient.DeleteFileWatcher(ctx, account.BucketName); err != nil {
				util.Log("DELETE_FILE_WATCHER_ERROR", err)
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       `{"error": "error deleting bucket notification"}`,
				}, nil
			}

			// NOTE: for the delete logic, there would also be a check for the
			// existence of Stripe values and if true the subscription would be
			// removed with an added subscrClient in the handler.

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
