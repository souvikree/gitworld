//go:build integration

package store_test

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go/modules/neo4j"
	"go.uber.org/zap"

	"github.com/souvikree/gitworld/analyzer/internal/graph"
	"github.com/souvikree/gitworld/analyzer/internal/store"
)

func TestIngestGraph_Idempotent(t *testing.T) {
	ctx := context.Background()

	container, err := neo4j.RunContainer(ctx,
		neo4j.WithAdminPassword("testpassword"),
	)
	if err != nil {
		t.Fatalf("failed to start neo4j container: %v", err)
	}
	defer container.Terminate(ctx)

	uri, err := container.BoltUrl(ctx)
	if err != nil {
		t.Fatalf("failed to get bolt url: %v", err)
	}

	log := zap.NewNop()
	s, err := store.NewNeo4jStore(ctx, uri, "neo4j", "testpassword", log)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer s.Close(ctx)

	g := &graph.Graph{
		Nodes: []graph.Node{{ID: "f1", Type: graph.NodeFile, Path: "a.js", Name: "a.js"}},
	}

	// Run twice — idempotency check
	if err := s.IngestGraph(ctx, g); err != nil {
		t.Fatalf("first ingest failed: %v", err)
	}
	if err := s.IngestGraph(ctx, g); err != nil {
		t.Fatalf("second ingest failed: %v", err)
	}

	count, err := s.CountNodes(ctx, "File")
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 File node after double ingest, got %d", count)
	}
}