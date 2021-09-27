//+build !test

package main

import (
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/forstmeier/findfile/pkg/db"
	"github.com/forstmeier/findfile/pkg/fql"
)

func main() {
	newSession := session.New()

	fqlClient := fql.New()

	dbClient := db.New(
		newSession,
		os.Getenv("DATABASE_NAME"),
		os.Getenv("DATABASE_BUCKET"),
	)

	httpSecurityHeader := os.Getenv("HTTP_SECURITY_HEADER")
	httpSecurityKey := os.Getenv("HTTP_SECURITY_KEY")

	lambda.Start(handler(fqlClient, dbClient, httpSecurityHeader, httpSecurityKey))
}
