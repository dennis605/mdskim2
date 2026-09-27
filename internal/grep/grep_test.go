package grep

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestSearch(t *testing.T) {
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "a.md"), "# Title\nFoo bar\n")
	writeFile(t, filepath.Join(tmp, "sub", "b.md"), "Baz qux\n")

	hits := Search(tmp, "foo")
	if len(hits) != 1 {
		t.Fatalf("erwartet 1 Hit, got %d", len(hits))
	}
	if hits[0].Line != 1 {
		t.Errorf("Line sollte 1 sein, got %d", hits[0].Line)
	}
}

func TestSearchEmptyPattern(t *testing.T) {
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "a.md"), "foo")
	hits := Search(tmp, "")
	if len(hits) != 0 {
		t.Errorf("empty pattern sollte 0 hits geben, got %d", len(hits))
	}
}
