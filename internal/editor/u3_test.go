package editor

import (
	"testing"
)

func TestSelectionBasic(t *testing.T) {
	b := NewEmpty()
	b.Lines = []string{"Hello World"}
	b.CursorCol = 5

	// Anchor at current (0,5), then extend left
	b.SelAnchorRow = 0
	b.SelAnchorCol = 5
	b.SetCursorWithSelection(0, 0, true)

	sr, sc, er, ec, has := b.SelectionRange()
	if !has {
		t.Fatal("expected selection")
	}
	if sr != 0 || sc != 0 || er != 0 || ec != 5 {
		t.Errorf("SelectionRange wrong: %d,%d → %d,%d", sr, sc, er, ec)
	}

	txt := b.SelectedText()
	if txt != "Hello" {
		t.Errorf("SelectedText = %q, want 'Hello'", txt)
	}
}

func TestSelectionReverse(t *testing.T) {
	b := NewEmpty()
	b.Lines = []string{"Hello World"}
	b.CursorCol = 5
	b.SetCursorWithSelection(0, 5, true)
	// Move left — anchor > cursor so range gets swapped
	b.SetCursorWithSelection(0, 0, true)

	sr, sc, er, ec, has := b.SelectionRange()
	if !has {
		t.Fatal("expected selection")
	}
	if sr != 0 || sc != 0 || er != 0 || ec != 5 {
		t.Errorf("expected normalized %d,%d → %d,%d, got %d,%d → %d,%d", 0, 0, 0, 5, sr, sc, er, ec)
	}
}

func TestDeleteSelectionSingleLine(t *testing.T) {
	b := NewEmpty()
	b.Lines = []string{"Hello World"}
	b.CursorCol = 6 // start at col 6 first, then extend selection
	b.SetCursorWithSelection(0, 0, true) // anchor goes to (0,6)... no
	// Reset properly: anchor at current cursor (6,0), then extend right to (11,0)
	b.SelAnchorRow = 0
	b.SelAnchorCol = 6
	b.SetCursorWithSelection(0, 11, true)
	b.DeleteSelection()
	if b.Lines[0] != "Hello " {
		t.Errorf("after DeleteSelection got %q, want 'Hello '", b.Lines[0])
	}
	if b.CursorRow != 0 || b.CursorCol != 6 {
		t.Errorf("cursor at %d,%d, want 0,6", b.CursorRow, b.CursorCol)
	}
}

func TestDeleteSelectionMultiLine(t *testing.T) {
	b := NewEmpty()
	b.Lines = []string{"abc", "def", "ghi"}
	// Move cursor to row 1, col 3, anchor at row 0 col 2, extend to (1,3)
	b.CursorRow = 1
	b.CursorCol = 3
	b.SelAnchorRow = 0
	b.SelAnchorCol = 2
	b.SetCursorWithSelection(1, 3, true)
	b.DeleteSelection()
	if len(b.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %v", len(b.Lines), b.Lines)
	}
	if b.Lines[0] != "ab" {
		t.Errorf("expected 'ab', got %q", b.Lines[0])
	}
	if b.Lines[1] != "ghi" {
		t.Errorf("expected 'ghi', got %q", b.Lines[1])
	}
}

func TestWordLeftRight(t *testing.T) {
	b := NewEmpty()
	b.Lines = []string{"hello world foo"}
	b.CursorCol = 16
	b.WordLeft()
	if b.CursorCol != 12 {
		t.Errorf("WordLeft from 16 → %d, want 12", b.CursorCol)
	}
	b.WordLeft()
	if b.CursorCol != 6 {
		t.Errorf("WordLeft from 12 → %d, want 6", b.CursorCol)
	}
	b.CursorCol = 0
	b.WordRight()
	if b.CursorCol != 5 {
		t.Errorf("WordRight from 0 → %d, want 5", b.CursorCol)
	}
}

func TestDuplicateLine(t *testing.T) {
	b := NewEmpty()
	b.Lines = []string{"a", "b", "c"}
	b.CursorRow = 1
	b.DuplicateLine()
	if len(b.Lines) != 4 || b.Lines[1] != "b" || b.Lines[2] != "b" {
		t.Errorf("got %v, want [a b b c]", b.Lines)
	}
	if b.CursorRow != 2 {
		t.Errorf("cursor at %d, want 2", b.CursorRow)
	}
}

func TestMoveLineUpDown(t *testing.T) {
	b := NewEmpty()
	b.Lines = []string{"a", "b", "c"}
	b.CursorRow = 1
	b.MoveLineDown()
	if b.Lines[1] != "c" || b.Lines[2] != "b" {
		t.Errorf("MoveLineDown got %v", b.Lines)
	}
	if b.CursorRow != 2 {
		t.Errorf("cursor at %d, want 2", b.CursorRow)
	}
	b.MoveLineUp()
	if b.Lines[1] != "b" || b.Lines[2] != "c" {
		t.Errorf("MoveLineUp got %v", b.Lines)
	}
}

func TestInsertNewLineAutoIndent(t *testing.T) {
	b := NewEmpty()
	b.Lines = []string{"    hello"}
	b.CursorCol = 4
	b.InsertNewLine()
	// Cursor mid-word: split "    |hello" → ["    ", "    hello"]
	if b.Lines[0] != "    " {
		t.Errorf("expected '    ' on line 0, got %q", b.Lines[0])
	}
	if b.Lines[1] != "    hello" {
		t.Errorf("expected auto-indented '    hello' on line 1, got %q", b.Lines[1])
	}
	if b.CursorCol != 4 {
		t.Errorf("CursorCol at %d, want 4", b.CursorCol)
	}
	if b.CursorRow != 1 {
		t.Errorf("CursorRow at %d, want 1", b.CursorRow)
	}
}

func TestGoToLine(t *testing.T) {
	b := NewEmpty()
	b.Lines = []string{"a", "b", "c", "d", "e"}
	b.GoToLine(3)
	if b.CursorRow != 2 {
		t.Errorf("GoToLine(3) → row %d, want 2", b.CursorRow)
	}
	b.GoToLine(100)
	if b.CursorRow != 4 {
		t.Errorf("GoToLine(100) clamped to row %d, want 4", b.CursorRow)
	}
	b.GoToLine(0)
	if b.CursorRow != 0 {
		t.Errorf("GoToLine(0) → row %d, want 0", b.CursorRow)
	}
}

func TestWordAndCharCount(t *testing.T) {
	b := NewEmpty()
	b.Lines = []string{"hello world", "foo bar"}
	if b.WordCount() != 4 {
		t.Errorf("WordCount = %d, want 4", b.WordCount())
	}
	if b.CharCount() != 18 { // "hello world" + "foo bar" = 11 + 7
		t.Errorf("CharCount = %d, want 18", b.CharCount())
	}
}

func TestLineIndent(t *testing.T) {
	b := NewEmpty()
	b.Lines = []string{"    foo", "bar"}
	if b.LineIndent(0) != "    " {
		t.Errorf("LineIndent(0) = %q, want '    '", b.LineIndent(0))
	}
	if b.LineIndent(1) != "" {
		t.Errorf("LineIndent(1) = %q, want ''", b.LineIndent(1))
	}
}
