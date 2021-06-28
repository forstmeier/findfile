//+build !test

package main

import (
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/cheesesteakio/api/pkg/acct"
	"github.com/cheesesteakio/api/pkg/fs"
)

func main() {
	newSession := session.New()

	acctClient := acct.New(newSession)

	fsClient, err := fs.New(newSession)
	if err != nil {
		log.Fatalf("error creating fs client: %s", err.Error())
	}

	lambda.Start(handler(acctClient, fsClient))
}
