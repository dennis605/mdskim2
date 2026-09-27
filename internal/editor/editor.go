// Package editor implementiert einen einfachen Text-Buffer mit
// Cursor-Position. Markdown-Syntax-Highlighting kommt in R5.
package editor

import (
	"strings"
	"unicode/utf8"

	"github.com/dennis605/mdskim2/internal/workspace"
)

// Buffer hält den Text-Inhalt + Cursor.
type Buffer struct {
	Lines     []string
	CursorRow int
	CursorCol int
	Modified  bool
	Path      string // leer = unbenannt
}

// LoadFromFile liest eine Datei in den Buffer.
func LoadFromFile(path string) (*Buffer, error) {
	// Verzeichnis ist kein File
	if info, err := workspace.StatFile(path); err == nil && info.IsDir() {
		return nil, &NotAFileError{Path: path}
	}

	content, err := workspace.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(content, "\n")
	return &Buffer{
		Lines:     lines,
		CursorRow: 0,
		CursorCol: 0,
		Modified:  false,
		Path:      path,
	}, nil
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
}

// DeleteChar löscht das Zeichen vor dem Cursor (Backspace).
func (b *Buffer) DeleteChar() {
	if b.CursorCol == 0 && b.CursorRow == 0 {
		return
	}
	if b.CursorCol == 0 {
		// Merge mit vorheriger Zeile
		prev := b.Lines[b.CursorRow-1]
		cur := b.Lines[b.CursorRow]
		b.Lines = append(b.Lines[:b.CursorRow-1], append([]string{prev + cur}, b.Lines[b.CursorRow+1:]...)...)
		b.CursorRow--
		b.CursorCol = utf8.RuneCountInString(prev)
		b.Modified = true
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
}

// DeleteCharForward löscht das Zeichen an Cursor (Delete-Key).
func (b *Buffer) DeleteCharForward() {
	line := b.Lines[b.CursorRow]
	runes := []rune(line)
	if b.CursorCol >= len(runes) {
		// Merge mit nächster Zeile
		if b.CursorRow < len(b.Lines)-1 {
			b.Lines[b.CursorRow] = line + b.Lines[b.CursorRow+1]
			b.Lines = append(b.Lines[:b.CursorRow+1], b.Lines[b.CursorRow+2:]...)
			b.Modified = true
		}
		return
	}
	runes = append(runes[:b.CursorCol], runes[b.CursorCol+1:]...)
	b.Lines[b.CursorRow] = string(runes)
	b.Modified = true
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
}

// ToString serialisiert den Buffer zurück zu einem String.
func (b *Buffer) ToString() string {
	return strings.Join(b.Lines, "\n")
}
