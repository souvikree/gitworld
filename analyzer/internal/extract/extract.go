package extract

import (
	"crypto/sha1"
	"encoding/hex"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/<you>/codeworld/analyzer/internal/graph"
	"github.com/<you>/codeworld/analyzer/internal/parser"
)

// FromAST walks a parsed file's tree and produces graph nodes/edges.
// This is intentionally minimal for the first pass: file node + import edges.
// Function/class extraction and call-graph edges are added incrementally
// once this base case has fixture coverage.
func FromAST(f *parser.FileAST) (graph.Graph, error) {
	fileID := nodeID(f.Path)
	g := graph.Graph{
		Nodes: []graph.Node{
			{ID: fileID, Type: graph.NodeFile, Path: f.Path, Name: f.Path},
		},
	}

	root := f.Tree.RootNode()
	walkImports(root, f.Source, func(importPath string) {
		g.Edges = append(g.Edges, graph.Edge{
			From: fileID,
			To:   importPath, // resolved to a real file path in a later pass
			Type: graph.EdgeImports,
		})
	})

	return g, nil
}

func walkImports(n *sitter.Node, src []byte, emit func(string)) {
	if n == nil {
		return
	}
	if n.Type() == "import_statement" {
		if src_ := n.ChildByFieldName("source"); src_ != nil {
			raw := src_.Content(src)
			emit(trimQuotes(raw))
		}
	}
	for i := 0; i < int(n.ChildCount()); i++ {
		walkImports(n.Child(i), src, emit)
	}
}

func trimQuotes(s string) string {
	if len(s) >= 2 {
		return s[1 : len(s)-1]
	}
	return s
}

func nodeID(path string) string {
	sum := sha1.Sum([]byte(path))
	return hex.EncodeToString(sum[:])
}