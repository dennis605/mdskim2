package smoke

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dennis605/mdskim2/internal/app"
)

// TestU8_NewFile: 'a' im Tree öffnet Prompt, Enter legt Datei an.
func TestU8_NewFile(t *testing.T) {
	dir := t.TempDir()
	m := app.New(dir)

	// Press 'a' to start new-file prompt
	updated, _ := m.Update(sharedKey('a'))
	mm := updated.(app.Model)

	if mm.PromptModeForTest() != "new-file" {
		t.Errorf("expected promptMode=new-file after 'a', got %q", mm.PromptModeForTest())
	}

	// Type filename
	updated, _ = mm.Update(sharedKey('t', 'e', 's', 't', '.', 'm', 'd'))
	mm = updated.(app.Model)
	if !strings.Contains(mm.PromptQueryForTest(), "test.md") {
		t.Errorf("expected promptQuery to contain 'test.md', got %q", mm.PromptQueryForTest())
	}

	// Enter creates the file
	updated, _ = mm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mm = updated.(app.Model)

	if mm.PromptModeForTest() != "" {
		t.Errorf("expected prompt to clear after Enter, got %q", mm.PromptModeForTest())
	}

	// File should exist
	if _, err := os.Stat(filepath.Join(dir, "test.md")); err != nil {
		t.Errorf("expected file test.md to exist: %v", err)
	}

	// File should be in tree
	if mm.CurrentFileForTest() == "" {
		t.Errorf("expected currentFile to be set after new-file")
	}
}

// TestU8_NewFolder: 'A' legt Ordner an.
func TestU8_NewFolder(t *testing.T) {
	dir := t.TempDir()
	m := app.New(dir)

	updated, _ := m.Update(sharedKey('A'))
	mm := updated.(app.Model)
	if mm.PromptModeForTest() != "new-folder" {
		t.Errorf("expected promptMode=new-folder after 'A', got %q", mm.PromptModeForTest())
	}

	updated, _ = mm.Update(sharedKey('s', 'u', 'b'))
	mm = updated.(app.Model)
	updated, _ = mm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mm = updated.(app.Model)

	info, err := os.Stat(filepath.Join(dir, "sub"))
	if err != nil {
		t.Errorf("expected folder 'sub' to exist: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("expected 'sub' to be a directory")
	}
}

// TestU8_Rename: 'r' benennt eine Datei um.
func TestU8_Rename(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "old.md"), []byte("# old"), 0644)

	m := app.New(dir)
	// Expand root
	_, _ = m.Update(tea.MouseMsg{X: 5, Y: 3, Type: tea.MouseLeft})
	// Click on old.md (Y=4 = flatList[1])
	updated, _ := m.Update(tea.MouseMsg{X: 5, Y: 4, Type: tea.MouseLeft})
	mm := updated.(app.Model)
	if !strings.HasSuffix(mm.CurrentFileForTest(), "old.md") {
		t.Skipf("setup failed: %q", mm.CurrentFileForTest())
	}

	// Press 'r' to rename
	updated, _ = mm.Update(sharedKey('r'))
	mm = updated.(app.Model)
	if mm.PromptModeForTest() != "rename" {
		t.Errorf("expected promptMode=rename after 'r', got %q", mm.PromptModeForTest())
	}

	updated, _ = mm.Update(sharedKey('n', 'e', 'w', '.', 'm', 'd'))
	mm = updated.(app.Model)
	updated, _ = mm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mm = updated.(app.Model)

	if _, err := os.Stat(filepath.Join(dir, "new.md")); err != nil {
		t.Errorf("expected new.md to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "old.md")); !os.IsNotExist(err) {
		t.Errorf("expected old.md to be gone")
	}
}

// TestU8_DeleteConfirm: 'd' requires 'y' confirmation.
func TestU8_DeleteConfirm(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "doomed.md"), []byte("# bye"), 0644)

	m := app.New(dir)
	_, _ = m.Update(tea.MouseMsg{X: 5, Y: 3, Type: tea.MouseLeft})
	updated, _ := m.Update(tea.MouseMsg{X: 5, Y: 4, Type: tea.MouseLeft})
	mm := updated.(app.Model)

	// Press 'd'
	updated, _ = mm.Update(sharedKey('d'))
	mm = updated.(app.Model)
	if mm.PromptModeForTest() != "confirm-delete" {
		t.Errorf("expected promptMode=confirm-delete after 'd', got %q", mm.PromptModeForTest())
	}

	// Type 'n' → cancel
	updated, _ = mm.Update(sharedKey('n'))
	mm = updated.(app.Model)
	updated, _ = mm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mm = updated.(app.Model)

	if _, err := os.Stat(filepath.Join(dir, "doomed.md")); err != nil {
		t.Errorf("expected doomed.md to still exist after 'n' confirmation")
	}

	// Now press 'd' and 'y' to confirm
	updated, _ = mm.Update(sharedKey('d'))
	mm = updated.(app.Model)
	updated, _ = mm.Update(sharedKey('y'))
	mm = updated.(app.Model)
	updated, _ = mm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mm = updated.(app.Model)

	if _, err := os.Stat(filepath.Join(dir, "doomed.md")); !os.IsNotExist(err) {
		t.Errorf("expected doomed.md to be deleted after 'y' confirmation")
	}
}

// TestU8_EscCancelsPrompt: Esc leert den Prompt-Mode.
func TestU8_EscCancelsPrompt(t *testing.T) {
	dir := t.TempDir()
	m := app.New(dir)

	updated, _ := m.Update(sharedKey('a'))
	mm := updated.(app.Model)
	if mm.PromptModeForTest() == "" {
		t.Fatal("expected prompt to be active")
	}

	updated, _ = mm.Update(tea.KeyMsg{Type: tea.KeyEsc})
	mm = updated.(app.Model)
	if mm.PromptModeForTest() != "" {
		t.Errorf("expected prompt to be cleared after Esc, got %q", mm.PromptModeForTest())
	}
}

// TestU8_RefreshTree: F5 lädt Tree neu (file added externally appears).
func TestU8_RefreshTree(t *testing.T) {
	dir := t.TempDir()
	m := app.New(dir)

	// Create file externally
	os.WriteFile(filepath.Join(dir, "external.md"), []byte("# ext"), 0644)

	// Tree initially doesn't have it
	flat0 := m.FlatListForTest()
	hasExternal := false
	for _, n := range flat0 {
		if strings.Contains(n, "external.md") {
			hasExternal = true
		}
	}
	if hasExternal {
		t.Skip("test setup error: file already visible")
	}

	// F5 refresh
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyF5})
	mm := updated.(app.Model)

	flat := mm.FlatListForTest()
	hasExternal = false
	for _, n := range flat {
		if strings.Contains(n, "external.md") {
			hasExternal = true
		}
	}
	if !hasExternal {
		t.Errorf("expected external.md to appear after F5 refresh, flatList=%v", flat)
	}
}

// sharedKey erzeugt eine tea.KeyMsg aus runes.
func sharedKey(runes ...rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: runes}
}
