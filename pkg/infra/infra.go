package infra

import "context"

// Infrastructurer defines the methods needed for working with
// resources required in the application.
type Infrastructurer interface {
	CreateFilesystem(ctx context.Context, accountID string) error
	DeleteFilesystem(ctx context.Context, accountID string) error
	CreateDatabase(ctx context.Context, accountID string) error
	DeleteDatabase(ctx context.Context, accountID string) error
}
