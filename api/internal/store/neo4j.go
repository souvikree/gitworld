package store

import (
	"context"
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

type Neo4jStore struct {
	driver neo4j.DriverWithContext
	log    *zap.Logger
}

type Node struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Path string `json:"path,omitempty"`
	Name string `json:"name"`
}

type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
}

type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

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

// GetGraph fetches the full graph. Bounded by limit to avoid an
// accidental full-database dump on a large real-world repo.
func (s *Neo4jStore) GetGraph(ctx context.Context, repoID string, limit int) (*Graph, error) {
	if repoID == "" {
		return nil, fmt.Errorf("store: repoID is required")
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	session := s.driver.NewSession(ctxTimeout, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctxTimeout)

	result, err := session.ExecuteRead(ctxTimeout, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(ctxTimeout,
			`MATCH (n {repoId: $repoID})
			 OPTIONAL MATCH (n)-[r]->(m {repoId: $repoID})
			 RETURN n, r, m, n.id AS fromID, m.id AS toID
			 LIMIT $limit`,
			map[string]any{"repoID": repoID, "limit": limit},
		)
		if err != nil {
			return nil, err
		}

		g := &Graph{Nodes: []Node{}, Edges: []Edge{}}
		seen := make(map[string]bool)

		for res.Next(ctxTimeout) {
			record := res.Record()

			if n, ok := record.Get("n"); ok && n != nil {
				node := n.(neo4j.Node)
				addNodeIfNew(g, node, seen)
			}
			if m, ok := record.Get("m"); ok && m != nil {
				node := m.(neo4j.Node)
				addNodeIfNew(g, node, seen)
			}
			if r, ok := record.Get("r"); ok && r != nil {
				fromID, _ := record.Get("fromID")
				toID, _ := record.Get("toID")
				fromIDStr, ok1 := fromID.(string)
				toIDStr, ok2 := toID.(string)
				if !ok1 || !ok2 {
					// Shouldn't happen if data is well-formed, but never
					// silently write a malformed edge — skip and let it
					// be absent rather than corrupt.
					continue
				}
				rel := r.(neo4j.Relationship)
				g.Edges = append(g.Edges, Edge{
					From: fromIDStr,
					To:   toIDStr,
					Type: rel.Type,
				})
			}
		}
		return g, res.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("store: get graph failed: %w", err)
	}

	return result.(*Graph), nil
}

func addNodeIfNew(g *Graph, n neo4j.Node, seen map[string]bool) {
	id, _ := n.Props["id"].(string)
	if id == "" || seen[id] {
		return
	}
	seen[id] = true

	label := ""
	if len(n.Labels) > 0 {
		label = n.Labels[0]
	}

	name, _ := n.Props["name"].(string)
	path, _ := n.Props["path"].(string)

	g.Nodes = append(g.Nodes, Node{ID: id, Type: label, Path: path, Name: name})
}

func (s *Neo4jStore) Close(ctx context.Context) error {
	return s.driver.Close(ctx)
}