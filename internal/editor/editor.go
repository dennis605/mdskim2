// Package editor implementiert einen Text-Buffer mit Cursor,
// Undo/Redo-History und Save.
package editor

import (
	"strings"
	"unicode/utf8"

	"github.com/dennis605/mdskim2/internal/workspace"
)

// Buffer hält den Text-Inhalt + Cursor + History.
type Buffer struct {
	Lines     []string
	CursorRow int
	CursorCol int
	Modified  bool
	Path      string

	// Undo/Redo: History ist eine Liste aller Zustände (inkl. aktueller).
	History    [][]string
	HistoryIdx int
	HistoryMax int
}

// LoadFromFile liest eine Datei in den Buffer.
func LoadFromFile(path string) (*Buffer, error) {
	if info, err := workspace.StatFile(path); err == nil && info.IsDir() {
		return nil, &NotAFileError{Path: path}
	}

	content, err := workspace.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(content, "\n")
	b := &Buffer{
		Lines:      lines,
		History:    [][]string{copyLines(lines)},
		HistoryIdx: 0,
		HistoryMax: 50,
		Path:       path,
	}
	return b, nil
}

// NewEmpty erzeugt einen leeren Buffer.
func NewEmpty() *Buffer {
	return &Buffer{
		Lines:      []string{""},
		History:    [][]string{{""}},
		HistoryIdx: 0,
		HistoryMax: 50,
	}
}

// NotAFileError signalisiert, dass versucht wurde, ein Verzeichnis zu öffnen.
type NotAFileError struct {
	Path string
}

func (e *NotAFileError) Error() string {
	return "ist ein Verzeichnis: " + e.Path
}

// TotalLines gibt die Anzahl Zeilen zurück.
func (b *Buffer) TotalLines() int {
	return len(b.Lines)
}

// TotalWords zählt die Wörter im Buffer.
func (b *Buffer) TotalWords() int {
	n := 0
	for _, line := range b.Lines {
		fields := strings.Fields(line)
		n += len(fields)
	}
	return n
}

// snapshotHistory speichert aktuelle Lines in History (nach der Änderung).
func (b *Buffer) SnapshotHistory() {
	// Wenn wir nach Undo schreiben, verwerfen spätere History
	if b.HistoryIdx >= 0 && b.HistoryIdx < len(b.History)-1 {
		b.History = b.History[:b.HistoryIdx+1]
	}
	cp := copyLines(b.Lines)
	b.History = append(b.History, cp)
	if len(b.History) > b.HistoryMax {
		b.History = b.History[len(b.History)-b.HistoryMax:]
	}
	b.HistoryIdx = len(b.History) - 1
}

func copyLines(src []string) []string {
	cp := make([]string, len(src))
	copy(cp, src)
	return cp
}

// Undo macht einen Schritt rückgängig.
func (b *Buffer) Undo() {
	if b.HistoryIdx <= 0 || len(b.History) == 0 {
		return
	}
	b.HistoryIdx--
	b.Lines = copyLines(b.History[b.HistoryIdx])
	b.Modified = true
	b.clampCursor()
}

// Redo stellt einen rückgängig gemachten Schritt wieder her.
func (b *Buffer) Redo() {
	if b.HistoryIdx >= len(b.History)-1 || len(b.History) == 0 {
		return
	}
	b.HistoryIdx++
	b.Lines = copyLines(b.History[b.HistoryIdx])
	b.Modified = true
	b.clampCursor()
}

func (b *Buffer) clampCursor() {
	if b.CursorRow >= len(b.Lines) {
		b.CursorRow = len(b.Lines) - 1
	}
	if b.CursorRow < 0 {
		b.CursorRow = 0
	}
	maxCol := utf8.RuneCountInString(b.Lines[b.CursorRow])
	if b.CursorCol > maxCol {
		b.CursorCol = maxCol
	}
}

// MoveCursor bewegt den Cursor (clamped).
func (b *Buffer) MoveCursor(dRow, dCol int) {
	newRow := b.CursorRow + dRow
	if newRow < 0 {
		newRow = 0
	}
	if newRow >= len(b.Lines) {
		newRow = len(b.Lines) - 1
	}
	if newRow < 0 {
		newRow = 0
	}
	b.CursorRow = newRow

	newCol := b.CursorCol + dCol
	if newCol < 0 {
		newCol = 0
	}
	if newCol > utf8.RuneCountInString(b.Lines[newRow]) {
		newCol = utf8.RuneCountInString(b.Lines[newRow])
	}
	b.CursorCol = newCol
}

// Home bewegt den Cursor an den Zeilenanfang.
func (b *Buffer) Home() {
	b.CursorCol = 0
}

