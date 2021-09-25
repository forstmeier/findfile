//+build !test

package main

import (
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/findfiledev/api/pkg/db"
)

func main() {
	newSession := session.New()

	dbClient := db.New(
		newSession,
		os.Getenv("DATABASE_NAME"),
		os.Getenv("DATABASE_BUCKET"),
	)

	lambda.Start(handler(dbClient, sendResponse))
}
