//+build !test

package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/cheesesteakio/api/pkg/acct"
	"github.com/cheesesteakio/api/pkg/fs"
)

func main() {
	newSession := session.New()

	acctClient := acct.New(newSession)

	fsClient := fs.New(newSession)

	lambda.Start(handler(acctClient, fsClient))
}