// End bewegt den Cursor ans Zeilenende.
func (b *Buffer) End() {
	b.CursorCol = utf8.RuneCountInString(b.Lines[b.CursorRow])
}

// InsertChar fügt ein Zeichen an Cursor-Position ein.
func (b *Buffer) InsertChar(ch rune) {
	line := b.Lines[b.CursorRow]
	runes := []rune(line)
	if b.CursorCol > len(runes) {
		b.CursorCol = len(runes)
	}
	runes = append(runes[:b.CursorCol], append([]rune{ch}, runes[b.CursorCol:]...)...)
	b.Lines[b.CursorRow] = string(runes)
	b.CursorCol++
	b.Modified = true
	b.SnapshotHistory()
}

// DeleteChar löscht das Zeichen vor dem Cursor (Backspace).
func (b *Buffer) DeleteChar() {
	if b.CursorCol == 0 && b.CursorRow == 0 {
		return
	}
	if b.CursorCol == 0 {
		prev := b.Lines[b.CursorRow-1]
		cur := b.Lines[b.CursorRow]
		b.Lines = append(b.Lines[:b.CursorRow-1], append([]string{prev + cur}, b.Lines[b.CursorRow+1:]...)...)
		b.CursorRow--
		b.CursorCol = utf8.RuneCountInString(prev)
		b.Modified = true
		b.SnapshotHistory()
		return
	}
	line := b.Lines[b.CursorRow]
	runes := []rune(line)
	if b.CursorCol > len(runes) {
		b.CursorCol = len(runes)
	}
	runes = append(runes[:b.CursorCol-1], runes[b.CursorCol:]...)
	b.Lines[b.CursorRow] = string(runes)
	b.CursorCol--
	b.Modified = true
	b.SnapshotHistory()
}

// DeleteCharForward löscht das Zeichen an Cursor (Delete-Key).
func (b *Buffer) DeleteCharForward() {
	line := b.Lines[b.CursorRow]
	runes := []rune(line)
	if b.CursorCol >= len(runes) {
		if b.CursorRow < len(b.Lines)-1 {
			b.Lines[b.CursorRow] = line + b.Lines[b.CursorRow+1]
			b.Lines = append(b.Lines[:b.CursorRow+1], b.Lines[b.CursorRow+2:]...)
			b.Modified = true
			b.SnapshotHistory()
		}
		return
	}
	runes = append(runes[:b.CursorCol], runes[b.CursorCol+1:]...)
	b.Lines[b.CursorRow] = string(runes)
	b.Modified = true
	b.SnapshotHistory()
}

// InsertNewLine fügt einen Newline an Cursor-Position ein.
func (b *Buffer) InsertNewLine() {
	line := b.Lines[b.CursorRow]
	runes := []rune(line)
	if b.CursorCol > len(runes) {
		b.CursorCol = len(runes)
	}
	left := string(runes[:b.CursorCol])
	right := string(runes[b.CursorCol:])
	b.Lines[b.CursorRow] = left
	b.Lines = append(b.Lines[:b.CursorRow+1], append([]string{right}, b.Lines[b.CursorRow+1:]...)...)
	b.CursorRow++
	b.CursorCol = 0
	b.Modified = true
	b.SnapshotHistory()
}

// InsertString fügt einen String an Cursor-Position ein.
func (b *Buffer) InsertString(s string) {
	if s == "" {
		return
	}
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if i > 0 {
			b.InsertNewLine()
		}
		for _, ch := range line {
			b.InsertChar(ch)
		}
	}
}

// DeleteLine löscht die aktuelle Zeile komplett.
func (b *Buffer) DeleteLine() {
	if len(b.Lines) <= 1 {
		b.Lines[0] = ""
		b.CursorRow = 0
		b.CursorCol = 0
		b.Modified = true
		b.SnapshotHistory()
		return
	}
	b.Lines = append(b.Lines[:b.CursorRow], b.Lines[b.CursorRow+1:]...)
	if b.CursorRow >= len(b.Lines) {
		b.CursorRow = len(b.Lines) - 1
	}
	b.CursorCol = 0
	b.Modified = true
	b.SnapshotHistory()
}

// ToString serialisiert den Buffer zurück zu einem String.
func (b *Buffer) ToString() string {
	return strings.Join(b.Lines, "\n")
}

// Save schreibt den Buffer auf Disk.
func (b *Buffer) Save() error {
	if b.Path == "" {
		return &NoPathError{}
	}
	content := b.ToString()
	if err := workspace.WriteFile(b.Path, content); err != nil {
		return err
	}
	b.Modified = false
	b.SnapshotHistory()
	return nil
}

// NoPathError signalisiert, dass der Buffer keinen Pfad hat.
type NoPathError struct{}

func (e *NoPathError) Error() string {
	return "Buffer hat keinen Pfad"
}
