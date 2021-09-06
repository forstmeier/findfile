//+build !test

package main

import (
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/cheesesteakio/api/pkg/acct"
	"github.com/cheesesteakio/api/pkg/cql"
	"github.com/cheesesteakio/api/pkg/db"
)

func main() {
	newSession := session.New()

	acctClient := acct.New(newSession, os.Getenv("TABLE_NAME"))

	cqlClient := cql.New()

	dbClient := db.New(
		newSession,
		os.Getenv("DATABASE_NAME"),
		os.Getenv("STORAGE_BUCKET"),
	)

	lambda.Start(handler(acctClient, cqlClient, dbClient))
}
