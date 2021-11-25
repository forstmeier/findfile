package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"

	"github.com/forstmeier/findfile/pkg/evt"
	"github.com/forstmeier/findfile/util"
)

type requestsPayload struct {
	Add    []string `json:"add,omitempty"`
	Remove []string `json:"remove,omitempty"`
}

func handler(evtClient evt.Eventer, httpSecurityHeader, httpSecurityKey string) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		httpSecurityKeyReceived, ok := request.Headers[httpSecurityHeader]
		if !ok {
			util.Log("SECURITY_KEY_HEADER_ERROR", fmt.Sprintf("security key header %q not provided", httpSecurityHeader))
			return events.APIGatewayProxyResponse{
				StatusCode:      http.StatusBadRequest,
				Body:            `{"error": "security key header not provided"}`,
				IsBase64Encoded: false,
			}, nil
		}

		if httpSecurityKeyReceived != httpSecurityKey {
			util.Log("SECURITY_KEY_VALUE_ERROR", fmt.Sprintf("security key %q incorrect", httpSecurityKeyReceived))
			return events.APIGatewayProxyResponse{
				StatusCode:      http.StatusBadRequest,
				Body:            fmt.Sprintf(`{"error": "security key %q incorrect"}`, httpSecurityKeyReceived),
				IsBase64Encoded: false,
			}, nil
		}

		requestJSON := requestsPayload{}
		if err := json.Unmarshal([]byte(request.Body), &requestJSON); err != nil {
			util.Log("UNMARSHAL_REQUEST_PAYLOAD_ERROR", err.Error())
			return events.APIGatewayProxyResponse{
				StatusCode:      http.StatusBadRequest,
				Body:            fmt.Sprintf(`{"error": %q}`, err),
				IsBase64Encoded: false,
			}, nil
		}

		if requestJSON.Add != nil {
			if err := evtClient.AddBucketListeners(ctx, requestJSON.Add); err != nil {
				util.Log("ADD_BUCKET_LISTENERS_ERROR", err.Error())
				return events.APIGatewayProxyResponse{
					StatusCode:      http.StatusInternalServerError,
					Body:            fmt.Sprintf(`{"error": %q}`, err),
					IsBase64Encoded: false,
				}, nil
			}
		}

		if requestJSON.Remove != nil {
			if err := evtClient.RemoveBucketListeners(ctx, requestJSON.Add); err != nil {
				util.Log("REMOVE_BUCKET_LISTENERS_ERROR", err.Error())
				return events.APIGatewayProxyResponse{
					StatusCode:      http.StatusInternalServerError,
					Body:            fmt.Sprintf(`{"error": %q}`, err),
					IsBase64Encoded: false,
				}, nil
			}
		}

		outputBody := fmt.Sprintf(`{"message": "success", "buckets_added": %d, "buckets_removed": %d}`, len(requestJSON.Add), len(requestJSON.Remove))

		util.Log("RESPONSE_BODY", outputBody)
		return events.APIGatewayProxyResponse{
			StatusCode:      http.StatusOK,
			Body:            outputBody,
			IsBase64Encoded: false,
		}, nil
	}
}
