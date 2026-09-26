package parser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

type FileAST struct {
	Path   string
	Source []byte
	Tree   *sitter.Tree
}

// ParseFile reads and parses a single file. Errors are always returned,
// never swallowed — callers decide whether to skip-and-log.
func ParseFile(ctx context.Context, path string) (*FileAST, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("parser: read %s: %w", path, err)
	}

	lang, err := languageFor(path)
	if err != nil {
		return nil, fmt.Errorf("parser: %s: %w", path, err)
	}

	p := sitter.NewParser()
	p.SetLanguage(lang)

	tree, err := p.ParseCtx(ctx, nil, src)
	if err != nil {
		return nil, fmt.Errorf("parser: parse %s: %w", path, err)
	}
	if tree.RootNode().HasError() {
		// Not fatal — tree-sitter is error-tolerant and gives a partial tree.
		// We still return it, but the caller should know parsing was imperfect.
		return &FileAST{Path: path, Source: src, Tree: tree},
			fmt.Errorf("parser: %s parsed with syntax errors (partial result)", path)
	}

	return &FileAST{Path: path, Source: src, Tree: tree}, nil
}

func languageFor(path string) (*sitter.Language, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".js", ".jsx":
		return javascript.GetLanguage(), nil
	case ".ts", ".tsx":
		return typescript.GetLanguage(), nil
	default:
		return nil, fmt.Errorf("unsupported extension: %s", filepath.Ext(path))
	}
}