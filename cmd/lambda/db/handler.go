package main

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/aws/aws-lambda-go/events"

	"github.com/cheesesteakio/api/pkg/acct"
	"github.com/cheesesteakio/api/pkg/db"
	"github.com/cheesesteakio/api/pkg/docpars"
	"github.com/cheesesteakio/api/util"
)

var (
	errorUnsupportedEvent        = errors.New("event type not supported")
	errorUnmarshalEvent          = errors.New("event json unmarshal error")
	errorGetAccount              = errors.New("get account error")
	errorParseFile               = errors.New("parse file error")
	errorCreateOrUpdateDocuments = errors.New("upsert documents error")
	errorDeleteDocuments         = errors.New("delete documents error")
)

func handler(acctClient acct.Accounter, docparsClient docpars.Parser, dbClient db.Databaser) func(ctx context.Context, event events.SNSEvent) error {
	return func(ctx context.Context, event events.SNSEvent) error {
		util.Log("EVENT_BODY", event)

		createOrUpdateDocs := [][3]string{}
		deleteDocs := []db.DocumentInfo{}

		for _, snsRecord := range event.Records {
			s3Event := events.S3Event{}

			if err := json.Unmarshal([]byte(snsRecord.SNS.Message), &s3Event); err != nil {
				return errorUnmarshalEvent
			}

			for _, s3Record := range s3Event.Records {
				account, err := acctClient.GetAccountBySecondaryID(ctx, s3Record.S3.Bucket.Name)
				if err != nil {
					return errorGetAccount
				}

				if account == nil {
					continue
				}

				if s3Record.EventName == "ObjectCreated:Put" {
					createOrUpdateDocs = append(createOrUpdateDocs, [3]string{
						account.ID,
						s3Record.S3.Object.Key,
						s3Record.S3.Bucket.Name,
					})

				} else if s3Record.EventName == "ObjectRemoved:Delete" {
					deleteDocs = append(deleteDocs, db.DocumentInfo{
						Filepath: s3Record.S3.Bucket.Name,
						Filename: s3Record.S3.Object.Key,
					})

				} else {
					return errorUnsupportedEvent
				}
			}
		}

		documents := make([]docpars.Document, len(createOrUpdateDocs))
		for i, doc := range createOrUpdateDocs {
			document, err := docparsClient.Parse(ctx, doc[0], doc[1], doc[2], nil)
			if err != nil {
				util.Log("PARSE_ERROR", err)
				return errorParseFile
			}

			documents[i] = *document
		}

		if err := dbClient.UpsertDocuments(ctx, documents); err != nil {
			util.Log("UPSERT_DOCUMENTS_ERROR", err)
			return errorCreateOrUpdateDocuments
		}

		if err := dbClient.DeleteDocuments(ctx, deleteDocs); err != nil {
			util.Log("DELETE_DOCUMENTS_ERROR", err)
			return errorDeleteDocuments
		}

		util.Log("RESPONSE_BODY", "successful invocation")
		return nil
	}
}
