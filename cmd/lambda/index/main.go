//+build !test

package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/forstmeier/findfile/pkg/db"
)

func main() {
	newSession := session.New()

	dbClient, err := db.New(
		newSession,
		os.Getenv("DATABASE_URL"),
		os.Getenv("DATABASE_USERNAME"),
		os.Getenv("DATABASE_PASSWORD"),
	)
	if err != nil {
		panic(fmt.Sprintf("error creating db client: %v", err))
	}

	lambda.Start(handler(dbClient, sendResponse))
}
