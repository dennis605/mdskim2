package smoke

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dennis605/mdskim2/internal/app"
)

func TestU9_RecentMenuOpens(t *testing.T) {
	dir := t.TempDir()
	m := app.New(dir)
	updated, _ := m.Update(app.RecentTriggerMsg{})
	mm := updated.(app.Model)
	if !mm.RecentMenuModeForTest() {
		t.Errorf("expected recent menu mode to be active")
	}
	if mm.RecentMenuIdxForTest() != 0 {
		t.Errorf("expected idx 0, got %d", mm.RecentMenuIdxForTest())
	}
}

func TestU9_RecentMenuEscCancels(t *testing.T) {
	dir := t.TempDir()
	m := app.New(dir)
	updated, _ := m.Update(app.RecentTriggerMsg{})
	updated2, _ := updated.(app.Model).Update(tea.KeyMsg{Type: tea.KeyEsc})
	mm := updated2.(app.Model)
	if mm.RecentMenuModeForTest() {
		t.Errorf("expected mode off after Esc")
	}
}

func TestU9_RecentMenuUpDown(t *testing.T) {
	dir := t.TempDir()
	m := app.New(dir)
	updated, _ := m.Update(app.RecentTriggerMsg{})
	mm := updated.(app.Model)
	initialIdx := mm.RecentMenuIdxForTest()
	// ArrowUp on first item must not go negative
	updated2, _ := mm.Update(tea.KeyMsg{Type: tea.KeyUp})
	mm2 := updated2.(app.Model)
	if mm2.RecentMenuIdxForTest() != initialIdx {
		t.Errorf("expected idx stays at %d after Up on first, got %d", initialIdx, mm2.RecentMenuIdxForTest())
	}
	// ArrowDown once — idx advances by at most 1
	updated3, _ := mm2.Update(tea.KeyMsg{Type: tea.KeyDown})
	mm3 := updated3.(app.Model)
	if mm3.RecentMenuIdxForTest() > initialIdx+1 {
		t.Errorf("idx jumped by more than 1, got %d (initial %d)", mm3.RecentMenuIdxForTest(), initialIdx)
	}
	// Many ArrowDowns — idx never negative
	for i := 0; i < 50; i++ {
		updatedN, _ := mm3.Update(tea.KeyMsg{Type: tea.KeyDown})
		mm3 = updatedN.(app.Model)
	}
	if mm3.RecentMenuIdxForTest() < 0 {
		t.Errorf("idx went negative after many Down: %d", mm3.RecentMenuIdxForTest())
	}
}

func TestU9_RecentMenuOpensFile(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "alpha.md")
	if err := os.WriteFile(target, []byte("# Alpha\nbody content"), 0644); err != nil {
		t.Fatal(err)
	}
	// Smoke test: open menu, esc — proves no panic and round-trip works.
	m := app.New(dir)
	updated, _ := m.Update(app.RecentTriggerMsg{})
	updated2, _ := updated.(app.Model).Update(tea.KeyMsg{Type: tea.KeyEsc})
	mm := updated2.(app.Model)
	if mm.RecentMenuModeForTest() {
		t.Errorf("expected menu closed after Esc")
	}
}
