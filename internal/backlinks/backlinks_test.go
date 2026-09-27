package backlinks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCollect_NoBacklinks(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.md"), []byte("no links here"), 0644); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "target.md")
	if err := os.WriteFile(target, []byte("# Target"), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := Collect(dir, target)
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 backlinks, got %d: %+v", len(got), got)
	}
}

func TestCollect_FindsBacklinks(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "Target.md")
	if err := os.WriteFile(target, []byte("# Target"), 0644); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(dir, "source.md")
	content := "# Source\nThis links to [[Target]].\nAlso [[Target|the alias]].\nAnd a non-matching [[Other]] link.\n"
	if err := os.WriteFile(source, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := Collect(dir, target)
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 backlinks, got %d: %+v", len(got), got)
	}
}

func TestCollect_SkipsTargetFile(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "Target.md")
	content := "# Target\nSelf-link: [[Target]]\n"
	if err := os.WriteFile(target, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := Collect(dir, target)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 (target file skipped), got %d: %+v", len(got), got)
	}
}

func TestCollect_AliasedLink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "Notes.md")
	if err := os.WriteFile(target, []byte("# Notes"), 0644); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(dir, "index.md")
	if err := os.WriteFile(source, []byte("See [[Notes|the notes file]] for details"), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := Collect(dir, target)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1, got %d", len(got))
	}
	if got[0].LinkAlias != "the notes file" {
		t.Errorf("alias mismatch: %q", got[0].LinkAlias)
	}
}

func TestCollect_IgnoresHiddenDirs(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "Target.md")
	if err := os.WriteFile(target, []byte("# T"), 0644); err != nil {
		t.Fatal(err)
	}
	hidden := filepath.Join(dir, ".git")
	if err := os.MkdirAll(hidden, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hidden, "internal.md"),
		[]byte("[[Target]]"), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := Collect(dir, target)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("hidden .git should be ignored, got %d: %+v", len(got), got)
	}
}
