//+build !test

package main

import (
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/cheesesteakio/api/pkg/acct"
	"github.com/cheesesteakio/api/pkg/fs"
)

func main() {
	newSession := session.New()

	tableName := os.Getenv("TABLE_NAME")

	acctClient := acct.New(newSession, tableName)

	fsClient := fs.New(newSession)

	lambda.Start(handler(acctClient, fsClient))
}
