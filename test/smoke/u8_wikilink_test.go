package smoke

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dennis605/mdskim2/internal/app"
)

func TestU8_WikiLinkPopupTriggers(t *testing.T) {
	demo := demoDir()
	m := app.New(demo)
	// Click Y=5 loads Teststrategie-SAP.md
	updated, _ := m.Update(tea.MouseMsg{X: 5, Y: 5, Type: tea.MouseLeft})
	mm := updated.(app.Model)
	if mm.CurrentFileForTest() == "" {
		t.Skip("file didn't load")
	}
	// Type '[' twice
	updated, _ = mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	mm = updated.(app.Model)
	updated, _ = mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	mm = updated.(app.Model)
	if !mm.WikiLinkPopupForTest() {
		t.Errorf("expected wikiLinkPopup=true after typing '[['")
	}
	matches := mm.WikiLinkMatchesForTest()
	if len(matches) == 0 {
		t.Errorf("expected some matches, got 0")
	}
}

func TestU8_WikiLinkEscCloses(t *testing.T) {
	demo := demoDir()
	m := app.New(demo)
	updated, _ := m.Update(tea.MouseMsg{X: 5, Y: 5, Type: tea.MouseLeft})
	mm := updated.(app.Model)
	if mm.CurrentFileForTest() == "" {
		t.Skip("setup failed")
	}
	updated, _ = mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	mm = updated.(app.Model)
	updated, _ = mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	mm = updated.(app.Model)
	if !mm.WikiLinkPopupForTest() {
		t.Skip("popup didn't trigger")
	}
	updated, _ = mm.Update(tea.KeyMsg{Type: tea.KeyEsc})
	mm = updated.(app.Model)
	if mm.WikiLinkPopupForTest() {
		t.Errorf("expected popup closed after Esc")
	}
}

func TestU8_WikiLinkTabAccepts(t *testing.T) {
	demo := demoDir()
	m := app.New(demo)
	updated, _ := m.Update(tea.MouseMsg{X: 5, Y: 5, Type: tea.MouseLeft})
	mm := updated.(app.Model)
	if mm.CurrentFileForTest() == "" {
		t.Skip("setup failed")
	}
	updated, _ = mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	mm = updated.(app.Model)
	updated, _ = mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	mm = updated.(app.Model)
	if !mm.WikiLinkPopupForTest() {
		t.Skip("popup didn't trigger")
	}
	updated, _ = mm.Update(tea.KeyMsg{Type: tea.KeyTab})
	mm = updated.(app.Model)
	if mm.WikiLinkPopupForTest() {
		t.Errorf("expected popup closed after Tab")
	}
	buf := mm.BufferForTest()
	if buf == nil {
		t.Fatal("no buffer")
	}
	all := strings.Join(buf.Lines, "\n")
	if !strings.Contains(all, "[[") {
		t.Errorf("expected buffer to contain wiki link, got %q", all)
	}
}
