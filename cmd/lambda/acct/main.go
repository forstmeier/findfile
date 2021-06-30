//+build !test

package main

import (
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/cheesesteakio/api/pkg/acct"
	"github.com/cheesesteakio/api/pkg/fs"
	"github.com/cheesesteakio/api/pkg/subscr"
)

func main() {
	newSession := session.New()

	acctClient := acct.New(newSession)

	fsClient := fs.New(newSession)

	stripeAPIKey := os.Getenv("STRIPE_API_KEY")

	stripeItemIDs := []string{}

	subscrClient := subscr.New(stripeAPIKey, stripeItemIDs)

	lambda.Start(handler(acctClient, subscrClient, fsClient))
}
