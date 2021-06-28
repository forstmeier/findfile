//+build !test

package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/cheesesteakio/api/pkg/acct"
	"github.com/cheesesteakio/api/pkg/subscr"
)

func main() {
	acctClient := acct.New(session.New())

	subscrClient := subscr.New()

	lambda.Start(handler(acctClient, subscrClient))
}
