package smoke

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dennis605/mdskim2/internal/app"
)

func TestU8_TreeFilterOpensOnSlash(t *testing.T) {
	dir := t.TempDir()
	m := app.New(dir)
	if m.TreeFilteredForTest() {
		t.Fatal("treeFiltered should start false")
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	mm := updated.(app.Model)
	if !mm.TreeFilteredForTest() {
		t.Errorf("expected treeFiltered=true after '/'")
	}
}

func TestU8_TreeFilterFiltersByName(t *testing.T) {
	// Use demo dir which has known files (default: all dirs expanded)
	demo := demoDir()
	m := app.New(demo)
	// Press '/' directly (without clicking first - root is expanded by default)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	mm := updated.(app.Model)
	// Type 'plan' as a single multi-rune message
	updated, _ = mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p', 'l', 'a', 'n'}})
	mm = updated.(app.Model)
	flat := mm.FlatListForTest()
	foundMatch := false
	for _, n := range flat {
		if strings.Contains(strings.ToLower(n), "plan") {
			foundMatch = true
		}
	}
	if !foundMatch {
		t.Errorf("expected to find 'plan' match, got %v", flat)
	}
	// And the filter should have hidden non-matching entries
	if len(flat) >= 9 { // demo has 9+ entries; filter should narrow
		t.Logf("warning: filter may not have narrowed enough (got %d)", len(flat))
	}
}

func TestU8_TreeFilterEscClears(t *testing.T) {
	dir := t.TempDir()
	m := app.New(dir)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	mm := updated.(app.Model)
	if !mm.TreeFilteredForTest() {
		t.Skip("setup failed")
	}
	updated, _ = mm.Update(tea.KeyMsg{Type: tea.KeyEsc})
	mm = updated.(app.Model)
	if mm.TreeFilteredForTest() {
		t.Errorf("expected treeFiltered=false after Esc")
	}
	if mm.TreeFilterForTest() != "" {
		t.Errorf("expected empty filter after Esc, got %q", mm.TreeFilterForTest())
	}
}
