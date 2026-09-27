package recent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRecentAddDedup(t *testing.T) {
	// Override filePath via New then write
	l := &List{filePath: filepath.Join(t.TempDir(), "recent.json")}
	l.Files = nil
	l.Add("/foo/bar.md")
	l.Add("/foo/baz.md")
	l.Add("/foo/bar.md") // duplicate
	all := l.All()
	if len(all) != 2 {
		t.Errorf("expected 2, got %d", len(all))
	}
	if all[0].Path != "/foo/bar.md" {
		t.Errorf("expected most-recent first, got %q", all[0].Path)
	}
}

func TestRecentPersist(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "recent.json")
	l := &List{filePath: path}
	l.Add("/foo/x.md")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected recent.json created, got %v", err)
	}
}
