// Package grep implementiert eine Mini-Suche über Workspace-Files.
package grep

import (
	"os"
	"path/filepath"
	"strings"
)

// Hit repräsentiert ein Match in einer Datei.
type Hit struct {
	Path  string
	Line  int
	Match string
}

// Search durchsucht alle Files unterhalb root nach pattern.
// Maximale Tiefe: rekursiv (filepath.Walk).
func Search(root, pattern string) []Hit {
	var hits []Hit
	if pattern == "" {
		return hits
	}
	patternLower := strings.ToLower(pattern)
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".md") && !strings.HasSuffix(path, ".markdown") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		for i, line := range strings.Split(string(data), "\n") {
			if strings.Contains(strings.ToLower(line), patternLower) {
				hits = append(hits, Hit{Path: path, Line: i, Match: line})
			}
		}
		return nil
	})
	return hits
}
