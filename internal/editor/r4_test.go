package editor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUndoRedo(t *testing.T) {
	buf := NewEmpty()
	buf.Lines = []string{"hello"}
	buf.CursorRow = 0
	buf.CursorCol = 5
	buf.SnapshotHistory() // [hello] als initial
	buf.InsertChar('!')
	if buf.Lines[0] != "hello!" {
		t.Errorf("nach Insert: %q, erwartet 'hello!'", buf.Lines[0])
	}

	buf.Undo()
	if buf.Lines[0] != "hello" {
		t.Errorf("nach Undo: %q, erwartet 'hello'", buf.Lines[0])
	}

	buf.Redo()
	if buf.Lines[0] != "hello!" {
		t.Errorf("nach Redo: %q, erwartet 'hello!'", buf.Lines[0])
	}
}

func TestSave(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.md")

	buf := NewEmpty()
	buf.Lines = []string{"# Test", "", "Content"}
	buf.Path = path
	if err := buf.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if buf.Modified {
		t.Error("Modified sollte false sein nach Save")
	}

	// Datei prüfen
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "# Test\n\nContent" {
		t.Errorf("File-Inhalt falsch: %q", string(content))
	}
}

func TestInsertString(t *testing.T) {
	buf := NewEmpty()
	buf.InsertString("Line1\nLine2")
	if buf.TotalLines() != 2 {
		t.Errorf("erwartet 2 lines, got %d", buf.TotalLines())
	}
	if buf.Lines[0] != "Line1" || buf.Lines[1] != "Line2" {
		t.Errorf("InsertString falsch: %v", buf.Lines)
	}
}

func TestDeleteLine(t *testing.T) {
	buf := &Buffer{
		Lines:      []string{"A", "B", "C"},
		CursorRow:  1,
		HistoryIdx: -1,
		HistoryMax: 50,
	}
	buf.DeleteLine()
	if buf.TotalLines() != 2 {
		t.Errorf("erwartet 2 lines, got %d", buf.TotalLines())
	}
	if buf.Lines[1] != "C" {
		t.Errorf("erwartet 'C' in line 1, got %q", buf.Lines[1])
	}
}

func TestUndoVerwerftRedoHistory(t *testing.T) {
	buf := NewEmpty()
	buf.Lines = []string{"a"}
	buf.CursorRow = 0
	buf.CursorCol = 1
	buf.SnapshotHistory() // [a] als initial
	buf.InsertChar('b') // History: [a, ab]
	buf.InsertChar('c') // History: [a, ab, abc]
	buf.Undo()          // History: [a, ab], Lines: ab
	buf.InsertChar('X') // History: [a, ab, abX] — Redo-Pfad verworfen
	if buf.Lines[0] != "abX" {
		t.Errorf("erwartet 'abX', got %q", buf.Lines[0])
	}
	buf.Redo() // sollte nichts tun
	if buf.Lines[0] != "abX" {
		t.Errorf("Redo nach Insert sollte 'abX' lassen, got %q", buf.Lines[0])
	}
}
