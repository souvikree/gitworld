package resolve_test

import (
	"context"
	"testing"

	"github.com/souvikree/gitworld/analyzer/internal/extract"
	"github.com/souvikree/gitworld/analyzer/internal/graph"
	"github.com/souvikree/gitworld/analyzer/internal/parser"
	"github.com/souvikree/gitworld/analyzer/internal/resolve"
	"github.com/souvikree/gitworld/analyzer/internal/walker"
	"go.uber.org/zap"
)

func TestResolve_RelativeImportMatchesRealNode(t *testing.T) {
	log := zap.NewNop()
	res, err := walker.Walk("../../testdata/fixtures/simple-import", log)
	if err != nil {
		t.Fatalf("walk failed: %v", err)
	}

	full := graph.Graph{}
	for _, f := range res.Files {
		ast, _ := parser.ParseFile(context.Background(), f)
		g, err := extract.FromAST(ast)
		if err != nil {
			t.Fatalf("extract failed: %v", err)
		}
		full.Nodes = append(full.Nodes, g.Nodes...)
		full.Edges = append(full.Edges, g.Edges...)
	}

	resolve.Resolve(&full)

	if len(full.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(full.Edges))
	}

	// find b.js's node ID
	var bID string
	for _, n := range full.Nodes {
		if n.Path[len(n.Path)-4:] == "b.js" {
			bID = n.ID
		}
	}
	if bID == "" {
		t.Fatal("b.js node not found")
	}
	if full.Edges[0].To != bID {
		t.Errorf("expected edge.To=%s (b.js node ID), got %s", bID, full.Edges[0].To)
	}
}