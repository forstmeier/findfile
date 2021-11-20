package db

import (
	"context"

	"github.com/forstmeier/findfile/pkg/pars"
)

// Databaser defines the methods for interacting with the parsed
// documents in the database.
type Databaser interface {
	SetupDatabase(ctx context.Context) error
	UpsertDocuments(ctx context.Context, documents []pars.Document) error
	DeleteDocuments(ctx context.Context, documentIDs []string) error
	QueryDocuments(ctx context.Context, query []byte) ([]pars.Document, error)
}
