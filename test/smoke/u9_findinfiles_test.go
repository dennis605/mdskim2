package smoke

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dennis605/mdskim2/internal/app"
)

func TestU9_FindInFilesOpens(t *testing.T) {
	dir := t.TempDir()
	m := app.New(dir)
	updated, _ := m.Update(app.FindInFilesTriggerMsg{})
	mm := updated.(app.Model)
	if !mm.FindInFilesModeForTest() {
		t.Errorf("expected find-in-files mode to be active")
	}
	if mm.FindInFilesQueryForTest() != "" {
		t.Errorf("expected empty query initially, got %q", mm.FindInFilesQueryForTest())
	}
}

func TestU9_FindInFilesSearchesFiles(t *testing.T) {
	dir := t.TempDir()
	// Setup workspace with files
	if err := os.WriteFile(filepath.Join(dir, "a.md"),
		[]byte("# Title\nhello world\nfoo bar"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.md"),
		[]byte("# Other\nhello again"), 0644); err != nil {
		t.Fatal(err)
	}
	m := app.New(dir)
	updated, _ := m.Update(app.FindInFilesTriggerMsg{})
	mm := updated.(app.Model)
	// Type "hello"
	updated2, _ := mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("hello")})
	mm2 := updated2.(app.Model)
	if !mm2.FindInFilesModeForTest() {
		t.Errorf("expected mode to stay active after typing")
	}
	results := mm2.FindInFilesResultsForTest()
	if len(results) != 2 {
		t.Errorf("expected 2 hits, got %d", len(results))
	}
	if mm2.FindInFilesQueryForTest() != "hello" {
		t.Errorf("expected query 'hello', got %q", mm2.FindInFilesQueryForTest())
	}
}

func TestU9_FindInFilesEscCancels(t *testing.T) {
	dir := t.TempDir()
	m := app.New(dir)
	updated, _ := m.Update(app.FindInFilesTriggerMsg{})
	updated2, _ := updated.(app.Model).Update(tea.KeyMsg{Type: tea.KeyEsc})
	mm := updated2.(app.Model)
	if mm.FindInFilesModeForTest() {
		t.Errorf("expected mode to be off after Esc, but still on")
	}
}

func TestU9_FindInFilesBackspace(t *testing.T) {
	dir := t.TempDir()
	// 'xylophone' is unique; backspace to 'xyloph' should match both still
	if err := os.WriteFile(filepath.Join(dir, "x.md"),
		[]byte("xylophone one\nxylophone two"), 0644); err != nil {
		t.Fatal(err)
	}
	m := app.New(dir)
	updated, _ := m.Update(app.FindInFilesTriggerMsg{})
	mm := updated.(app.Model)
	updated2, _ := mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("xylophone")})
	mm2 := updated2.(app.Model)
	if len(mm2.FindInFilesResultsForTest()) != 2 {
		t.Fatalf("expected 2 hits for 'xylophone', got %d", len(mm2.FindInFilesResultsForTest()))
	}
	// Backspace removes last char → "xylophon" still matches both (substring)
	updated3, _ := mm2.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	mm3 := updated3.(app.Model)
	if mm3.FindInFilesQueryForTest() != "xylophon" {
		t.Errorf("expected query 'xylophon' after backspace, got %q", mm3.FindInFilesQueryForTest())
	}
	if len(mm3.FindInFilesResultsForTest()) != 2 {
		t.Errorf("expected 2 hits for 'xylophon', got %d", len(mm3.FindInFilesResultsForTest()))
	}
	// Empty query → no hits
	for i := 0; i < 8; i++ {
		updated4, _ := mm3.Update(tea.KeyMsg{Type: tea.KeyBackspace})
		mm3 = updated4.(app.Model)
	}
	if mm3.FindInFilesQueryForTest() != "" {
		t.Errorf("expected empty query after many backspaces, got %q", mm3.FindInFilesQueryForTest())
	}
	if len(mm3.FindInFilesResultsForTest()) != 0 {
		t.Errorf("expected 0 hits for empty query, got %d", len(mm3.FindInFilesResultsForTest()))
	}
}

func TestU9_FindInFilesEnterOpensFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "z.md"),
		[]byte("# Title\nunique-marker-here"), 0644); err != nil {
		t.Fatal(err)
	}
	m := app.New(dir)
	updated, _ := m.Update(app.FindInFilesTriggerMsg{})
	mm := updated.(app.Model)
	updated2, _ := mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("unique")})
	mm2 := updated2.(app.Model)
	if len(mm2.FindInFilesResultsForTest()) == 0 {
		t.Fatalf("expected at least 1 hit for 'unique'")
	}
	// Enter opens the file
	updated3, _ := mm2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mm3 := updated3.(app.Model)
	if mm3.FindInFilesModeForTest() {
		t.Errorf("expected mode to close after Enter")
	}
	if !filepath.IsAbs(mm3.CurrentFileForTest()) {
		t.Errorf("expected absolute path to file, got %q", mm3.CurrentFileForTest())
	}
	if filepath.Base(mm3.CurrentFileForTest()) != "z.md" {
		t.Errorf("expected loaded file z.md, got %s", filepath.Base(mm3.CurrentFileForTest()))
	}
}
