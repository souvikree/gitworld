package store

import "context"

type GraphReader interface {
	GetGraph(ctx context.Context, repoID string, limit int) (*Graph, error)
}