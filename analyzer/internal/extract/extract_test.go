package extract_test

import (
	"context"
	"testing"

	"github.com/souvikree/gitworld/analyzer/internal/extract"
	"github.com/souvikree/gitworld/analyzer/internal/graph"
	"github.com/souvikree/gitworld/analyzer/internal/parser"
)

func TestFromAST_SimpleImport(t *testing.T) {
	ast, err := parser.ParseFile(context.Background(), "../../testdata/fixtures/simple-import/a.js")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	g, err := extract.FromAST(ast, "test-repo")
	if err != nil {
		t.Fatalf("extract failed: %v", err)
	}

	if len(g.Nodes) != 1 {
		t.Errorf("expected 1 file node, got %d", len(g.Nodes))
	}
	if len(g.Edges) != 1 {
		t.Fatalf("expected 1 import edge, got %d", len(g.Edges))
	}
	if g.Edges[0].Type != graph.EdgeImports {
		t.Errorf("expected IMPORTS edge, got %s", g.Edges[0].Type)
	}
	if g.Edges[0].To != "./b.js" {
		t.Errorf("expected edge to './b.js', got %s", g.Edges[0].To)
	}
}