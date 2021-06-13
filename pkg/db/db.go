package db

import (
	"context"

	"github.com/cheesesteakio/api/pkg/docpars"
)

// Databaser defines the methods for interacting with the parsed
// documents in the database.
type Databaser interface {
	Create(ctx context.Context, documents []docpars.Document) error
	Update(ctx context.Context, documents []docpars.Document) error
	Delete(ctx context.Context, documentsInfo []DocumentInfo) error
	Query(ctx context.Context, query []byte) ([]docpars.Document, error)
}

// DocumentInfo holds data related to a document.
type DocumentInfo struct {
	Filename string
	Filepath string
}
