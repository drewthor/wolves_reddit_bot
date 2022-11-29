package nba

import "context"

type OutputWriter interface {
	Put(ctx context.Context, b []byte) error
}
