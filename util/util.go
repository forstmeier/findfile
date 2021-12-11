package util

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

// Log provides a basic wrapper to format log output.
func Log(key string, value interface{}) {
	logMessage(key, value)
}

// SendResponse is a helper function for sending responses
// to API Gateway.
func SendResponse(statusCode int, payload interface{}, message string) (events.APIGatewayProxyResponse, error) {
	var body interface{}

	switch t := payload.(type) {
	case error:
		body = struct {
			Error string `json:"error"`
		}{
			Error: payload.(error).Error(),
		}

	case []string:
		body = struct {
			Message   string   `json:"message"`
			FilePaths []string `json:"file_paths"`
		}{
			Message:   "success",
			FilePaths: payload.([]string),
		}

	case map[string]int:
		body = struct {
			Message        string `json:"message"`
			BucketsAdded   int    `json:"buckets_added"`
			BucketsRemoved int    `json:"buckets_removed"`
		}{
			Message:        "success",
			BucketsAdded:   t["buckets_added"],
			BucketsRemoved: t["buckets_removed"],
		}

	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		logMessage("MARSHAL_RESPONSE_PAYLOAD_ERROR", err.Error())
		return events.APIGatewayProxyResponse{
			StatusCode:      http.StatusInternalServerError,
			Body:            fmt.Sprintf(`{"error": %q}`, err),
			IsBase64Encoded: false,
		}, err
	}

	logMessage("RESPONSE_BODY", string(bodyBytes))
	return events.APIGatewayProxyResponse{
		StatusCode:      statusCode,
		Body:            string(bodyBytes),
		IsBase64Encoded: false,
	}, nil
}

func logMessage(key string, value interface{}) {
	log.Printf(`{"%s": "%+v"}`, key, value)
}
