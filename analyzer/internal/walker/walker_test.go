package walker_test

import (
	"testing"

	"go.uber.org/zap"
	"github.com/souvikree/gitworld/analyzer/internal/walker"
)

func TestWalk_SkipsNodeModulesAndGit(t *testing.T) {
	log := zap.NewNop()
	res, err := walker.Walk("../../testdata/fixtures/skip-dirs", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, f := range res.Files {
		if contains(f, "node_modules") || contains(f, ".git") {
			t.Errorf("walker should have skipped %s", f)
		}
	}
}

func TestWalk_NonexistentRoot(t *testing.T) {
	log := zap.NewNop()
	_, err := walker.Walk("/definitely/does/not/exist", log)
	if err == nil {
		t.Error("expected error for nonexistent root, got nil")
	}
}