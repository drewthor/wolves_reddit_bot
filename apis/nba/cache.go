package nba

import (
	"context"
	"io"
)

// ObjectCacher defines the interface to implement to use for caching nba calls.
type ObjectCacher interface {
	// GetObject gets the object given the key from the cache; returns ErrNotFound if the object is not found in the cache
	GetObject(ctx context.Context, key string) ([]byte, error)
	PutObject(ctx context.Context, key string, obj io.Reader) error
}
