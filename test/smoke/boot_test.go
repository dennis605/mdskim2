package smoke

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dennis605/mdskim2/internal/app"
)

// TestBootView prüft, dass die initiale View() der App einen gültigen
// Bootstrap-Snapshot erzeugt. Dies ist der MVP-R1-Gate-Smoke-Test.
func TestBootView(t *testing.T) {
	m := app.New("testdata/sample-workspace")

	// Wir testen direkt ohne bubbletea-Runtime, weil die Snapshot-View
	// nur vom Model abhängt, nicht von der Program-Loop.
	snapshot := m.View()

	if !strings.Contains(snapshot, "mdskim2") {
		t.Errorf("erwartet Header 'mdskim2' in View, got:\n%s", snapshot)
	}
	if !strings.Contains(snapshot, "FILES") {
		t.Errorf("erwartet Sidebar-Label 'FILES' in View")
	}
	if !strings.Contains(snapshot, "TOC") {
		t.Errorf("erwartet Right-Pane Tab 'TOC' in View")
	}
	if !strings.Contains(snapshot, "Preview") {
		t.Errorf("erwartet Right-Pane Tab 'Preview' in View")
	}
	if !strings.Contains(snapshot, "Backlinks") {
		t.Errorf("erwartet Right-Pane Tab 'Backlinks' in View")
	}
	if !strings.Contains(snapshot, "Ctrl+S") {
		t.Errorf("erwartet Shortcut-Hint 'Ctrl+S' in Footer")
	}
	if !strings.Contains(snapshot, "Ctrl+Q") {
		t.Errorf("erwartet 'Ctrl+Q' in Footer für Quit")
	}
	if !strings.Contains(snapshot, "testdata/sample-workspace") {
		t.Errorf("erwartet Workspace-Pfad im Header")
	}
}

// TestQuitVerarbeitetCtrlQ prüft, dass die App auf Ctrl+Q mit Quit antwortet.
func TestQuitVerarbeitetCtrlQ(t *testing.T) {
	m := app.New("testdata/sample-workspace")

	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlQ}
	_, cmd := m.Update(keyMsg)

	if cmd == nil {
		t.Error("erwartet Quit-Command nach Ctrl+Q, got nil")
	}
}
