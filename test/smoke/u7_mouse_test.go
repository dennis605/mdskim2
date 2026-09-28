package smoke

import (
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dennis605/mdskim2/internal/app"
)

// workspaceDemo resolves "demo/Test Engineering" relative to test/smoke/.
func workspaceDemo() string {
	a, _ := filepath.Abs("../../demo/Test Engineering")
	return a
}

// Layout (post-truncation, demo/Test Engineering):
//   Y=0  Header
//   Y=1  Toolbar
//   Y=2  Body row 0 — FILES label (sidebar) + Tab bar (right) + Editor header
//   Y=3  Body row 1 — tree[0] = "Test Engineer…" (root dir)
//   Y=4  Body row 2 — tree[1] = "Projekte" (subdir)
//   Y=5  Body row 3 — tree[2] = "Teststrategie…" (file)
//   Y=6  Body row 4 — tree[3] = "README.md"
//   Y=7  Body row 5 — tree[4] = "Testarten.md"
//   Y=8  Body row 6 — tree[5] = "Testdesign.md"
//   Y=9  Body row 7 — tree[6] = "Testmethoden…"
//   Y=10 Body row 8 — tree[7] = "Testplanung.md"
//   Y=11 Body row 9 — tree[8] = "Teststrategie…"
//
// Pane X-boundaries (default 120-col layout):
//   Sidebar: X [0..24)
//   Editor:  X [24..88)
//   RightPane: X [88..120)
//
// Right-pane tab-bar:
//   Y=2, tabWidth ≈ 10. Click X in [88+10, 88+20) → TOC (tab 1).

func click(m app.Model, x, y int) app.Model {
	updated, _ := m.Update(tea.MouseMsg{X: x, Y: y, Type: tea.MouseLeft})
	if updated == nil {
		return m
	}
	mm, ok := updated.(app.Model)
	if !ok {
		return m
	}
	return mm
}

func TestU7_ClickOpensFile(t *testing.T) {
	m := app.New(workspaceDemo())

	// Y=5 → flatList[2] → Teststrategie-SAP.md
	mm := click(m, 5, 5)
	if mm.CurrentFileForTest() == "" {
		t.Errorf("expected file loaded after click Y=5, got empty")
	}

	// Y=6 → flatList[3] → README.md
	mm2 := click(app.New(workspaceDemo()), 5, 6)
	if mm2.CurrentFileForTest() == "" || filepath.Base(mm2.CurrentFileForTest()) != "README.md" {
		t.Errorf("expected README.md after click Y=6, got %q", mm2.CurrentFileForTest())
	}
}

func TestU7_ClickSwitchesRightPaneTab(t *testing.T) {
	m := app.New(workspaceDemo())
	m.LoadFileForTest(filepath.Join(workspaceDemo(), "README.md"))

	// Click X in [98..108) → second tab (TOC)
	mm := click(m, 99, 2)
	if mm.RightTab() != 1 {
		t.Errorf("expected rightTab=1 after click on TOC tab, got %d", mm.RightTab())
	}

	// Click X in [108..120) → third tab (Backlinks)
	mm2 := click(m, 115, 2)
	if mm2.RightTab() != 2 {
		t.Errorf("expected rightTab=2 after click on Backlinks tab, got %d", mm2.RightTab())
	}
}

func TestU7_ClickTOCJumpsCursorAndFocusesEditor(t *testing.T) {
	m := app.New(workspaceDemo())
	m.LoadFileForTest(filepath.Join(workspaceDemo(), "README.md"))
	m.SetRightTab(1)

	mm := click(m, 100, 5)
	if mm.FocusForTest() != "editor" {
		t.Errorf("expected focus 'editor' after TOC click, got %q", mm.FocusForTest())
	}
}

func TestU7_ClickInEditorFocusesEditor(t *testing.T) {
	m := app.New(workspaceDemo())
	m.LoadFileForTest(filepath.Join(workspaceDemo(), "README.md"))

	// Click in editor pane (X=30, Y=5)
	mm := click(m, 30, 5)
	if mm.FocusForTest() != "editor" {
		t.Errorf("expected focus 'editor' after click in editor, got %q", mm.FocusForTest())
	}
}

func TestU7_ClickInSidebarFocusesTree(t *testing.T) {
	m := app.New(workspaceDemo())
	m.LoadFileForTest(filepath.Join(workspaceDemo(), "README.md"))
	// Switch focus away from tree first
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})

	// Click on tree entry
	mm := click(m, 5, 6)
	if mm.FocusForTest() != "tree" {
		t.Errorf("expected focus 'tree' after sidebar click, got %q", mm.FocusForTest())
	}
}

func TestU7_ClickOutsideBoundsIsHarmless(t *testing.T) {
	m := app.New(workspaceDemo())

	// Click on header (Y=0) - should be a no-op (or harmless)
	mm := click(m, 0, 0)
	_ = mm // just verify no panic

	// Click in status bar (Y=m.height-1 = 39)
	mm = click(m, 0, 39)
	_ = mm
}
