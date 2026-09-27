package smoke

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dennis605/mdskim2/internal/app"
)

func TestAppR2Init(t *testing.T) {
	m := app.New("testdata/sample-workspace")
	snapshot := m.View()

	if !strings.Contains(snapshot, "FILES") {
		t.Errorf("erwartet 'FILES' im Sidebar")
	}
	// R2: Editor-Buffer kommt in R3. Wir prüfen Tree statt Editor-Heading.
	if !strings.Contains(snapshot, "sample-workspace") {
		t.Errorf("erwartet sample-workspace im Tree")
	}
	if !strings.Contains(snapshot, "Teststrategie") {
		t.Errorf("erwartet Teststrategie.md im Tree")
	}
}

func TestAppR2Navigation(t *testing.T) {
	m := app.New("testdata/sample-workspace")

	// Erste Cursor-Bewegung: Down
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	model, ok := m2.(app.Model)
	if !ok {
		t.Fatalf("Update returnt kein Model: %T", m2)
	}

	// Snapshot nach Navigation sollte sich unterscheiden (Cursor-Position)
	// Wir prüfen, dass die App stabil bleibt
	snapshot := model.View()
	if len(snapshot) == 0 {
		t.Error("View ist leer nach Navigation")
	}
}

func TestAppR2MouseSidebarClick(t *testing.T) {
	m := app.New("testdata/sample-workspace")

	mouseMsg := tea.MouseMsg{
		Type: tea.MouseLeft,
		X:    5,
		Y:    5,
	}
	m2, _ := m.Update(mouseMsg)
	model, ok := m2.(app.Model)
	if !ok {
		t.Fatalf("Update returnt kein Model")
	}

	// Snapshot muss stabil bleiben
	snapshot := model.View()
	if len(snapshot) == 0 {
		t.Error("View nach Mouse-Click leer")
	}
}
