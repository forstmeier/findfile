//+build !test

package main

import (
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/cheesesteakio/api/pkg/acct"
)

func main() {
	newSession := session.New()

	acctClient := acct.New(newSession, os.Getenv("TABLE_NAME"))

	lambda.Start(handler(acctClient))
}
