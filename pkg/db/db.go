package db

import (
	"context"

	"github.com/cheesesteakio/api/pkg/docpars"
)

// Databaser defines the methods for interacting with the parsed
// documents in the database.
type Databaser interface {
	CreateDocuments(ctx context.Context, documents []docpars.Document) error
	UpdateDocuments(ctx context.Context, documents []docpars.Document) error
	DeleteDocuments(ctx context.Context, documentsInfo []DocumentInfo) error
	QueryDocuments(ctx context.Context, query []byte) ([]docpars.Document, error)
}

// DocumentInfo holds data related to a document.
type DocumentInfo struct {
	Filename string
	Filepath string
}
