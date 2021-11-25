//+build !test

package main

import (
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/forstmeier/findfile/pkg/evt"
)

func main() {
	newSession := session.New()

	evtClient := evt.New(
		newSession,
		os.Getenv("TRAIL_NAME"),
	)

	httpSecurityHeader := os.Getenv("HTTP_SECURITY_HEADER")
	httpSecurityKey := os.Getenv("HTTP_SECURITY_KEY")

	lambda.Start(handler(evtClient, httpSecurityHeader, httpSecurityKey))
}
