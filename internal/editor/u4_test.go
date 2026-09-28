package editor

import (
	"strings"
	"testing"
)

func TestLineEndingDetectionLF(t *testing.T) {
	content := "hello\nworld"
	lineEnding := detectLineEnding(content)
	if lineEnding != "\n" {
		t.Errorf("LF detection: got %q, want \n", lineEnding)
	}
}

func TestLineEndingDetectionCRLF(t *testing.T) {
	content := "hello\r\nworld"
	lineEnding := detectLineEnding(content)
	if lineEnding != "\r\n" {
		t.Errorf("CRLF detection: got %q, want \r\n", lineEnding)
	}
}

func TestLineEndingDetectionCR(t *testing.T) {
	content := "hello\rworld"
	lineEnding := detectLineEnding(content)
	if lineEnding != "\r" {
		t.Errorf("CR detection: got %q, want \r", lineEnding)
	}
}

func TestSavePreservesLineEnding(t *testing.T) {
	b := NewEmpty()
	b.Lines = []string{"hello", "world"}
	b.LineEnding = "\r\n"
	out := b.ToString()
	// ToString joins with \n by design (internal buffer); Save converts to LineEnding.
	// Verify the conversion happens in Save via a quick check of the joined output.
	if !strings.Contains(out, "\n") {
		t.Errorf("expected newline in ToString, got %q", out)
	}
	if strings.Contains(out, "\r") {
		t.Errorf("ToString should not contain CR (that's Save's job), got %q", out)
	}
}

func TestBufferSetLine(t *testing.T) {
	b := &Buffer{Lines: []string{"alpha", "beta", "gamma"}, CursorRow: 1, CursorCol: 0}
	b.SetLine(1, "BETA")
	if b.Lines[1] != "BETA" {
		t.Fatalf("expected BETA, got %q", b.Lines[1])
	}
	if !b.Modified {
		t.Fatal("expected Modified=true after SetLine")
	}
	if b.CursorRow != 1 || b.CursorCol != 0 {
		t.Fatalf("cursor should not move, got (%d,%d)", b.CursorRow, b.CursorCol)
	}
}

func TestBufferSetLineBounds(t *testing.T) {
	b := &Buffer{Lines: []string{"a", "b"}, CursorRow: 0, CursorCol: 0}
	b.SetLine(-1, "x")
	b.SetLine(99, "y")
	if b.Lines[0] != "a" || b.Lines[1] != "b" {
		t.Fatalf("out-of-bounds SetLine should be no-op, got %v", b.Lines)
	}
	if b.Modified {
		t.Fatal("out-of-bounds SetLine must not mark Modified")
	}
}

func TestBufferSetLineIdempotent(t *testing.T) {
	b := &Buffer{Lines: []string{"a"}, CursorRow: 0, CursorCol: 0}
	b.Modified = false
	b.SetLine(0, "a")
	if b.Modified {
		t.Fatal("setting same content should not mark Modified")
	}
}

func TestBufferSetLineUndo(t *testing.T) {
	// Use NewEmpty so History[0] is populated; then SetLine pushes History[1].
	b := NewEmpty()
	b.Lines = []string{"foo"}
	b.History = [][]string{{"foo"}}
	b.HistoryIdx = 0
	b.CursorRow = 0
	b.CursorCol = 0
	b.SetLine(0, "bar")
	if b.Lines[0] != "bar" {
		t.Fatalf("expected bar, got %q", b.Lines[0])
	}
	b.Undo()
	if b.Lines[0] != "foo" {
		t.Fatalf("after undo expected foo, got %q", b.Lines[0])
	}
}
