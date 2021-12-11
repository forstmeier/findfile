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
			return util.SendResponse(
				http.StatusBadRequest,
				fmt.Errorf("security key header '%s' not provided", httpSecurityHeader),
				"SECURITY_KEY_HEADER_ERROR",
			)
		}

		if httpSecurityKeyReceived != httpSecurityKey {
			return util.SendResponse(
				http.StatusBadRequest,
				fmt.Errorf("security key '%s' incorrect", httpSecurityKeyReceived),
				"SECURITY_KEY_VALUE_ERROR",
			)
		}

		requestJSON := requestsPayload{}
		if err := json.Unmarshal([]byte(request.Body), &requestJSON); err != nil {
			return util.SendResponse(
				http.StatusBadRequest,
				err,
				"UNMARSHAL_REQUEST_PAYLOAD_ERROR",
			)
		}

		if requestJSON.Add != nil {
			if err := evtClient.AddBucketListeners(ctx, requestJSON.Add); err != nil {
				return util.SendResponse(
					http.StatusInternalServerError,
					err,
					"ADD_BUCKET_LISTENERS_ERROR",
				)
			}

			ctx, cancel := context.WithTimeout(ctx, 25*time.Second)
			defer cancel()

			for _, bucket := range requestJSON.Add {
				fileKeys, err := fsClient.ListFiles(ctx, bucket)
				if err != nil {
					return util.SendResponse(
						http.StatusInternalServerError,
						err,
						"LIST_FILES_ERROR",
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
						return util.SendResponse(
							http.StatusInternalServerError,
							err,
							"PARSE_DOCUMENTS_ERROR",
						)
					}

					documents := []pars.Document{}
					for document := range documentsChannel {
						documents = append(documents, document)
					}

					if err := dbClient.UpsertDocuments(ctx, documents); err != nil {
						return util.SendResponse(
							http.StatusInternalServerError,
							err,
							"UPSERT_DOCUMENTS_ERROR",
						)
					}
				}
			}
		}

		if requestJSON.Remove != nil {
			if err := evtClient.RemoveBucketListeners(ctx, requestJSON.Remove); err != nil {
				return util.SendResponse(
					http.StatusInternalServerError,
					err,
					"REMOVE_BUCKET_LISTENERS_ERROR",
				)
			}

			if err := dbClient.DeleteDocumentsByBuckets(ctx, requestJSON.Remove); err != nil {
				return util.SendResponse(
					http.StatusInternalServerError,
					err,
					"DELETE_DOCUMENTS_BY_BUCKETS_ERROR",
				)
			}
		}

		return util.SendResponse(
			http.StatusOK,
			map[string]int{
				"buckets_added":   len(requestJSON.Add),
				"buckets_removed": len(requestJSON.Remove),
			},
			"RESPONSE_BODY",
		)
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
