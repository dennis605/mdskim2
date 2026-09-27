package smoke

import (
	"strings"
	"testing"

	"github.com/dennis605/mdskim2/internal/app"
)

// TestU6_1_CtrlShiftArrowInView prüft, dass die neuen Shortcuts im Toolbar sind.
func TestU6_1_CtrlShiftArrowInView(t *testing.T) {
	m := app.New("testdata/sample-workspace")
	m.LoadFileForTest("testdata/sample-workspace/README.md")
	snap := m.View()
	if !strings.Contains(snap, "Panes") && !strings.Contains(snap, "panes") {
		t.Errorf("erwartet 'Panes' im Toolbar, got:\n%s", snap)
	}
}

// TestU6_1_CycleFocus_Logic prüft die cycleFocus/setFocus Helpers mit geladenem Buffer.
func TestU6_1_CycleFocus_Logic(t *testing.T) {
	m := app.New("testdata/sample-workspace")
	m.LoadFileForTest("testdata/sample-workspace/README.md")
	
	// Initial focus = tree
	if m.FocusForTest() != "tree" {
		t.Errorf("expected initial focus 'tree', got %q", m.FocusForTest())
	}
	
	// Cycle forward: tree → editor
	next := m.CycleFocusForTest(1)
	m.SetFocusForTest(next)
	if m.FocusForTest() != "editor" {
		t.Errorf("after cycleFocus(+1) expected 'editor', got %q", m.FocusForTest())
	}
	
	// Cycle forward: editor → toc
	next = m.CycleFocusForTest(1)
	m.SetFocusForTest(next)
	if m.FocusForTest() != "toc" {
		t.Errorf("after cycleFocus(+1) expected 'toc', got %q", m.FocusForTest())
	}
	
	// Cycle backward: toc → editor
	prev := m.CycleFocusForTest(-1)
	m.SetFocusForTest(prev)
	if m.FocusForTest() != "editor" {
		t.Errorf("after cycleFocus(-1) expected 'editor', got %q", m.FocusForTest())
	}
}
