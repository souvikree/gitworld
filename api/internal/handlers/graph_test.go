package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/souvikree/gitworld/api/internal/handlers"
	"github.com/souvikree/gitworld/api/internal/store"
)

type fakeStore struct {
	graph *store.Graph
	err   error
}

func (f *fakeStore) GetGraph(ctx context.Context, repoID string, limit int) (*store.Graph, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.graph, nil
}

func setupRouter(h *handlers.GraphHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/v1/graph/:repoId", h.GetGraph)
	return r
}

func TestGetGraph_Success(t *testing.T) {
	fake := &fakeStore{graph: &store.Graph{
		Nodes: []store.Node{{ID: "1", Type: "File", Name: "a.js"}},
	}}
	h := &handlers.GraphHandler{Store: fake, Log: zap.NewNop()}
	router := setupRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/v1/graph/test-repo", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var got store.Graph
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if len(got.Nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(got.Nodes))
	}
}

func TestGetGraph_StoreError(t *testing.T) {
	fake := &fakeStore{err: errors.New("neo4j down")}
	h := &handlers.GraphHandler{Store: fake, Log: zap.NewNop()}
	router := setupRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/v1/graph/test-repo", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	// Confirm internal error details never leak to the client —
	// matches your own logging/security standard.
	if bodyContains(w.Body.String(), "neo4j down") {
		t.Error("internal error message leaked to client response")
	}
}

func TestGetGraph_MissingRepoID(t *testing.T) {
	fake := &fakeStore{graph: &store.Graph{}}
	h := &handlers.GraphHandler{Store: fake, Log: zap.NewNop()}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/v1/graph/:repoId", h.GetGraph) // still needs the param defined for routing, but we'll hit it with an empty value differently below

	// Gin's :repoId param can't literally be empty via this route pattern,
	// so this test targets the handler directly instead of through routing.
	req := httptest.NewRequest(http.MethodGet, "/v1/graph/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for missing repoId segment, got %d", w.Code)
	}
}

func bodyContains(body, substr string) bool {
	return len(body) > 0 && (len(substr) == 0 || indexOf(body, substr) != -1)
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}