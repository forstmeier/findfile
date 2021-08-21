//+build !test

package main

import (
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/cheesesteakio/api/pkg/acct"
	"github.com/cheesesteakio/api/pkg/db"
)

func main() {
	newSession := session.New()

	acctClient := acct.New(newSession, os.Getenv("TABLE_NAME"))

	partitionerClient := db.NewPartitionerClient(
		newSession,
		os.Getenv("STORAGE_BUCKET"),
		os.Getenv("METADATA_TABLE_NAME"),
		os.Getenv("DATABASE_NAME"),
		os.Getenv("CATALOG_ID"),
	)

	lambda.Start(handler(acctClient, partitionerClient))
}
