package smoke

import (
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dennis605/mdskim2/internal/app"
)

func demoDir() string {
	a, _ := filepath.Abs("../../demo/Test Engineering")
	return a
}

func openFile() app.Model {
	m := app.New(demoDir())
	// Y=6 → README.md (known to load successfully per U7)
	updated, _ := m.Update(tea.MouseMsg{X: 5, Y: 6, Type: tea.MouseLeft})
	return updated.(app.Model)
}

func TestU8_AutoPairParens(t *testing.T) {
	mm := openFile()
	if mm.BufferForTest() == nil {
		t.Skipf("setup failed: no buffer (currentFile=%q)", mm.CurrentFileForTest())
	}
	updated, _ := mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'('}})
	mm = updated.(app.Model)
	buf := mm.BufferForTest()
	if !strings.Contains(strings.Join(buf.Lines, "\n"), "()") {
		t.Errorf("expected auto-pair '()' somewhere, got %v", buf.Lines)
	}
}

func TestU8_AutoPairBrackets(t *testing.T) {
	mm := openFile()
	if mm.BufferForTest() == nil {
		t.Skipf("setup failed: no buffer")
	}
	updated, _ := mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	mm = updated.(app.Model)
	buf := mm.BufferForTest()
	if !strings.Contains(strings.Join(buf.Lines, "\n"), "[]") {
		t.Errorf("expected '[]' somewhere, got %v", buf.Lines)
	}
}

func TestU8_AutoPairSkipClosing(t *testing.T) {
	mm := openFile()
	if mm.BufferForTest() == nil {
		t.Skipf("setup failed: no buffer")
	}
	// Move cursor to start so we know where we are
	before := mm.BufferForTest().CursorCol
	// Type '(' — autopaired
	updated, _ := mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'('}})
	mm = updated.(app.Model)
	// Type ')' — should skip over, no double-close
	updated, _ = mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{')'}})
	mm = updated.(app.Model)
	buf := mm.BufferForTest()
	if buf.CursorCol != before+2 {
		t.Errorf("expected cursor at col=%d (after ( and skip)), got %d", before+2, buf.CursorCol)
	}
}
