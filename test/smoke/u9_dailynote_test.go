package smoke

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dennis605/mdskim2/internal/app"
)

func TestU9_DailyNoteCreatesFile(t *testing.T) {
	dir := t.TempDir()
	m := app.New(dir)
	updated, _ := m.Update(app.DailyNoteTriggerMsg{})
	_ = updated
	notesDir := filepath.Join(dir, "Daily Notes")
	entries, err := os.ReadDir(notesDir)
	if err != nil {
		t.Fatalf("expected Daily Notes dir to be created, got err=%v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected exactly 1 file in Daily Notes, got %d", len(entries))
	}
	today := time.Now().Format("2006-01-02")
	if !strings.HasSuffix(entries[0].Name(), today+".md") {
		t.Errorf("expected file %s.md, got %s", today, entries[0].Name())
	}
}

func TestU9_DailyNoteIsLoaded(t *testing.T) {
	dir := t.TempDir()
	m := app.New(dir)
	updated, _ := m.Update(app.DailyNoteTriggerMsg{})
	mm := updated.(app.Model)
	if mm.CurrentFileForTest() == "" {
		t.Errorf("expected file to be loaded, got empty currentFile")
	}
	if !strings.Contains(mm.CurrentFileForTest(), "Daily Notes") {
		t.Errorf("expected loaded file in Daily Notes dir, got %q", mm.CurrentFileForTest())
	}
}

func TestU9_DailyNoteOpensExisting(t *testing.T) {
	dir := t.TempDir()
	// Pre-create a daily note for today
	notesDir := filepath.Join(dir, "Daily Notes")
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		t.Fatal(err)
	}
	today := time.Now().Format("2006-01-02")
	preset := filepath.Join(notesDir, today+".md")
	if err := os.WriteFile(preset, []byte("# Existing\n"), 0644); err != nil {
		t.Fatal(err)
	}
	m := app.New(dir)
	updated, _ := m.Update(app.DailyNoteTriggerMsg{})
	mm := updated.(app.Model)
	if mm.CurrentFileForTest() != preset {
		t.Errorf("expected loaded file %q, got %q", preset, mm.CurrentFileForTest())
	}
}
