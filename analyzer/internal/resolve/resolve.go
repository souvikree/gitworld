package resolve

import (
	"path/filepath"
	"strings"

	"github.com/souvikree/gitworld/analyzer/internal/graph"
)

var resolvableExt = []string{".js", ".jsx", ".ts", ".tsx"}

// Resolve rewrites edge.To values in place: relative imports become the
// matching node's ID; bare specifiers (npm packages, node: builtins)
// become External nodes instead, so no edge ever points at a dangling string.
func Resolve(g *graph.Graph) {
	pathToID := make(map[string]string, len(g.Nodes))
	for _, n := range g.Nodes {
		pathToID[n.Path] = n.ID
	}

	externalSeen := make(map[string]string) // specifier -> node ID, dedupe

	for i, e := range g.Edges {
		if isRelative(e.To) {
			fromNode := findNode(g.Nodes, e.From)
			if fromNode == nil {
				continue // shouldn't happen, but never panic on bad data
			}
			resolved, ok := resolveRelative(fromNode.Path, e.To, pathToID)
			if ok {
				g.Edges[i].To = resolved
				continue
			}
			// Couldn't resolve to a real file (e.g. import of a .css, a
			// generated file, or something outside the walked set).
			// Leave it as-is but mark it, rather than silently pretending
			// it's fine — "fail loud, don't skip silently."
			g.Edges[i].To = e.To + " [unresolved]"
			continue
		}

		// Bare specifier: npm package or node: builtin — not a file we walked.
		id, exists := externalSeen[e.To]
		fromNode := findNode(g.Nodes, e.From)
		if !exists {
			id = "external:" + e.To
			externalSeen[e.To] = id
			g.Nodes = append(g.Nodes, graph.Node{
				ID:   id,
				Type: graph.NodeModule,
				Name: e.To,
				RepoID: fromNode.RepoID,
			})
		}
		g.Edges[i].To = id
	}
}

func isRelative(spec string) bool {
	return strings.HasPrefix(spec, "./") || strings.HasPrefix(spec, "../")
}

func resolveRelative(fromPath, spec string, pathToID map[string]string) (string, bool) {
	dir := filepath.Dir(fromPath)
	base := filepath.Join(dir, spec)

	// Try exact match first, then each resolvable extension.
	candidates := []string{base}
	for _, ext := range resolvableExt {
		candidates = append(candidates, base+ext)
	}

	for _, c := range candidates {
		if id, ok := pathToID[filepath.Clean(c)]; ok {
			return id, true
		}
	}
	return "", false
}

func findNode(nodes []graph.Node, id string) *graph.Node {
	for i := range nodes {
		if nodes[i].ID == id {
			return &nodes[i]
		}
	}
	return nil
}