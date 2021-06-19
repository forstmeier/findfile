package main

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

type mockFSClient struct {
	mockGeneratePresignedURLOutput string
	mockGeneratePresignedURLError  error
}

func (m *mockFSClient) GenerateUploadURL(filename string) (string, error) {
	return m.mockGeneratePresignedURLOutput, m.mockGeneratePresignedURLError
}

func Test_handler(t *testing.T) {
	tests := []struct {
		description                    string
		request                        events.APIGatewayProxyRequest
		mockGeneratePresignedURLOutput string
		mockGeneratePresignedURLError  error
		statusCode                     int
		body                           string
	}{
		{
			description: "incorrect http method received",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodGet,
			},
			mockGeneratePresignedURLOutput: "",
			mockGeneratePresignedURLError:  nil,
			statusCode:                     http.StatusBadRequest,
			body:                           `{"error": "http method [GET] not supported"}`,
		},
		{
			description: "error unmarshalling request body",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodPost,
				Body:       "---------",
			},
			mockGeneratePresignedURLOutput: "",
			mockGeneratePresignedURLError:  nil,
			statusCode:                     http.StatusInternalServerError,
			body:                           `{"error": "error unmarshalling filenames array"}`,
		},
		{
			description: "error generating presigned url",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodPost,
				Body:       `["filename.jpg"]`,
			},
			mockGeneratePresignedURLOutput: "",
			mockGeneratePresignedURLError:  errors.New("mock generate presigned url error"),
			statusCode:                     http.StatusInternalServerError,
			body:                           `{"error": "error generating presigned urls"}`,
		},
		{
			description: "successful handler invocation",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodPost,
				Body:       `["filename.jpg"]`,
			},
			mockGeneratePresignedURLOutput: "https//presigned-url/filename.jpg",
			mockGeneratePresignedURLError:  nil,
			statusCode:                     http.StatusOK,
			body:                           `{"message": "success"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &mockFSClient{
				mockGeneratePresignedURLOutput: test.mockGeneratePresignedURLOutput,
				mockGeneratePresignedURLError:  test.mockGeneratePresignedURLError,
			}

			handlerFunc := handler(client)

			response, _ := handlerFunc(context.Background(), test.request)

			if response.StatusCode != test.statusCode {
				t.Errorf("incorrect status code, received: %d, expected: %d", response.StatusCode, test.statusCode)
			}

			if response.Body != test.body {
				t.Errorf("incorrect body, received: %s, expected: %s", response.Body, test.body)
			}
		})
	}
}
