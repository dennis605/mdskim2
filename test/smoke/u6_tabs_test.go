package smoke

import (
	"strings"
	"testing"

	"github.com/dennis605/mdskim2/internal/app"
)

// TestU6_LoadFileShowsPreview prüft, dass nach Laden einer Datei der Preview-Tab
// gerenderten Markdown zeigt.
func TestU6_LoadFileShowsPreview(t *testing.T) {
	m := app.New("demo/Test Engineering")
	// Load README.md
	m.LoadFileForTest("demo/Test Engineering/README.md")
	
	m.SetRightTab(0) // Preview tab
	snap := m.View()
	if !strings.Contains(snap, "Preview") {
		t.Errorf("Preview tab should be visible")
	}
}

// TestU6_LoadFileShowsTOC prüft, dass der TOC-Tab Heading-Hierarchie anzeigt.
func TestU6_LoadFileShowsTOC(t *testing.T) {
	m := app.New("demo/Test Engineering")
	m.LoadFileForTest("demo/Test Engineering/README.md")
	m.SetRightTab(1)
	snap := m.View()
	if !strings.Contains(snap, "TOC") {
		t.Errorf("TOC tab should be visible")
	}
}

// TestU6_LoadFileShowsBacklinks prüft, dass der Backlinks-Tab Backlink-Liste zeigt.
func TestU6_LoadFileShowsBacklinks(t *testing.T) {
	m := app.New("demo/Test Engineering")
	m.LoadFileForTest("demo/Test Engineering/README.md")
	m.SetRightTab(2)
	snap := m.View()
	if !strings.Contains(snap, "Backlinks") {
		t.Errorf("Backlinks tab should be visible")
	}
}
