//+build !test

package main

import (
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/cheesesteakio/api/pkg/acct"
	"github.com/cheesesteakio/api/pkg/subscr"
)

func main() {
	stripeAPIKey := os.Getenv("STRIPE_API_KEY")

	stripeItemIDs := []string{}

	acctClient := acct.New(session.New())

	subscrClient := subscr.New(stripeAPIKey, stripeItemIDs)

	lambda.Start(handler(acctClient, subscrClient))
}
