package main

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

type mockEvtClient struct {
	mockAddBucketListenersError    error
	mockRemoveBucketListenersError error
}

func (m *mockEvtClient) AddBucketListeners(ctx context.Context, buckets []string) error {
	return m.mockAddBucketListenersError
}

func (m *mockEvtClient) RemoveBucketListeners(ctx context.Context, buckets []string) error {
	return m.mockRemoveBucketListenersError
}

func Test_handler(t *testing.T) {
	tests := []struct {
		description                    string
		request                        events.APIGatewayProxyRequest
		mockAddBucketListenersError    error
		mockRemoveBucketListenersError error
		statusCode                     int
		body                           string
	}{
		{
			description:                    "no security header received",
			request:                        events.APIGatewayProxyRequest{},
			mockAddBucketListenersError:    nil,
			mockRemoveBucketListenersError: nil,
			statusCode:                     400,
			body:                           `{"error": "security key header not provided"}`,
		},
		{
			description: "incorrect security header received",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"http-security-header": "incorrect-value",
				},
			},
			mockAddBucketListenersError:    nil,
			mockRemoveBucketListenersError: nil,
			statusCode:                     400,
			body:                           `{"error": "security key "incorrect-value" incorrect"}`,
		},
		{
			description: "unmarshal request body error",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"http-security-header": "http-security-header-value",
				},
				Body: "invalid-json",
			},
			mockAddBucketListenersError:    nil,
			mockRemoveBucketListenersError: nil,
			statusCode:                     400,
			body:                           `{"error": "invalid character 'i' looking for beginning of value"}`,
		},
		{
			description: "add bucket listeners error",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"http-security-header": "http-security-header-value",
				},
				Body: `{"add": ["bucket"]}`,
			},
			mockAddBucketListenersError:    errors.New("mock add bucket listeners error"),
			mockRemoveBucketListenersError: nil,
			statusCode:                     500,
			body:                           `{"error": "mock add bucket listeners error"}`,
		},
		{
			description: "remove bucket listeners error",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"http-security-header": "http-security-header-value",
				},
				Body: `{"remove": ["bucket"]}`,
			},
			mockAddBucketListenersError:    nil,
			mockRemoveBucketListenersError: errors.New("mock remove bucket listeners error"),
			statusCode:                     500,
			body:                           `{"error": "mock remove bucket listeners error"}`,
		},
		{
			description: "successful invocation",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"http-security-header": "http-security-header-value",
				},
				Body: `{"add":["add_bucket"], "remove": ["remove_bucket"]}`,
			},
			mockAddBucketListenersError:    nil,
			mockRemoveBucketListenersError: nil,
			statusCode:                     200,
			body:                           `{"message": "success", "buckets_added": 1, "buckets_removed": 1}`,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			evtClient := &mockEvtClient{
				mockAddBucketListenersError:    test.mockAddBucketListenersError,
				mockRemoveBucketListenersError: test.mockRemoveBucketListenersError,
			}

			handlerFunc := handler(evtClient, "http-security-header", "http-security-header-value")

			response, _ := handlerFunc(context.Background(), test.request)

			if response.StatusCode != test.statusCode {
				t.Errorf("incorrect status code, received: %d, expected: %d", test.statusCode, response.StatusCode)
			}

			if response.Body != test.body {
				t.Errorf("incorrect body, received: %q, expected: %q", test.body, response.Body)
			}
		})
	}
}
