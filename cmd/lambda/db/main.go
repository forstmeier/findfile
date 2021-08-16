//+build !test

package main

import (
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/cheesesteakio/api/pkg/acct"
	"github.com/cheesesteakio/api/pkg/db"
	"github.com/cheesesteakio/api/pkg/docpars"
)

func main() {
	newSession := session.New()

	acctClient := acct.New(newSession, os.Getenv("TABLE_NAME"))

	docparsClient := docpars.New(newSession)

	dbClient := db.New(newSession, os.Getenv("DATABASE_NAME"), os.Getenv("STORAGE_BUCKET"))

	lambda.Start(handler(acctClient, docparsClient, dbClient))
}
