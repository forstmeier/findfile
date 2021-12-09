package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"

	"github.com/forstmeier/findfile/pkg/db"
	"github.com/forstmeier/findfile/pkg/evt"
	"github.com/forstmeier/findfile/pkg/fs"
	"github.com/forstmeier/findfile/pkg/pars"
	"github.com/forstmeier/findfile/util"
)

type requestsPayload struct {
	Add    []string `json:"add,omitempty"`
	Remove []string `json:"remove,omitempty"`
}

func handler(
	evtClient evt.Eventer,
	fsClient fs.Filesystemer,
	parsClient pars.Parser,
	dbClient db.Databaser,
	httpSecurityHeader, httpSecurityKey string,
) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		httpSecurityKeyReceived, ok := request.Headers[httpSecurityHeader]
		if !ok {
			util.Log("SECURITY_KEY_HEADER_ERROR", fmt.Sprintf("security key header %q not provided", httpSecurityHeader))
			return sendResponse(
				http.StatusBadRequest,
				`{"error": "security key header not provided"}`,
			)
		}

		if httpSecurityKeyReceived != httpSecurityKey {
			util.Log("SECURITY_KEY_VALUE_ERROR", fmt.Sprintf("security key %q incorrect", httpSecurityKeyReceived))
			return sendResponse(
				http.StatusBadRequest,
				fmt.Sprintf(`{"error": "security key %q incorrect"}`, httpSecurityKeyReceived),
			)
		}

		requestJSON := requestsPayload{}
		if err := json.Unmarshal([]byte(request.Body), &requestJSON); err != nil {
			util.Log("UNMARSHAL_REQUEST_PAYLOAD_ERROR", err.Error())
			return sendResponse(
				http.StatusBadRequest,
				fmt.Sprintf(`{"error": %q}`, err),
			)
		}

		if requestJSON.Add != nil {
			if err := evtClient.AddBucketListeners(ctx, requestJSON.Add); err != nil {
				util.Log("ADD_BUCKET_LISTENERS_ERROR", err.Error())
				return sendResponse(
					http.StatusInternalServerError,
					fmt.Sprintf(`{"error": %q}`, err),
				)
			}

			ctx, cancel := context.WithTimeout(ctx, 25*time.Second)
			defer cancel()

			for _, bucket := range requestJSON.Add {
				fileKeys, err := fsClient.ListFiles(ctx, bucket)
				if err != nil {
					util.Log("LIST_FILES_ERROR", err.Error())
					return sendResponse(
						http.StatusInternalServerError,
						fmt.Sprintf(`{"error": %q}`, err),
					)
				}

				chunk := 25
				for i := 0; i < len(fileKeys); i += chunk {
					var waitGroup sync.WaitGroup

					end := i + chunk
					if end > len(fileKeys) {
						end = len(fileKeys)
					}

					documentsChannel := make(chan pars.Document, chunk)
					errorsChannel := make(chan error)
					doneChannel := make(chan bool)

					chunkFileKeys := fileKeys[i:end]
					for _, chunkFileKey := range chunkFileKeys {
						waitGroup.Add(1)
						if !hasSuffix(chunkFileKey) {
							waitGroup.Done()
							continue
						}

						go func(bucket, chunkFileKey string) {
							defer waitGroup.Done()

							select {
							case <-errorsChannel:
								return
							case <-ctx.Done():
								return
							default:
							}

							document, err := parsClient.Parse(ctx, bucket, chunkFileKey)
							if err != nil {
								errorsChannel <- err
								return
							}
							documentsChannel <- *document

						}(bucket, chunkFileKey)
					}

					go func() {
						waitGroup.Wait()
						close(doneChannel)
					}()

					select {
					case <-doneChannel:
						close(documentsChannel)
					case err := <-errorsChannel:
						close(errorsChannel)
						util.Log("UPSERT_DOCUMENTS_ERROR", err.Error())
						return sendResponse(
							http.StatusInternalServerError,
							fmt.Sprintf(`{"error": %q}`, err),
						)
					}

					documents := []pars.Document{}
					for document := range documentsChannel {
						documents = append(documents, document)
					}

					if err := dbClient.UpsertDocuments(ctx, documents); err != nil {
						util.Log("UPSERT_DOCUMENTS_ERROR", err.Error())
						return sendResponse(
							http.StatusInternalServerError,
							fmt.Sprintf(`{"error": %q}`, err),
						)
					}
				}
			}
		}

		if requestJSON.Remove != nil {
			if err := evtClient.RemoveBucketListeners(ctx, requestJSON.Remove); err != nil {
				util.Log("REMOVE_BUCKET_LISTENERS_ERROR", err.Error())
				return sendResponse(
					http.StatusInternalServerError,
					fmt.Sprintf(`{"error": %q}`, err),
				)
			}

			if err := dbClient.DeleteDocumentsByBuckets(ctx, requestJSON.Remove); err != nil {
				util.Log("DELETE_DOCUMENTS_BY_BUCKETS_ERROR", err.Error())
				return sendResponse(
					http.StatusInternalServerError,
					fmt.Sprintf(`{"error": %q}`, err),
				)
			}
		}

		outputBody := fmt.Sprintf(`{"message": "success", "buckets_added": %d, "buckets_removed": %d}`, len(requestJSON.Add), len(requestJSON.Remove))
		util.Log("RESPONSE_BODY", outputBody)
		return sendResponse(http.StatusOK, outputBody)
	}
}

func hasSuffix(fileKey string) bool {
	suffixes := []string{"png", "jpg", "jpeg", "tiff", "pdf"}

	for _, suffix := range suffixes {
		if strings.HasSuffix(fileKey, suffix) {
			return true
		}
	}

	return false
}

func sendResponse(statusCode int, body string) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode:      statusCode,
		Body:            body,
		IsBase64Encoded: false,
	}, nil
}
