//+build !test

package main

import (
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/findfiledev/api/pkg/db"
	"github.com/findfiledev/api/pkg/pars"
)

func main() {
	newSession := session.New()

	parsClient := pars.New(newSession)

	dbClient := db.New(
		newSession,
		os.Getenv("DATABASE_NAME"),
		os.Getenv("DATABASE_BUCKET"),
	)

	lambda.Start(handler(parsClient, dbClient))
}
