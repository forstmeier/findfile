//+build !test

package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/forstmeier/findfile/pkg/db"
	"github.com/forstmeier/findfile/pkg/evt"
	"github.com/forstmeier/findfile/pkg/fs"
	"github.com/forstmeier/findfile/pkg/pars"
)

func main() {
	newSession := session.New()

	evtClient := evt.New(
		newSession,
		os.Getenv("TRAIL_NAME"),
	)

	fsClient := fs.New(
		newSession,
	)

	parsClient := pars.New(
		newSession,
	)

	dbClient, err := db.New(
		newSession,
		os.Getenv("DATABASE_URL"),
		os.Getenv("DATABASE_USERNAME"),
		os.Getenv("DATABASE_PASSWORD"),
	)
	if err != nil {
		panic(fmt.Sprintf("error creating db client: %v", err))
	}

	httpSecurityHeader := os.Getenv("HTTP_SECURITY_HEADER")
	httpSecurityKey := os.Getenv("HTTP_SECURITY_KEY")

	lambda.Start(handler(evtClient, fsClient, parsClient, dbClient, httpSecurityHeader, httpSecurityKey))
}
