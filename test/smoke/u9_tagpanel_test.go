package smoke

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dennis605/mdskim2/internal/app"
)

func TestU9_TagsTabRenders(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "tags.md")
	body := "# Title\n\nThis file has #alpha and #beta tags.\n\nAlso #alpha again.\n"
	if err := os.WriteFile(target, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	m := app.New(dir)
	mm := m
	mm.LoadFileForTest(target)
	// Switch to Tags tab (idx 3)
	mm.SetRightTab(3)
	if mm.RightTab() != 3 {
		t.Fatalf("expected rightTab=3, got %d", mm.RightTab())
	}
	rendered := mm.TagsTabTextForTest()
	if !strings.Contains(rendered, "#alpha") {
		t.Errorf("expected rendered Tags tab to contain '#alpha', got %q", rendered)
	}
	if !strings.Contains(rendered, "#beta") {
		t.Errorf("expected rendered Tags tab to contain '#beta', got %q", rendered)
	}
	if !strings.Contains(rendered, "2x") {
		t.Errorf("expected count '2x' for #alpha, got %q", rendered)
	}
}

func TestU9_TagsTabAlt8Shortcut(t *testing.T) {
	dir := t.TempDir()
	m := app.New(dir)
	// Simulate Alt+8
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'8'}, Alt: true})
	mm := updated.(app.Model)
	if mm.RightTab() != 3 {
		t.Errorf("expected Alt+8 to set rightTab=3, got %d", mm.RightTab())
	}
}

func TestU9_TagsTabEmptyBuffer(t *testing.T) {
	dir := t.TempDir()
	m := app.New(dir)
	mm := m
	mm.SetRightTab(3)
	rendered := mm.TagsTabTextForTest()
	if !strings.Contains(rendered, "keine") {
		t.Errorf("expected friendly hint for empty buffer, got %q", rendered)
	}
}
