package graph

type NodeType string

const (
	NodeFile     NodeType = "File"
	NodeFunction NodeType = "Function"
	NodeModule   NodeType = "Module"
)

type Node struct {
	ID       string         `json:"id"`
	Type     NodeType       `json:"type"`
	Path     string         `json:"path"`
	Name     string         `json:"name"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type EdgeType string

const (
	EdgeImports EdgeType = "IMPORTS"
	EdgeCalls   EdgeType = "CALLS"
)

type Edge struct {
	From string   `json:"from"`
	To   string   `json:"to"`
	Type EdgeType `json:"type"`
}

type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}