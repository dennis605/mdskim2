package smoke

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dennis605/mdskim2/internal/app"
)

func writeTestFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func openFileViaEnter(t *testing.T, m app.Model, path string) app.Model {
	t.Helper()
	for i, n := range m.FlatList() {
		if n.Path == path {
			m = m.WithCursor(i)
			break
		}
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	return updated.(app.Model)
}

func TestAppR5LadeDateiMitMarkdown(t *testing.T) {
	tmp := t.TempDir()
	md := "# Title\n\nSome **bold** text.\n"
	path := writeTestFile(t, tmp, "doc.md", md)

	m := app.New(tmp)
	m = openFileViaEnter(t, m, path)

	if !m.IsBufferOpen() {
		t.Fatal("Buffer sollte offen sein")
	}

	hl := m.HighlightedLines()
	if len(hl) == 0 {
		t.Fatal("HighlightedLines sollte nicht leer sein")
	}
	// hl[0] = "# Title", hl[1] = empty, hl[2] = "Some **bold** text.", hl[3] = empty
	// Title ist in Zeile 0
	titleFound := false
	for _, line := range hl {
		if strings.Contains(line, "Title") {
			titleFound = true
			break
		}
	}
	if !titleFound {
		t.Errorf("Title sollte in HighlightedLines sein: %v", hl)
	}
}

func TestAppR5HighlightBoldErkannt(t *testing.T) {
	tmp := t.TempDir()
	md := "Some **bold** text"
	path := writeTestFile(t, tmp, "doc.md", md)

	m := app.New(tmp)
	m = openFileViaEnter(t, m, path)

	hl := m.HighlightedLines()
	// Zeile 0 ist die Heading-Line "# filename", Zeile 1 ist "Some **bold** text"
	var found bool
	for _, line := range hl {
		if strings.Contains(line, "\x1b[1mbold\x1b[22m") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Bold-ANSI fehlt in HighlightedLines: %v", hl)
	}
}

func TestAppR5TOCZeigtHierarchie(t *testing.T) {
	tmp := t.TempDir()
	md := "# H1\n\nPara\n\n## H2\n\nText\n\n### H3\n\nSub\n"
	path := writeTestFile(t, tmp, "doc.md", md)

	m := app.New(tmp)
	m = openFileViaEnter(t, m, path)

	if !m.IsBufferOpen() {
		t.Fatal("Buffer sollte offen sein")
	}
	toc := m.TocText()
	if !strings.Contains(toc, "H1") {
		t.Errorf("TOC sollte H1 enthalten: %q", toc)
	}
	if !strings.Contains(toc, "H2") {
		t.Errorf("TOC sollte H2 enthalten: %q", toc)
	}
	if !strings.Contains(toc, "H3") {
		t.Errorf("TOC sollte H3 enthalten: %q", toc)
	}
	// Hierarchie
	if strings.Count(toc, "  ▸ H2") < 1 {
		t.Errorf("H2 sollte eingerückt sein: %q", toc)
	}
	if strings.Count(toc, "    ▸ H3") < 1 {
		t.Errorf("H3 sollte tiefer eingerückt sein: %q", toc)
	}
}

func TestAppR5HeadingsMitHierarchieCount(t *testing.T) {
	tmp := t.TempDir()
	md := "# H1\n## H2\n### H3\n#### H4\n"
	path := writeTestFile(t, tmp, "doc.md", md)

	m := app.New(tmp)
	m = openFileViaEnter(t, m, path)

	if got := m.HeadingsCount(); got != 4 {
		t.Errorf("erwartet 4 Headings, got %d", got)
	}
}

func TestR5QuitCtrlQ(t *testing.T) {
	tmp := t.TempDir()
	m := app.New(tmp)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlQ})
	if !updated.(app.Model).Quitting() {
		t.Error("Quit sollte Quitting=true setzen")
	}
}
