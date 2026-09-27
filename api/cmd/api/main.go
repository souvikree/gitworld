package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/souvikree/gitworld/api/internal/config"
	"github.com/souvikree/gitworld/api/internal/handlers"
	"github.com/souvikree/gitworld/api/internal/logger"
	"github.com/souvikree/gitworld/api/internal/middleware"
	"github.com/souvikree/gitworld/api/internal/store"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("logger: %w", err)
	}
	defer log.Sync()

	ctx := context.Background()

	neo4jStore, err := store.NewNeo4jStore(ctx, cfg.Neo4jURI, cfg.Neo4jUser, cfg.Neo4jPassword, log)
	if err != nil {
		return fmt.Errorf("neo4j connect: %w", err)
	}
	defer neo4jStore.Close(ctx)

	router := gin.New()
	router.Use(gin.Recovery()) // never let a panic in a handler crash the whole server

		router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // Vite dev server
		AllowMethods:     []string{"GET"},
		AllowHeaders:     []string{"X-API-Key", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	protected := router.Group("/v1")
	protected.Use(middleware.APIKeyAuth(cfg.APIKey))
	protected.Use(middleware.RateLimit(5, 10))

	graphHandler := &handlers.GraphHandler{Store: neo4jStore, Log: log}
	// protected.GET("/graph", graphHandler.GetGraph)
	protected.GET("/graph/:repoId", graphHandler.GetGraph)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second, // per your standards: explicit timeouts, always
		WriteTimeout: 10 * time.Second,
	}

	log.Info("api server starting", zap.String("port", cfg.Port))
	return srv.ListenAndServe()
}
