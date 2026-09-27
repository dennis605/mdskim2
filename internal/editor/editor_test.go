package editor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.md")
	content := "# Title\n\nPara 1\nPara 2\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	buf, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile: %v", err)
	}

	if buf.TotalLines() != 5 {
		t.Errorf("erwartet 4 lines, got %d", buf.TotalLines())
	}
	if buf.Lines[0] != "# Title" {
		t.Errorf("erwartet '# Title' in line 0, got %q", buf.Lines[0])
	}
	if buf.TotalWords() != 6 {
		t.Errorf("erwartet 6 words, got %d", buf.TotalWords())
	}
	if buf.Modified {
		t.Error("Buffer sollte nicht modified sein nach Load")
	}
}

func TestInsertChar(t *testing.T) {
	buf := &Buffer{
		Lines:     []string{"Helo"},
		CursorRow: 0,
		CursorCol: 2,
	}
	buf.InsertChar('l')
	if buf.Lines[0] != "Hello" {
		t.Errorf("nach InsertChar: %q, erwartet 'Hello'", buf.Lines[0])
	}
	if buf.CursorCol != 3 {
		t.Errorf("CursorCol sollte 3 sein, got %d", buf.CursorCol)
	}
	if !buf.Modified {
		t.Error("Modified sollte true sein")
	}
}

func TestDeleteChar(t *testing.T) {
	buf := &Buffer{
		Lines:     []string{"Hello"},
		CursorRow: 0,
		CursorCol: 5,
	}
	buf.DeleteChar()
	if buf.Lines[0] != "Hell" {
		t.Errorf("nach Backspace: %q, erwartet 'Hell'", buf.Lines[0])
	}
}

func TestNewLine(t *testing.T) {
	buf := &Buffer{
		Lines:     []string{"abc"},
		CursorRow: 0,
		CursorCol: 1,
	}
	buf.InsertNewLine()
	if buf.TotalLines() != 2 {
		t.Errorf("erwartet 2 lines, got %d", buf.TotalLines())
	}
	if buf.Lines[0] != "a" || buf.Lines[1] != "bc" {
		t.Errorf("Split falsch: %v", buf.Lines)
	}
}

func TestMoveCursorClamp(t *testing.T) {
	buf := &Buffer{
		Lines:     []string{"ab", "cd"},
		CursorRow: 0,
		CursorCol: 0,
	}
	buf.MoveCursor(-1, 0) // sollte clamp auf 0
	if buf.CursorRow != 0 {
		t.Errorf("Cursor sollte 0 sein, got %d", buf.CursorRow)
	}
	buf.MoveCursor(0, -1)
	if buf.CursorCol != 0 {
		t.Errorf("CursorCol sollte 0 sein, got %d", buf.CursorCol)
	}
	buf.MoveCursor(0, 99) // sollte clamp auf 2
	if buf.CursorCol != 2 {
		t.Errorf("CursorCol sollte 2 sein (geclamped), got %d", buf.CursorCol)
	}
}

func TestLoadFromDirectoryFails(t *testing.T) {
	tmp := t.TempDir()
	_, err := LoadFromFile(tmp)
	if err == nil {
		t.Error("erwartet Fehler bei LoadFromFile auf Verzeichnis")
	}
}
