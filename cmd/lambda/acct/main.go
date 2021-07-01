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

	acctClient := acct.New(newSession, os.Getenv("TABLE_NAME"))

	fsClient := fs.New(newSession)

	stripeItemIDs := []string{
		os.Getenv("STRIPE_MONTHLY_PRICE_ID"),
		os.Getenv("STRIPE_METERED_PRICE_ID"),
	}

	subscrClient := subscr.New(os.Getenv("STRIPE_API_KEY"), stripeItemIDs)

	lambda.Start(handler(acctClient, subscrClient, fsClient))
}
