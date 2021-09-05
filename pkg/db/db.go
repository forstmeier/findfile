package db

import (
	"context"

	"github.com/cheesesteakio/api/pkg/pars"
)

// Databaser defines the methods for interacting with the parsed
// documents in the database.
type Databaser interface {
	UpsertDocuments(ctx context.Context, documents []pars.Document) error
	DeleteDocuments(ctx context.Context, documentsInfo []DocumentInfo) error
	QueryDocuments(ctx context.Context, query []byte) ([]pars.Document, error)
}

// DocumentInfo holds data related to a document.
type DocumentInfo struct {
	AccountID string
	Filename  string
	Filepath  string
}
