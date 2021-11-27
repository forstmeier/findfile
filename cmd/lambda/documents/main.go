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
	)
	if err != nil {
		panic(fmt.Sprintf("error creating db client: %v", err))
	}

	httpSecurityHeader := os.Getenv("HTTP_SECURITY_HEADER")
	httpSecurityKey := os.Getenv("HTTP_SECURITY_KEY")

	lambda.Start(handler(dbClient, httpSecurityHeader, httpSecurityKey))
}
