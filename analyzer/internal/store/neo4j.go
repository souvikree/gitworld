package store

import (
	"context"
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"

	"github.com/souvikree/gitworld/analyzer/internal/graph"
)

type Neo4jStore struct {
	driver neo4j.DriverWithContext
	log    *zap.Logger
}

// NewNeo4jStore opens a driver and verifies connectivity immediately —
// fail loud at startup, not on the first real query.
func NewNeo4jStore(ctx context.Context, uri, user, password string, log *zap.Logger) (*Neo4jStore, error) {
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(user, password, ""))
	if err != nil {
		return nil, fmt.Errorf("store: failed to create driver: %w", err)
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := driver.VerifyConnectivity(ctxTimeout); err != nil {
		driver.Close(ctx)
		return nil, fmt.Errorf("store: cannot reach Neo4j at %s: %w", uri, err)
	}

	return &Neo4jStore{driver: driver, log: log}, nil
}

// IngestGraph upserts every node and edge. Idempotent — running the same
// graph twice produces the same end state, not duplicates.
func (s *Neo4jStore) IngestGraph(ctx context.Context, g *graph.Graph) error {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		for _, n := range g.Nodes {
			label := string(n.Type)
			query := fmt.Sprintf(
				"MERGE (x:%s {id: $id}) SET x.path = $path, x.name = $name, x.repoId = $repoId", label,
			)
			if _, err := tx.Run(ctx, query, map[string]any{
				"id":     n.ID,
				"path":   n.Path,
				"name":   n.Name,
				"repoId": n.RepoID,
			}); err != nil {
				return nil, fmt.Errorf("upsert node %s: %w", n.ID, err)
			}
		}

		for _, e := range g.Edges {
			query := fmt.Sprintf(
				`MATCH (a {id: $from}), (b {id: $to})
				 MERGE (a)-[:%s]->(b)`, string(e.Type),
			)
			if _, err := tx.Run(ctx, query, map[string]any{
				"from": e.From,
				"to":   e.To,
			}); err != nil {
				return nil, fmt.Errorf("upsert edge %s->%s: %w", e.From, e.To, err)
			}
		}
		return nil, nil
	})

	if err != nil {
		s.log.Error("ingestion failed", zap.Error(err))
		return fmt.Errorf("store: ingest failed: %w", err)
	}

	s.log.Info("ingestion complete", zap.Int("nodes", len(g.Nodes)), zap.Int("edges", len(g.Edges)))
	return nil
}

// CountNodes is a test/debug helper — returns total node count with a given label.
func (s *Neo4jStore) CountNodes(ctx context.Context, label string) (int64, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(ctx, fmt.Sprintf("MATCH (n:%s) RETURN count(n) AS c", label), nil)
		if err != nil {
			return nil, err
		}
		record, err := res.Single(ctx)
		if err != nil {
			return nil, err
		}
		return record.Values[0].(int64), nil
	})
	if err != nil {
		return 0, fmt.Errorf("count nodes: %w", err)
	}
	return result.(int64), nil
}

// EnsureSchema creates required constraints if they don't already exist.
// Safe to call on every startup — idempotent by design (IF NOT EXISTS).
func (s *Neo4jStore) EnsureSchema(ctx context.Context) error {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	constraints := []string{
		"CREATE CONSTRAINT file_id_unique IF NOT EXISTS FOR (f:File) REQUIRE f.id IS UNIQUE",
		"CREATE CONSTRAINT module_id_unique IF NOT EXISTS FOR (m:Module) REQUIRE m.id IS UNIQUE",
	}

	for _, c := range constraints {
		if _, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			return tx.Run(ctx, c, nil)
		}); err != nil {
			return fmt.Errorf("store: failed to apply constraint: %w", err)
		}
	}

	s.log.Info("schema constraints ensured")
	return nil
}

func (s *Neo4jStore) Close(ctx context.Context) error {
	return s.driver.Close(ctx)
}
