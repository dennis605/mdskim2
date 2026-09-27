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

	// U3: Selection (anchor..head sortiert)
	SelAnchorRow int // Selection-Anker (oder -1 wenn keine Selection)
	SelAnchorCol int
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
		Lines:       []string{""},
		History:     [][]string{{""}},
		HistoryIdx:  0,
		HistoryMax:  50,
		SelAnchorRow: -1,
		SelAnchorCol: -1,
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
	if b.HasSelection() {
		b.DeleteSelection()
	}
	if len(b.Lines) == 0 {
		b.Lines = []string{""}
	}
	if b.CursorRow >= len(b.Lines) {
		b.CursorRow = len(b.Lines) - 1
	}
	line := b.Lines[b.CursorRow]
	runes := []rune(line)
	if b.CursorCol > len(runes) {
		b.CursorCol = len(runes)
	}
	left := string(runes[:b.CursorCol])
	right := string(runes[b.CursorCol:])
	// Auto-Indent: leading whitespace of current line (inkl. Tab)
	indent := ""
	for _, r := range runes {
		if r == ' ' || r == '\t' {
			indent += string(r)
		} else {
			break
		}
	}
	b.Lines[b.CursorRow] = left
	newLineText := indent + right
	b.Lines = append(b.Lines[:b.CursorRow+1], append([]string{newLineText}, b.Lines[b.CursorRow+1:]...)...)
	b.CursorRow++
	b.CursorCol = utf8.RuneCountInString(indent)
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


// GoToLine bewegt den Cursor zu einer bestimmten Zeile (1-indexed).
func (b *Buffer) GoToLine(line1 int) {
	if len(b.Lines) == 0 {
		return
	}
	if line1 < 1 {
		line1 = 1
	}
	if line1 > len(b.Lines) {
		line1 = len(b.Lines)
	}
	b.CursorRow = line1 - 1
	if b.CursorRow >= len(b.Lines) {
		b.CursorRow = len(b.Lines) - 1
	}
	b.CursorCol = 0
}

// HasSelection reports whether there's an active selection.
func (b *Buffer) HasSelection() bool {
	if b.SelAnchorRow < 0 {
		return false
	}
	if b.SelAnchorRow == b.CursorRow && b.SelAnchorCol == b.CursorCol {
		return false
	}
	return true
}

// ClearSelection löscht die Selection.
func (b *Buffer) ClearSelection() {
	b.SelAnchorRow = -1
	b.SelAnchorCol = -1
}

// SelectionRange returns (startRow, startCol, endRow, endCol) normalisiert (start <= end).
func (b *Buffer) SelectionRange() (sr, sc, er, ec int, hasSel bool) {
	if !b.HasSelection() {
		return 0, 0, 0, 0, false
	}
	if b.SelAnchorRow > b.CursorRow || (b.SelAnchorRow == b.CursorRow && b.SelAnchorCol > b.CursorCol) {
		return b.CursorRow, b.CursorCol, b.SelAnchorRow, b.SelAnchorCol, true
	}
	return b.SelAnchorRow, b.SelAnchorCol, b.CursorRow, b.CursorCol, true
}

// SelectedText returns den aktuell markierten Text.
func (b *Buffer) SelectedText() string {
	sr, sc, er, ec, has := b.SelectionRange()
	if !has || (sr == er && sc == ec) {
		return ""
	}
	var b2 strings.Builder
	if sr == er {
		// Single-line
		r := []rune(b.Lines[sr])
		if ec > len(r) {
			ec = len(r)
		}
		b2.WriteString(string(r[sc:ec]))
	} else {
		// Multi-line
		b2.WriteString(b.Lines[sr][utf8.RuneCountInString(b.Lines[sr][:sc*0]):])
		// Hmm — need to use rune-based. Just rebuild:
		srcRunes := []rune(b.Lines[sr])
		if sc > len(srcRunes) {
			sc = len(srcRunes)
		}
		b2.Reset()
		b2.WriteString(string(srcRunes[sc:]))
		for i := sr + 1; i < er; i++ {
			b2.WriteByte('\n')
			b2.WriteString(b.Lines[i])
		}
		b2.WriteByte('\n')
		destRunes := []rune(b.Lines[er])
		if ec > len(destRunes) {
			ec = len(destRunes)
		}
		b2.WriteString(string(destRunes[:ec]))
	}
	return b2.String()
}

