package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aws/aws-lambda-go/events"

	"github.com/findfiledev/api/pkg/db"
	"github.com/findfiledev/api/pkg/pars"
	"github.com/findfiledev/api/util"
)

var (
	errorUnsupportedEvent = errors.New("event type not supported")
	errorParseFile        = errors.New("parse file error")
	errorQueryDocuments   = errors.New("query documents error")
	errorUpsertDocuments  = errors.New("upsert documents error")
	errorDeleteDocuments  = errors.New("delete documents error")
)

type fileInfo struct {
	key    string
	bucket string
}

const deleteDocumentsQuery = `
select
	documents.id as document_id,
	pages.id as page_id,
	lines.id as line_id,
	coordinates.id as coordinates_id
from coordinates
inner join lines
on coordinates.line_id = lines.id
inner join pages
on lines.page_id = pages.id
inner join (
	select id, file_key, file_bucket
	from documents
	where file_key in (%s)
	and file_bucket = '%s'
) as documents
on pages.document_id = documents.id;
`

func handler(docparsClient pars.Parser, dbClient db.Databaser) func(ctx context.Context, event events.S3Event) error {
	return func(ctx context.Context, event events.S3Event) error {
		util.Log("EVENT_BODY", event)

		upsertFiles := []fileInfo{}
		deleteFiles := map[string][]string{}

		for _, s3Record := range event.Records {
			extension := filepath.Ext(s3Record.S3.Object.Key)
			if extension != ".jpg" && extension != ".jpeg" && extension != ".png" {
				continue
			}

			if s3Record.EventName == "ObjectCreated:Put" {
				upsertFiles = append(upsertFiles, fileInfo{
					key:    s3Record.S3.Object.Key,
					bucket: s3Record.S3.Bucket.Name,
				})
			} else if s3Record.EventName == "ObjectRemoved:Delete" {
				if _, ok := deleteFiles[s3Record.S3.Bucket.Name]; ok {
					deleteFiles[s3Record.S3.Bucket.Name] = append(
						deleteFiles[s3Record.S3.Bucket.Name],
						s3Record.S3.Object.Key,
					)
				} else {
					deleteFiles[s3Record.S3.Bucket.Name] = []string{s3Record.S3.Object.Key}
				}
			} else {
				util.Log("UNSUPPORTED_EVENT", fmt.Sprintf("event [%s] not supported", s3Record.EventName))
				return errorUnsupportedEvent
			}
		}

		upsertDocuments := make([]pars.Document, len(upsertFiles))
		for i, file := range upsertFiles {
			document, err := docparsClient.Parse(ctx, file.key, file.bucket)
			if err != nil {
				util.Log("PARSE_ERROR", err)
				return errorParseFile
			}

			upsertDocuments[i] = *document
		}

		deleteDocumentKeys := []string{}
		for fileBucket, fileKeys := range deleteFiles {
			fileKeysString := "'" + strings.Join(fileKeys, "','") + "'"

			query := fmt.Sprintf(
				deleteDocumentsQuery,
				fileKeysString,
				fileBucket,
			)

			queryDeleteDocumentKeys, err := dbClient.QueryDocumentKeysByFileInfo(ctx, []byte(query))
			if err != nil {
				util.Log("QUERY_DOCUMENTS", err)
				return errorQueryDocuments
			}

			deleteDocumentKeys = append(deleteDocumentKeys, queryDeleteDocumentKeys...)
		}

		if err := dbClient.UpsertDocuments(ctx, upsertDocuments); err != nil {
			util.Log("UPSERT_DOCUMENTS_ERROR", err)
			return errorUpsertDocuments
		}

		if err := dbClient.DeleteDocuments(ctx, deleteDocumentKeys); err != nil {
			util.Log("DELETE_DOCUMENTS_ERROR", err)
			return errorDeleteDocuments
		}

		util.Log("RESPONSE_BODY", "successful invocation")
		return nil
	}
}
