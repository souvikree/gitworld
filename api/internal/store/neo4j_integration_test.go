//go:build integration

package store_test

import (
	"context"
	"testing"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	tcneo4j "github.com/testcontainers/testcontainers-go/modules/neo4j"
	"go.uber.org/zap"

	"github.com/souvikree/gitworld/api/internal/store"
)

func TestGetGraph_ReturnsSeededData(t *testing.T) {
	ctx := context.Background()

	container, err := tcneo4j.RunContainer(ctx,
		tcneo4j.WithAdminPassword("testpassword"),
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

	if err := seedTestData(ctx, uri); err != nil {
		t.Fatalf("seed failed: %v", err)
	}

	g, err := s.GetGraph(ctx, "test-repo", 100)
	if err != nil {
		t.Fatalf("GetGraph failed: %v", err)
	}

	if len(g.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(g.Nodes))
	}
	if len(g.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(g.Edges))
	}
	if g.Edges[0].From != "a" || g.Edges[0].To != "b" {
		t.Errorf("expected edge a->b, got %s->%s", g.Edges[0].From, g.Edges[0].To)
	}
}

func seedTestData(ctx context.Context, uri string) error {
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth("neo4j", "testpassword", ""))
	if err != nil {
		return err
	}
	defer driver.Close(ctx)

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err = session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		return tx.Run(ctx, `
			MERGE (a:File {id: 'a', name: 'a.js'})
			MERGE (b:File {id: 'b', name: 'b.js'})
			MERGE (a)-[:IMPORTS]->(b)
		`, nil)
	})
	return err
}