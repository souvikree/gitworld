package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/<you>/codeworld/analyzer/internal/config"
	"github.com/<you>/codeworld/analyzer/internal/extract"
	"github.com/<you>/codeworld/analyzer/internal/graph"
	"github.com/<you>/codeworld/analyzer/internal/logger"
	"github.com/<you>/codeworld/analyzer/internal/parser"
	"github.com/<you>/codeworld/analyzer/internal/walker"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}

func run() error {
	path := flag.String("path", ".", "repo path to analyze")
	flag.Parse()

	// Config load is here mainly for LOG_LEVEL; DB config unused by CLI yet.
	cfg, err := config.Load()
	if err != nil {
		// CLI mode doesn't need Neo4j creds — degrade gracefully, don't crash.
		cfg = &config.Config{LogLevel: "info"}
	}

	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("logger init: %w", err)
	}
	defer log.Sync()

	res, err := walker.Walk(*path, log)
	if err != nil {
		return fmt.Errorf("walk failed: %w", err)
	}

	ctx := context.Background()
	full := graph.Graph{}
	var parseFailures int

	for _, f := range res.Files {
		ast, err := parser.ParseFile(ctx, f)
		if err != nil && ast == nil {
			// hard failure, no usable tree at all
			parseFailures++
			log.Warn("parse failed", zapErr("path", f), zapErr("error", err))
			continue
		}
		if err != nil {
			// partial tree with syntax errors — still counted, still logged
			parseFailures++
			log.Warn("parsed with errors", zapErr("path", f), zapErr("error", err))
		}

		g, err := extract.FromAST(ast)
		if err != nil {
			parseFailures++
			log.Warn("extract failed", zapErr("path", f), zapErr("error", err))
			continue
		}
		full.Nodes = append(full.Nodes, g.Nodes...)
		full.Edges = append(full.Edges, g.Edges...)
	}

	log.Info("analysis complete",
		zapInt("files_walked", len(res.Files)),
		zapInt("parse_failures", parseFailures),
		zapInt("nodes", len(full.Nodes)),
		zapInt("edges", len(full.Edges)),
	)

	out, err := json.MarshalIndent(full, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal graph: %w", err)
	}
	fmt.Println(string(out))
	return nil
}