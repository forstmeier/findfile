package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"

	"github.com/forstmeier/findfile/pkg/db"
	"github.com/forstmeier/findfile/pkg/pars"
	"github.com/forstmeier/findfile/util"
)

type detailsPayload struct {
	EventName         string            `json:"eventName"`
	RequestParameters requestParameters `json:"requestParameters"`
}

type requestParameters struct {
	BucketName string `json:"bucketName"`
	Key        string `json:"key"`
}

func handler(parsClient pars.Parser, dbClient db.Databaser) func(ctx context.Context, event events.CloudWatchEvent) error {
	return func(ctx context.Context, event events.CloudWatchEvent) error {
		detailsJSON := detailsPayload{}
		if err := json.Unmarshal(event.Detail, &detailsJSON); err != nil {
			util.Log("UNMARSHAL_REQUEST_PAYLOAD_ERROR", err.Error())
			return err
		}

		if detailsJSON.EventName == "PutObject" {
			document, err := parsClient.Parse(ctx, detailsJSON.RequestParameters.BucketName, detailsJSON.RequestParameters.Key)
			if err != nil {
				util.Log("PARSE_ERROR", err.Error())
				return err
			}

			if err := dbClient.UpsertDocuments(ctx, []pars.Document{*document}); err != nil {
				util.Log("UPSERT_DOCUMENTS_ERROR", err.Error())
				return err
			}

		} else if detailsJSON.EventName == "DeleteObjects" {
			if err := dbClient.DeleteDocumentsByIDs(ctx, []string{detailsJSON.RequestParameters.Key}); err != nil {
				util.Log("DELETE_DOCUMENTS_ERROR", err.Error())
				return err
			}
		}

		return nil
	}
}
