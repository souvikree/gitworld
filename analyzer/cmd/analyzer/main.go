package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/souvikree/gitworld/analyzer/internal/config"
	"github.com/souvikree/gitworld/analyzer/internal/extract"
	"github.com/souvikree/gitworld/analyzer/internal/graph"
	"github.com/souvikree/gitworld/analyzer/internal/logger"
	"github.com/souvikree/gitworld/analyzer/internal/parser"
	"github.com/souvikree/gitworld/analyzer/internal/resolve"
	"github.com/souvikree/gitworld/analyzer/internal/store"
	"github.com/souvikree/gitworld/analyzer/internal/walker"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}

func run() error {
	path := flag.String("path", ".", "repo path to analyze")
	repoID := flag.String("repo-id", "", "unique identifier for this repo (required for multi-tenant ingestion)")
	flag.Parse()

	if *repoID == "" {
		return fmt.Errorf("-repo-id is required")
	}

	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{LogLevel: "info"}
	}

	
	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("logger init: %w", err)
	}
	defer log.Sync()
	
	ctx := context.Background()

	neo4jStore, err := store.NewNeo4jStore(ctx, cfg.Neo4jURI, cfg.Neo4jUser, cfg.Neo4jPassword, log)
	if err != nil {
		return fmt.Errorf("neo4j connect: %w", err)
	}
	defer neo4jStore.Close(ctx)

	if err := neo4jStore.EnsureSchema(ctx); err != nil {
		return fmt.Errorf("schema setup: %w", err)
	}
	
	res, err := walker.Walk(*path, log)
	if err != nil {
		return fmt.Errorf("walk failed: %w", err)
	}

	// ctx := context.Background()
	full := graph.Graph{}
	var parseFailures int

	for _, f := range res.Files {
		ast, err := parser.ParseFile(ctx, f)
		if err != nil && ast == nil {
			parseFailures++
			log.Warn("parse failed", zap.String("path", f), zap.Error(err))
			continue
		}
		if err != nil {
			parseFailures++
			log.Warn("parsed with errors", zap.String("path", f), zap.Error(err))
		}

		g, err := extract.FromAST(ast, *repoID)
		if err != nil {
			parseFailures++
			log.Warn("extract failed", zap.String("path", f), zap.Error(err))
			continue
		}
		full.Nodes = append(full.Nodes, g.Nodes...)
		full.Edges = append(full.Edges, g.Edges...)
	}

	resolve.Resolve(&full)

	if err := neo4jStore.IngestGraph(ctx, &full); err != nil {
		return fmt.Errorf("ingestion: %w", err)
	}

	log.Info("analysis complete",
		zap.Int("files_walked", len(res.Files)),
		zap.Int("parse_failures", parseFailures),
		zap.Int("nodes", len(full.Nodes)),
		zap.Int("edges", len(full.Edges)),
	)

	out, err := json.MarshalIndent(full, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal graph: %w", err)
	}
	fmt.Println(string(out))
	return nil
}
