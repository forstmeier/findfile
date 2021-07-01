//+build !test

package main

import (
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/cheesesteakio/api/pkg/acct"
	"github.com/cheesesteakio/api/pkg/csql"
	"github.com/cheesesteakio/api/pkg/db"
	"github.com/cheesesteakio/api/pkg/fs"
)

func main() {
	newSession := session.New()

	ddb, err := mongo.NewClient(nil)
	if err != nil {
		log.Fatalf("error creating mongo db client: %s", err.Error())
	}

	tableName := os.Getenv("TABLE_NAME")

	acctClient := acct.New(newSession, tableName)

	csqlClient := csql.New()

	dbClient := db.New(ddb, "main", "documents")

	fsClient := fs.New(newSession)

	lambda.Start(handler(acctClient, csqlClient, dbClient, fsClient))
}