// DeleteSelection entfernt die markierte Region und positioniert den Cursor an deren Start.
func (b *Buffer) DeleteSelection() {
	sr, sc, er, ec, has := b.SelectionRange()
	if !has {
		return
	}
	if len(b.Lines) == 0 {
		return
	}
	if sr >= len(b.Lines) {
		return
	}
	if er >= len(b.Lines) {
		er = len(b.Lines) - 1
	}
	srcR := []rune(b.Lines[sr])
	if sc > len(srcR) {
		sc = len(srcR)
	}
	if sc < 0 {
		sc = 0
	}
	head := string(srcR[:sc])
	if sr == er {
		destR := srcR
		if ec > len(destR) {
			ec = len(destR)
		}
		b.Lines[sr] = head + string(destR[ec:])
		b.CursorRow = sr
		b.CursorCol = sc
	} else {
		destLine := b.Lines[er]
		destR := []rune(destLine)
		if ec > len(destR) {
			ec = len(destR)
		}
		tailRest := string(destR[ec:])
		// Compose: head + (remainder of original lines as separate entries)
		newLines := []string{head + tailRest}
		// Append the lines after er as separate lines (keep them as they were)
		for i := er + 1; i < len(b.Lines); i++ {
			newLines = append(newLines, b.Lines[i])
		}
		b.Lines = append(b.Lines[:sr], newLines...)
		if len(b.Lines) == 0 {
			b.Lines = []string{""}
		}
		b.CursorRow = sr
		b.CursorCol = sc
	}
	b.ClearSelection()
	b.Modified = true
	b.SnapshotHistory()
}

// SetCursorWithSelection bewegt den Cursor und setzt/löscht Selection.
func (b *Buffer) SetCursorWithSelection(row, col int, extend bool) {
	if !extend {
		b.ClearSelection()
	} else if !b.HasSelection() {
		b.SelAnchorRow = b.CursorRow
		b.SelAnchorCol = b.CursorCol
	}
	// Clamp row/col
	if row < 0 {
		row = 0
	}
	if row >= len(b.Lines) {
		row = len(b.Lines) - 1
	}
	if row < 0 {
		return
	}
	lineLen := utf8.RuneCountInString(b.Lines[row])
	if col < 0 {
		col = 0
	}
	if col > lineLen {
		col = lineLen
	}
	b.CursorRow = row
	b.CursorCol = col
}

// MoveCursorExt setzt Cursor wie MoveCursor aber mit optionaler Selection-Extension.
func (b *Buffer) MoveCursorExt(dRow, dCol int, extend bool) {
	if !extend && b.HasSelection() {
		// Bei normalem Pfeil: Selection auf Cursorposition kollabieren
		b.ClearSelection()
	}
	b.MoveCursor(dRow, dCol)
	if extend && b.HasSelection() {
		// Anchored; cursor moved already
	}
}

// InsertCharWithSelection löscht zuerst Selection, dann insert.
func (b *Buffer) InsertCharWithSelection(r rune) {
	if b.HasSelection() {
		b.DeleteSelection()
	}
	b.InsertChar(r)
}

// WordLeft bewegt den Cursor wortweise nach links.
func (b *Buffer) WordLeft() {
	if len(b.Lines) == 0 {
		return
	}
	r := b.CursorRow
	c := b.CursorCol
	if r < 0 {
		r = 0
	}
	if r >= len(b.Lines) {
		r = len(b.Lines) - 1
	}
	lineLen := utf8.RuneCountInString(b.Lines[r])
	if c > lineLen {
		c = lineLen
	}
	// More than one step left in current line: scan within line
	for r >= 0 {
		line := []rune(b.Lines[r])
		if c == 0 {
			// At BOL — try move to previous line
			if r > 0 {
				r--
				line = []rune(b.Lines[r])
				c = len(line)
				// Could repeat if previous line was empty: try again
				if c == 0 {
					continue
				}
				// Skip past trailing whitespace, then past word
				for c > 0 && isWordSep(line[c-1]) {
					c--
				}
				for c > 0 && !isWordSep(line[c-1]) {
					c--
				}
				b.CursorRow = r
				b.CursorCol = c
				return
			}
			b.CursorRow = r
			return
		}
		// Scan left within current line
		startC := c
		for c > 0 && isWordSep(line[c-1]) {
			c--
		}
		if c > 0 {
			for c > 0 && !isWordSep(line[c-1]) {
				c--
			}
		}
		if c < startC {
			b.CursorRow = r
			b.CursorCol = c
			return
		}
		// Couldn't move left in this line — go to EOL of previous
		if r > 0 {
			r--
			c = len([]rune(b.Lines[r]))
			continue
		}
		b.CursorRow = 0
		b.CursorCol = 0
		return
	}
}

