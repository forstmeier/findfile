//+build !test

package main

import (
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/cheesesteakio/api/pkg/db"
	"github.com/cheesesteakio/api/pkg/docpars"
)

func main() {
	connectionURI := db.GetConnectionURI(
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_ENDPOINT"),
	)
	tlsConfig, err := db.GetTLSConfig("rds-combined-ca-bundle.pem")
	if err != nil {
		log.Fatalf("error parsing tls config: %s", err.Error())
	}
	ddb, err := mongo.NewClient(options.Client().ApplyURI(connectionURI).SetTLSConfig(tlsConfig))
	if err != nil {
		log.Fatalf("error creating mongo db client: %s", err.Error())
	}

	docparsClient := docpars.New(session.New())

	dbClient := db.New(ddb, "main", "documents")

	lambda.Start(handler(docparsClient, dbClient))
}
