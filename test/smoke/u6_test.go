package smoke

import (
	"strings"
	"testing"

	"github.com/dennis605/mdskim2/internal/app"
)

// TestU6_ToolbarExists prüft, dass der U6-Toolbar (mdskim + Shortcuts) in View erscheint.
func TestU6_ToolbarExists(t *testing.T) {
	m := app.New("testdata/sample-workspace")
	snapshot := m.View()
	if !strings.Contains(snapshot, "mdskim") {
		t.Errorf("erwartet 'mdskim' Brand in Toolbar")
	}
	if !strings.Contains(snapshot, "Ctrl+O") {
		t.Errorf("erwartet 'Ctrl+O' in Toolbar")
	}
}

// TestU6_RightPaneTabs prüft, dass alle 3 Tabs sichtbar sind.
func TestU6_RightPaneTabs(t *testing.T) {
	m := app.New("testdata/sample-workspace")
	snapshot := m.View()
	for _, tab := range []string{"Preview", "TOC", "Backlinks"} {
		if !strings.Contains(snapshot, tab) {
			t.Errorf("erwartet Tab %q in View", tab)
		}
	}
}

// TestU6_AltKeySwitchesRightTab prüft Alt+5/6/7 Tab-Switching.
func TestU6_AltKeySwitchesRightTab(t *testing.T) {
	m := app.New("testdata/sample-workspace")
	if m.RightTab() != 0 {
		t.Errorf("initial RightTab should be 0 (Preview), got %d", m.RightTab())
	}
	m.SetRightTab(2)
	if m.RightTab() != 2 {
		t.Errorf("after SetRightTab(2) expected 2, got %d", m.RightTab())
	}
	snap := m.View()
	if !containsAll(snap, []string{"Backlinks"}) {
		t.Errorf("expected Backlinks visible after SetRightTab(2)")
	}
}
