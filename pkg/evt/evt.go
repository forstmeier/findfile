package evt

import "context"

// Eventer defines the methods for manipulating the event listener
// resources for bucket data sources.
type Eventer interface {
	AddBucketListeners(ctx context.Context, buckets []string) error
	RemoveBucketListeners(ctx context.Context, buckets []string) error
}
