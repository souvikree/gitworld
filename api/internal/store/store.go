package store

import "context"

type GraphReader interface {
	GetGraph(ctx context.Context, limit int) (*Graph, error)
}