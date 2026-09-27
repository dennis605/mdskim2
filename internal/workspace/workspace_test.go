package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadWorkspace(t *testing.T) {
	// Testdata: tmp-Verzeichnis mit Struktur
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, "sub", "deep"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "a.md"), []byte("# A"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "sub", "b.md"), []byte("# B"), 0644); err != nil {
		t.Fatal(err)
	}

	ws, err := Load(tmp)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if ws.TotalFiles() != 2 {
		t.Errorf("erwartet 2 files, got %d", ws.TotalFiles())
	}

	if !ws.Root.IsDir {
		t.Error("Root muss Verzeichnis sein")
	}
}

func TestRenderTree(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "alpha.md"), []byte("# A"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "zulu.md"), []byte("# Z"), 0644); err != nil {
		t.Fatal(err)
	}

	ws, _ := Load(tmp)
	r := NewTreeRenderer()
	out := r.Render(ws)

	if !strings.Contains(out, "alpha.md") {
		t.Errorf("erwartet alpha.md in: %s", out)
	}
	if !strings.Contains(out, "zulu.md") {
		t.Errorf("erwartet zulu.md in: %s", out)
	}
}

func TestToggleDir(t *testing.T) {
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, "d"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "d", "inner.md"), []byte("# I"), 0644); err != nil {
		t.Fatal(err)
	}

	ws, _ := Load(tmp)
	r := NewTreeRenderer()
	dirPath := filepath.Join(tmp, "d")

	// Initial: expanded, inner.md sichtbar
	if !strings.Contains(r.Render(ws), "inner.md") {
		t.Error("init expanded: inner.md sollte sichtbar sein")
	}

	// Collapse
	r.ToggleDir(dirPath)
	if strings.Contains(r.Render(ws), "inner.md") {
		t.Error("nach collapse: inner.md sollte weg sein")
	}

	// Expand wieder
	r.ToggleDir(dirPath)
	if !strings.Contains(r.Render(ws), "inner.md") {
		t.Error("nach expand: inner.md sollte wieder sichtbar sein")
	}
}

func TestIgnoriertHiddenFiles(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, ".hidden"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "visible.md"), []byte("# V"), 0644); err != nil {
		t.Fatal(err)
	}

	ws, _ := Load(tmp)
	if ws.TotalFiles() != 1 {
		t.Errorf("erwartet 1 sichtbare Datei, got %d", ws.TotalFiles())
	}
}
