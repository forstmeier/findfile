//+build !test

package main

import (
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/cheesesteakio/api/pkg/db"
	"github.com/cheesesteakio/api/pkg/docpars"
)

func main() {
	ddb, err := mongo.NewClient(nil)
	if err != nil {
		log.Fatalf("error creating mongo db client: %s", err.Error())
	}

	docparsClient := docpars.New(session.New())

	dbClient := db.New(ddb, "main", "documents")

	lambda.Start(handler(docparsClient, dbClient))
}
