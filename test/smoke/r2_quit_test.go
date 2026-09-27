package smoke

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dennis605/mdskim2/internal/app"
)

func TestR2QuitViaCtrlQ(t *testing.T) {
	m := app.New("testdata/sample-workspace")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlQ})
	if cmd == nil {
		t.Error("erwartet Quit-Command")
	}
}
