package main

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-lambda-go/events"

	"github.com/cheesesteakio/api/pkg/db"
	"github.com/cheesesteakio/api/pkg/docpars"
)

var (
	errorUnsupportedEvent        = errors.New("event type not supported")
	errorParseFile               = errors.New("parse file error")
	errorCreateOrUpdateDocuments = errors.New("create or update documents error")
	errorDeleteDocuments         = errors.New("delete documents error")
)

func handler(docparsClient docpars.Parser, dbClient db.Databaser) func(ctx context.Context, event events.S3Event) error {
	return func(ctx context.Context, event events.S3Event) error {
		createOrUpdateDocs := [][3]string{}
		deleteDocs := []db.DocumentInfo{}

		for _, record := range event.Records {
			if record.EventName == "s3:ObjectCreated:Put" {
				keyElements := strings.Split(record.S3.Object.Key, "/")
				accountID := keyElements[len(keyElements)-2]

				createOrUpdateDocs = append(createOrUpdateDocs, [3]string{
					accountID,
					record.S3.Object.Key,
					record.S3.Bucket.Name,
				})

			} else if record.EventName == "s3:ObjectRemoved:Delete" {
				deleteDocs = append(deleteDocs, db.DocumentInfo{
					Filepath: record.S3.Bucket.Name,
					Filename: record.S3.Object.Key,
				})

			} else {
				return errorUnsupportedEvent
			}
		}

		documents := make([]docpars.Document, len(createOrUpdateDocs))
		for i, doc := range createOrUpdateDocs {
			document, err := docparsClient.Parse(ctx, doc[0], doc[1], doc[2], nil)
			if err != nil {
				return errorParseFile
			}

			documents[i] = *document
		}

		if err := dbClient.CreateOrUpdateDocuments(ctx, documents); err != nil {
			return errorCreateOrUpdateDocuments
		}

		if err := dbClient.DeleteDocuments(ctx, deleteDocs); err != nil {
			return errorDeleteDocuments
		}

		return nil
	}
}
