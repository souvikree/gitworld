package walker

import (
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

var skipDirs = map[string]bool{
	"node_modules": true,
	".git":         true,
	"dist":         true,
	"build":        true,
	"vendor":       true,
}

var validExt = map[string]bool{
	".js": true, ".jsx": true, ".ts": true, ".tsx": true,
}

type Result struct {
	Files  []string
	Errors []error // collected, not swallowed
}

func Walk(root string, log *zap.Logger) (*Result, error) {
	// Reject path traversal / anything escaping intended root
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("walker: invalid root path: %w", err)
	}
	if _, err := os.Stat(absRoot); err != nil {
		return nil, fmt.Errorf("walker: root path does not exist or is inaccessible: %w", err)
	}

	res := &Result{}

	err = filepath.WalkDir(absRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// don't abort the whole walk on one bad entry — record and continue
			res.Errors = append(res.Errors, fmt.Errorf("walk error at %s: %w", path, err))
			log.Warn("walk entry error", zap.String("path", path), zap.Error(err))
			return nil
		}
		if d.IsDir() && skipDirs[d.Name()] {
			return filepath.SkipDir
		}
		if !d.IsDir() && validExt[strings.ToLower(filepath.Ext(path))] {
			res.Files = append(res.Files, path)
		}
		return nil
	})
	if err != nil {
		return res, fmt.Errorf("walker: walk failed: %w", err)
	}

	log.Info("walk complete", zap.Int("files_found", len(res.Files)), zap.Int("errors", len(res.Errors)))
	return res, nil
}