// WordRight bewegt den Cursor wortweise nach rechts.
func (b *Buffer) WordRight() {
	if len(b.Lines) == 0 {
		return
	}
	r := b.CursorRow
	c := b.CursorCol
	for r < len(b.Lines) {
		line := []rune(b.Lines[r])
		if c >= len(line) && r+1 < len(b.Lines) {
			r++
			c = 0
			line = []rune(b.Lines[r])
			for c < len(line) && isWordSep(line[c]) {
				c++
			}
			b.CursorRow = r
			b.CursorCol = c
			return
		}
		if c >= len(line) {
			b.CursorRow = r
			b.CursorCol = len(line)
			return
		}
		startC := c
		// If at start of word, skip past word; if at whitespace, skip past whitespace
		if c < len(line) && isWordSep(line[c]) {
			// skip whitespace
			for c < len(line) && isWordSep(line[c]) {
				c++
			}
		} else if c < len(line) {
			// skip word
			for c < len(line) && !isWordSep(line[c]) {
				c++
			}
		}
		if c > startC {
			b.CursorRow = r
			b.CursorCol = c
			return
		}
	b.CursorRow = r
	b.CursorCol = c
}
}

// isWordSep returns true if rune is whitespace or punctuation.
func isWordSep(r rune) bool {
	if r == ' ' || r == '\t' || r == '\n' {
		return true
	}
	if r >= 'a' && r <= 'z' {
		return false
	}
	if r >= 'A' && r <= 'Z' {
		return false
	}
	if r >= '0' && r <= '9' {
		return false
	}
	return r == '_'
}

// DuplicateLine dupliziert die aktuelle Zeile.
func (b *Buffer) DuplicateLine() {
	if b.CursorRow < 0 || b.CursorRow >= len(b.Lines) {
		return
	}
	line := b.Lines[b.CursorRow]
	b.Lines = append(b.Lines[:b.CursorRow+1], append([]string{line}, b.Lines[b.CursorRow+1:]...)...)
	b.CursorRow++
	b.Modified = true
	b.SnapshotHistory()
}

// MoveLineUp bewegt die aktuelle Zeile eine Position nach oben.
func (b *Buffer) MoveLineUp() {
	if b.CursorRow <= 0 || b.CursorRow >= len(b.Lines) {
		return
	}
	b.Lines[b.CursorRow], b.Lines[b.CursorRow-1] = b.Lines[b.CursorRow-1], b.Lines[b.CursorRow]
	b.CursorRow--
	b.Modified = true
	b.SnapshotHistory()
}

// MoveLineDown bewegt die aktuelle Zeile eine Position nach unten.
func (b *Buffer) MoveLineDown() {
	if b.CursorRow < 0 || b.CursorRow >= len(b.Lines)-1 {
		return
	}
	b.Lines[b.CursorRow], b.Lines[b.CursorRow+1] = b.Lines[b.CursorRow+1], b.Lines[b.CursorRow]
	b.CursorRow++
	b.Modified = true
	b.SnapshotHistory()
}

// WordCount returns die Anzahl Wörter im Buffer.
func (b *Buffer) WordCount() int {
	n := 0
	for _, line := range b.Lines {
		inWord := false
		for _, r := range line {
			if r == ' ' || r == '\t' {
				inWord = false
			} else {
				if !inWord {
					n++
					inWord = true
				}
			}
		}
	}
	return n
}

// CharCount returns die Anzahl Zeichen (ohne Newlines).
func (b *Buffer) CharCount() int {
	n := 0
	for _, line := range b.Lines {
		n += utf8.RuneCountInString(line)
	}
	return n
}

// LineIndent returns die Einrückung (Whitespace-Prefix) der Zeile row.
func (b *Buffer) LineIndent(row int) string {
	if row < 0 || row >= len(b.Lines) {
		return ""
	}
	line := b.Lines[row]
	for i, r := range line {
		if r != ' ' && r != '\t' {
			return line[:i]
		}
	}
	return line
}