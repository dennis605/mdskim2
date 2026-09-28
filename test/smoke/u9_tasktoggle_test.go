package smoke

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dennis605/mdskim2/internal/app"
)

func TestU9_TaskToggleUncheckedToChecked(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "tasks.md")
	body := "# Tasks\n\n- [ ] Buy milk\n- [ ] Walk dog\n- regular line\n"
	if err := os.WriteFile(target, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	m := app.New(dir)
	mm := m
	mm.LoadFileForTest(target)
	buf := mm.BufferForTest()
	buf.CursorRow = 2
	buf.CursorCol = 0
	mm.Update(app.TaskToggleTriggerMsg{})
	if !strings.Contains(buf.Lines[2], "- [x] Buy milk") {
		t.Errorf("expected line to be checked, got %q", buf.Lines[2])
	}
}

func TestU9_TaskToggleCheckedToUnchecked(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "tasks.md")
	body := "# Tasks\n\n- [x] Done task\n"
	if err := os.WriteFile(target, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	m := app.New(dir)
	mm := m
	mm.LoadFileForTest(target)
	buf := mm.BufferForTest()
	buf.CursorRow = 2
	buf.CursorCol = 0
	mm.Update(app.TaskToggleTriggerMsg{})
	if !strings.Contains(buf.Lines[2], "- [ ] Done task") {
		t.Errorf("expected line to be unchecked, got %q", buf.Lines[2])
	}
}

func TestU9_TaskToggleIgnoresRegularLines(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "tasks.md")
	body := "# Tasks\n\n- regular list\nregular paragraph\n"
	if err := os.WriteFile(target, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	m := app.New(dir)
	mm := m
	mm.LoadFileForTest(target)
	buf := mm.BufferForTest()
	originalLine := buf.Lines[2]
	buf.CursorRow = 2
	buf.CursorCol = 0
	mm.Update(app.TaskToggleTriggerMsg{})
	if buf.Lines[2] != originalLine {
		t.Errorf("expected regular line unchanged, got %q (was %q)", buf.Lines[2], originalLine)
	}
}

func TestU9_TaskToggleWithStar(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "tasks.md")
	body := "* [ ] Starred task\n"
	if err := os.WriteFile(target, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	m := app.New(dir)
	mm := m
	mm.LoadFileForTest(target)
	buf := mm.BufferForTest()
	buf.CursorRow = 0
	buf.CursorCol = 0
	mm.Update(app.TaskToggleTriggerMsg{})
	if !strings.Contains(buf.Lines[0], "* [x] Starred task") {
		t.Errorf("expected starred task to toggle, got %q", buf.Lines[0])
	}
}
