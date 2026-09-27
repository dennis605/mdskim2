// Package app implementiert die mdskim2-Anwendung als Bubble-Tea-Model.
package app

import (
	"fmt"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/atotto/clipboard"

	"github.com/dennis605/mdskim2/internal/editor"
	"github.com/dennis605/mdskim2/internal/grep"
	"github.com/dennis605/mdskim2/internal/markdown"
	"github.com/dennis605/mdskim2/internal/preview"
	"github.com/dennis605/mdskim2/internal/search"
	"github.com/dennis605/mdskim2/internal/palette"
	"github.com/dennis605/mdskim2/internal/tabs"
	"github.com/dennis605/mdskim2/internal/ui"
	"github.com/dennis605/mdskim2/internal/workspace"
)

type Model struct {
	workspacePath string
	width         int
	height        int
	quitting      bool

	workspace  *workspace.Workspace
	treeRender *workspace.TreeRenderer
	cursorIdx  int
	flatList   []*workspace.FileNode

	buffer      *editor.Buffer
		clipboard   string
	selection   string
	mode        string
	currentFile string

	// R6: Search-Modal
	searchQuery    string
	searchResults  []search.Match
	searchIdx      int
	searchActive   bool

	// R7: Preview-Mode
	previewMode    bool   // wenn true, zeige Preview statt Editor-Buffer
	previewSplit   bool   // Split-View (Editor + Preview)
	previewCache   string // cached preview render

	// R8: Tabs + Quick-Open
	tabs           *tabs.Manager
	quickOpenMode  bool
	quickOpenQuery string

	theme   ui.Theme
	layout  ui.Layout
	version string

	// R9: Command Palette
	palette        *palette.Registry
	paletteActive  bool
	paletteQuery   string

	// R10: Workspace-Grep
	workspaceHits  []grep.Hit

	saveError string

	// U2 Focus-Modell: "tree" oder "editor"
	focus string

	// U3: Editor-Features
	lineWrap      bool
	goToLineMode  bool
	editorScrollOffset int
	goToLineQuery string
}

func New(workspacePath string) Model {
	ws, _ := workspace.Load(workspacePath)
	r := workspace.NewTreeRenderer()
	flat := ws.FlatList(r.CollapsedDirs)

	m := Model{
		workspacePath: workspacePath,
		width:         120,
		height:        40,
		mode:          "EDIT",
		focus:         "tree",
		version:       "R10: Polish + Stubs",
		theme:         ui.Light(),
		layout:        ui.DefaultLayout(),
		workspace:     ws,
		treeRender:    r,
		flatList:      flat,
		tabs:          tabs.NewManager(),
		palette:       palette.NewRegistry(),
	}

	// Register default commands
	m.palette.Register(palette.Command{Name: "save", Description: "Save current file", Keywords: []string{"write", "store", "ctrl+s"}})
	m.palette.Register(palette.Command{Name: "open", Description: "Open file", Keywords: []string{"load", "read"}})
	m.palette.Register(palette.Command{Name: "preview", Description: "Toggle preview", Keywords: []string{"render", "glamour"}})
	m.palette.Register(palette.Command{Name: "find", Description: "Open search modal", Keywords: []string{"search", "fuzzy"}})
	m.palette.Register(palette.Command{Name: "replace", Description: "Replace in file", Keywords: []string{"substitute"}})
	m.palette.Register(palette.Command{Name: "bold", Description: "Insert bold markup", Keywords: []string{"strong"}})
	m.palette.Register(palette.Command{Name: "italic", Description: "Insert italic markup", Keywords: []string{"em"}})
	m.palette.Register(palette.Command{Name: "quit", Description: "Quit application", Keywords: []string{"exit", "close"}})
	return m
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layout.Compute(m.width, m.height)
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.MouseMsg:
		return m.handleMouse(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlQ:
		m.quitting = true
		return m, tea.Quit
	case tea.KeyCtrlS:
		return m.saveBuffer()
	case tea.KeyCtrlZ:
		if m.buffer != nil {
			m.buffer.Undo()
		}
		return m, nil
	case tea.KeyCtrlY:
		if m.buffer != nil {
			m.buffer.Redo()
		}
		return m, nil
	case tea.KeyCtrlC:
		if m.buffer != nil {
			line := m.buffer.Lines[m.buffer.CursorRow]
			clipboard.WriteAll(line)
			m.clipboard = line
		}
		return m, nil
	case tea.KeyCtrlX:
		if m.buffer != nil {
			line := m.buffer.Lines[m.buffer.CursorRow]
			clipboard.WriteAll(line)
			m.clipboard = line
			m.buffer.DeleteLine()
		}
		return m, nil
	case tea.KeyCtrlP:
		// Toggle Preview-Mode
		if m.buffer != nil {
			m.previewMode = !m.previewMode
			if m.previewMode {
				rendered, err := preview.Render(m.buffer.ToString())
				if err == nil {
					m.previewCache = rendered
				}
			}
		}
		return m, nil

	case tea.KeyCtrlT:
		// Toggle Quick-Open
		m.quickOpenMode = !m.quickOpenMode
		if !m.quickOpenMode {
			m.quickOpenQuery = ""
		}
		return m, nil

	case tea.KeyCtrlK:
		// Toggle Command Palette
		m.paletteActive = !m.paletteActive
		if !m.paletteActive {
			m.paletteQuery = ""
		}
		return m, nil

	case tea.KeyCtrlD:
		// Zeile duplizieren
		if m.buffer != nil && m.focus == "editor" {
			m.buffer.DuplicateLine()
		}
		return m, nil

	case tea.KeyCtrlW:
		// Soft-Wrap Toggle
		m.lineWrap = !m.lineWrap
		return m, nil

	case tea.KeyCtrlG:
		// Go-to-line Modal
		m.goToLineMode = !m.goToLineMode
		if !m.goToLineMode {
			m.goToLineQuery = ""
		}
		return m, nil

	case tea.KeyCtrlV:
		if m.buffer != nil {
			text, err := clipboard.ReadAll()
			if err == nil && text != "" {
				m.buffer.InsertString(text)
				m.clipboard = text
			}
		}
		return m, nil
	case tea.KeyCtrlB:
		if m.buffer != nil {
			m.buffer.InsertString("****")
			m.buffer.MoveCursor(0, -2)
		}
		return m, nil
	case tea.KeyCtrlF:
		// Toggle Search-Modal
		m.searchActive = !m.searchActive
		if m.searchActive && m.buffer == nil {
			m.searchActive = false
		}
		if !m.searchActive {
			m.searchQuery = ""
			m.searchResults = nil
			m.searchIdx = 0
		}
		return m, nil

	case tea.KeyCtrlH:
		// Replace: einfach Ersetzen aller Vorkommen (ein-Schritt-Modal)
		if m.buffer == nil || m.searchQuery == "" {
			return m, nil
		}
		opt := search.Options{CaseSensitive: false}
		newLines, count := search.ReplaceAll(m.buffer.Lines, m.searchQuery, "REPLACED", opt)
		if count > 0 {
			m.buffer.Lines = newLines
			m.buffer.Modified = true
			m.buffer.SnapshotHistory()
		}
		m.mode = fmt.Sprintf("REPLACE: %d", count)
		return m, nil

	case tea.KeyCtrlI:
		if m.buffer != nil {
			m.buffer.InsertString("**")
			m.buffer.MoveCursor(0, -1)
		}
		return m, nil
		case tea.KeyEsc:
		// Esc priorisiert: Modals > Tree > Quit
		if m.searchActive {
			m.searchActive = false
			return m, nil
		}
		if m.quickOpenMode {
			m.quickOpenMode = false
			return m, nil
		}
		if m.paletteActive {
			m.paletteActive = false
			return m, nil
		}
		if m.goToLineMode {
			m.goToLineMode = false
			m.goToLineQuery = ""
			return m, nil
		}
		// Esc zurück zum Tree
		m.focus = "tree"
		return m, nil
	case tea.KeyUp:
		if m.focus == "editor" && m.buffer != nil {
			m.buffer.MoveCursor(-1, 0)
		} else {
			// Tree-Navigation: Pfeiltasten bleiben im Tree
			m.focus = "tree"
			if m.cursorIdx > 0 {
				m.cursorIdx--
				m.autoOpenAtCursor()
			}
		}
		return m, nil
	case tea.KeyDown:
		if m.focus == "editor" && m.buffer != nil {
			m.buffer.MoveCursor(1, 0)
		} else {
			// Tree-Navigation
			m.focus = "tree"
			if m.cursorIdx < len(m.flatList)-1 {
				m.cursorIdx++
				m.autoOpenAtCursor()
			}
		}
		return m, nil
	case tea.KeyLeft:
		if m.focus == "editor" && m.buffer != nil {
			m.buffer.MoveCursor(0, -1)
		} else if m.cursorIdx >= 0 && m.cursorIdx < len(m.flatList) {
			node := m.flatList[m.cursorIdx]
			if node.IsDir {
				// Im Tree: Links auf collapsed dir → expand, expanded → focus tree
				m.treeRender.ToggleDir(node.Path)
				m.flatList = m.workspace.FlatList(m.treeRender.CollapsedDirs)
			}
		}
		return m, nil
	case tea.KeyRight:
		if m.focus == "editor" && m.buffer != nil {
			m.buffer.MoveCursor(0, 1)
		} else if m.cursorIdx >= 0 && m.cursorIdx < len(m.flatList) {
			node := m.flatList[m.cursorIdx]
			if node.IsDir {
				// Im Tree: Rechts auf collapsed dir → expand
				m.treeRender.ToggleDir(node.Path)
				m.flatList = m.workspace.FlatList(m.treeRender.CollapsedDirs)
			}
		}
		return m, nil
	case tea.KeyHome:
		if m.focus == "editor" && m.buffer != nil {
			m.buffer.Home()
		} else {
			m.focus = "tree"
			m.cursorIdx = 0
			m.autoOpenAtCursor()
		}
		return m, nil
	case tea.KeyEnd:
		if m.focus == "editor" && m.buffer != nil {
			m.buffer.End()
		} else {
			m.focus = "tree"
			if len(m.flatList) > 0 {
				m.cursorIdx = len(m.flatList) - 1
				m.autoOpenAtCursor()
			}
		}
		return m, nil
	case tea.KeyEnter:
		// Go-to-line Modal: Enter führt Sprung aus
		if m.goToLineMode {
			if m.buffer != nil {
				var n int
				fmt.Sscanf(m.goToLineQuery, "%d", &n)
				m.buffer.GoToLine(n)
				m.focus = "editor"
			}
			m.goToLineMode = false
			m.goToLineQuery = ""
			return m, nil
		}
		// Search-Modal: Enter zyklisch durch Matches
		if m.searchActive && len(m.searchResults) > 0 {
			m.searchIdx++
			if m.searchIdx >= len(m.searchResults) {
				m.searchIdx = 0
			}
			match := m.searchResults[m.searchIdx]
			if m.buffer != nil {
				m.buffer.CursorRow = match.Line
				m.buffer.CursorCol = match.Col
				m.focus = "editor"
			}
			return m, nil
		}
		// Im Tree: Enter toggelt directory ODER wechselt in Editor
		if m.focus == "tree" && m.buffer != nil {
			node := m.flatList[m.cursorIdx]
			if node.IsDir {
				m.treeRender.ToggleDir(node.Path)
				m.flatList = m.workspace.FlatList(m.treeRender.CollapsedDirs)
				return m, nil
			}
			// File: Wechsel in Editor-Focus
			m.focus = "editor"
			return m, nil
		}
		// Im Editor ohne Modals: Enter = neue Zeile
		if m.buffer != nil && m.focus == "editor" {
			m.buffer.InsertNewLine()
			return m, nil
		}
		// Sonst: selectCurrent (legacy fallback)
		return m.selectCurrent()
	case tea.KeyBackspace:
		if m.goToLineMode && m.goToLineQuery != "" {
			m.goToLineQuery = m.goToLineQuery[:len(m.goToLineQuery)-1]
			return m, nil
		}
		if m.searchActive && m.searchQuery != "" {
			if len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
			}
			if m.buffer != nil {
				m.searchResults = search.Find(m.buffer.Lines, m.searchQuery, search.Options{})
			}
			return m, nil
		}
		if m.buffer != nil && m.focus == "editor" {
			m.buffer.DeleteChar()
			return m, nil
		}
		return m, nil
	case tea.KeyDelete:
		if m.buffer != nil && m.focus == "editor" {
			m.buffer.DeleteCharForward()
		}
		return m, nil
	}
	if msg.Type == tea.KeyRunes {
		// Go-to-line Mode: Query aufbauen
		if m.goToLineMode {
			for _, r := range msg.Runes {
				if r >= '0' && r <= '9' {
					m.goToLineQuery += string(r)
				}
			}
			return m, nil
		}
		// Wenn Search aktiv: Query aufbauen
		if m.searchActive && m.buffer != nil {
			for _, r := range msg.Runes {
				m.searchQuery += string(r)
			}
			// Recompute matches
			m.searchResults = search.Find(m.buffer.Lines, m.searchQuery, search.Options{})
			m.searchIdx = 0
			return m, nil
		}
		if m.buffer != nil {
			for _, r := range msg.Runes {
				m.buffer.InsertChar(r)
			}
		}
		return m, nil
	}
	return m, nil
}

func (m Model) saveBuffer() (tea.Model, tea.Cmd) {
	if m.buffer == nil {
		return m, nil
	}
	if err := m.buffer.Save(); err != nil {
		m.saveError = err.Error()
		m.mode = "SAVE-ERROR"
	} else {
		m.saveError = ""
		m.mode = "EDIT"
	}
	return m, nil
}

func (m Model) toggleCurrentDir() (tea.Model, tea.Cmd) {
	if m.cursorIdx < 0 || m.cursorIdx >= len(m.flatList) {
		return m, nil
	}
	node := m.flatList[m.cursorIdx]
	if node.IsDir {
		m.treeRender.ToggleDir(node.Path)
		m.flatList = m.workspace.FlatList(m.treeRender.CollapsedDirs)
		if m.cursorIdx >= len(m.flatList) {
			m.cursorIdx = len(m.flatList) - 1
		}
	}
	return m, nil
}

func (m Model) selectCurrent() (tea.Model, tea.Cmd) {
	if m.cursorIdx < 0 || m.cursorIdx >= len(m.flatList) {
		return m, nil
	}
	node := m.flatList[m.cursorIdx]
	if node.IsDir {
		m.treeRender.ToggleDir(node.Path)
		m.flatList = m.workspace.FlatList(m.treeRender.CollapsedDirs)
		if m.cursorIdx >= len(m.flatList) {
			m.cursorIdx = len(m.flatList) - 1
		}
		return m, nil
	}
	buf, err := editor.LoadFromFile(node.Path)
	if err == nil {
		m.buffer = buf
		m.currentFile = node.Path
		m.mode = "EDIT"
		if m.tabs != nil {
			m.tabs.Open(node.Path, buf)
		}
	}
	return m, nil
}

// autoOpenAtCursor öffnet automatisch die Datei unter dem Cursor im Tree
// wenn es ein File ist (kein Directory). Skip bei Buffern, die gerade editiert werden.
func (m Model) autoOpenAtCursor() {
	if m.cursorIdx < 0 || m.cursorIdx >= len(m.flatList) {
		return
	}
	node := m.flatList[m.cursorIdx]
	if node.IsDir {
		return
	}
	// Vermeide Reload wenn schon offen (User tippt noch)
	if m.buffer != nil && m.currentFile == node.Path {
		return
	}
	buf, err := editor.LoadFromFile(node.Path)
	if err == nil {
		m.buffer = buf
		m.currentFile = node.Path
		m.mode = "EDIT"
		if m.tabs != nil {
			m.tabs.Open(node.Path, buf)
		}
	}
}

func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Type != tea.MouseLeft {
		return m, nil
	}
	if msg.X >= 1 && msg.X < m.layout.SidebarWidth-1 {
		idx := msg.Y - 2
		if idx < 0 {
			idx = 0
		}
		if idx >= 0 && idx < len(m.flatList) {
			m.cursorIdx = idx
			return m.selectCurrent()
		}
	}
	return m, nil
}

func (m Model) View() string {
	m.clampScrollOffset()

	if m.quitting {
		return "mdskim2 — beendet.\n"
	}
	if m.width == 0 || m.height == 0 {
		return fmt.Sprintf("mdskim2 — Workspace: %s — R5 MD-Highlight\n", m.workspacePath)
	}

	header := m.layout.Header(m.workspacePath, m.theme)
	sidebar := m.layout.Sidebar(m.renderSidebar(), m.theme, true)
	editorContent := m.renderEditor()
	if m.paletteActive && m.palette != nil {
		matches := m.palette.Search(m.paletteQuery)
		var matchLines []string
		matchLines = append(matchLines, fmt.Sprintf("CMD: %s", m.paletteQuery))
		for i, name := range matches {
			if i >= 5 { break }
			desc, _ := m.palette.Run(name)
			matchLines = append(matchLines, fmt.Sprintf("  > %s — %s", name, desc))
		}
		editorContent = strings.Join(matchLines, "\n") + "\n" + editorContent
	}
	if m.searchActive {
		count := len(m.searchResults)
		barText := fmt.Sprintf("SUCHE: %s [Enter: Jump, Esc: Close, Ctrl+H: Replace, %d Treffer]", m.searchQuery, count)
		editorContent = barText + "\n" + editorContent
	}
	editor := m.layout.Editor(editorContent, m.theme, m.buffer != nil)
	toc := m.layout.TOC(m.renderTOC(), m.theme, len(m.tocHeadings()) > 0)
	body := m.layout.Compose(sidebar, editor, toc)

	mode := m.mode
	if m.saveError != "" {
		mode = "ERR: " + m.saveError
	}

	statusInfo := ui.StatusInfo{
		Workspace: m.workspacePath,
		File:      m.currentFile,
		Encoding:  "UTF-8",
		Mode:      mode,
		Version:   m.version,
	}
	if m.buffer != nil {
		statusInfo.Lines = m.buffer.TotalLines()
		statusInfo.Words = m.buffer.TotalWords()
		statusInfo.Modified = m.buffer.Modified
	}
	status := m.layout.RenderStatus(statusInfo, m.theme)
	footer := m.layout.RenderFooter(ui.DefaultShortcuts(), m.theme)

	return strings.Join([]string{header, body, status, footer}, "\n")
}

func (m Model) renderSidebar() string {
	if m.workspace == nil {
		return "FILES\n(kein Workspace)"
	}
	headerMark := "\x1b[2m FILES \x1b[0m"
	if m.focus == "tree" {
		headerMark = "\x1b[48;5;63m\x1b[1;37m FILES \x1b[0m"
	}
	var lines []string
	lines = append(lines, headerMark)
	for i, node := range m.flatList {
		var line string
		if i == m.cursorIdx {
			line = "\x1b[1;33m▶\x1b[0m " + strings.TrimPrefix(m.treeRender.FormatNode(node), " ")
			if m.focus == "tree" {
				// aktive Zeile zusätzlich hervorheben
				line = "\x1b[7m" + strings.TrimPrefix(m.treeRender.FormatNode(node), " ") + "\x1b[0m"
				line = "\x1b[1;33m▶\x1b[0m " + line
			}
		} else {
			line = "  " + m.treeRender.FormatNode(node)
		}
		lines = append(lines, line)
	}
	if m.workspace.TotalFiles() == 0 {
		lines = append(lines, "(leer)")
	}
	footerHint := "\n\x1b[2m ↑↓ Navig · Enter Tree→Edit \x1b[0m"
	return strings.Join(lines, "\n") + footerHint
}

func (m Model) renderEditor() string {
	if m.buffer == nil {
		return "\x1b[1;36m[NO-BUFFER]\x1b[0m \x1b[1;37m# Willkommen bei mdskim2\x1b[0m\n\n" +
			"\x1b[33m↑↓\x1b[0m navigiert Tree und lädt Datei automatisch\n" +
			"\x1b[33mEnter\x1b[0m wechselt in den Editor\n" +
			"\x1b[33mEsc\x1b[0m zurück zum Tree\n" +
			"\x1b[33mCtrl+S\x1b[0m Speichern, \x1b[33mCtrl+Q\x1b[0m Beenden\n\n" +
			"Workspace: \x1b[1m" + m.workspacePath + "\x1b[0m"
	}

	// Header-Bar mit Datei + Modifiziert-Flag + Focus
	dirtyMark := "  "
	if m.buffer.Modified {
		dirtyMark = "\x1b[1;31m●\x1b[0m"
	}
	focusMark := "\x1b[48;5;240m\x1b[37m TREE \x1b[0m"
	if m.focus == "editor" {
		focusMark = "\x1b[48;5;63m\x1b[1;37m EDIT \x1b[0m"
	}
	wrapMark := ""
	if m.lineWrap {
		wrapMark = " \x1b[2;37m[WRAP]\x1b[0m"
	}
	headerLine := focusMark + " \x1b[1;37m" + trimPath(m.currentFile, 50) + "\x1b[0m " + dirtyMark + wrapMark

	// Selection-Range
	srSel, scSel, erSel, ecSel, hasSel := m.buffer.SelectionRange()

	// Line numbers + body mit Block-Cursor + Selection-Highlight
	totalLines := len(m.buffer.Lines)
	lineNumWidth := len(fmt.Sprintf("%d", totalLines))
	if lineNumWidth < 2 {
		lineNumWidth = 2
	}

	// Viewport-Scrolling: zeige Buffer ab scrollOffset
	viewportHeight := m.editorViewportHeight()
	visibleStart := m.editorScrollOffset
	visibleEnd := visibleStart + viewportHeight
	if visibleEnd > totalLines {
		visibleEnd = totalLines
	}
	if visibleStart >= visibleEnd {
		visibleStart = 0
		visibleEnd = totalLines
		if totalLines > viewportHeight {
			visibleEnd = viewportHeight
		}
	}

	var bodyLines []string
	for i := visibleStart; i < visibleEnd; i++ {
		line := m.buffer.Lines[i]
		// Gutter: line number
		var numFmt string
		if i == m.buffer.CursorRow {
			numFmt = fmt.Sprintf("\x1b[1;33m%*d\x1b[0m", lineNumWidth, i+1)
		} else {
			numFmt = fmt.Sprintf("\x1b[2;37m%*d\x1b[0m", lineNumWidth, i+1)
		}

		// Build content with highlighting
		p := markdown.HighlightParams{CurrentLine: m.buffer.CursorRow}
		rendered := markdown.HighlightLineWith(line, p, i)

		// Selection-Markierung: reverse-video Background auf markiertem Bereich
		if hasSel && i >= srSel && i <= erSel {
			rendered = applySelectionBG(rendered, line, i, srSel, scSel, erSel, ecSel)
		}

		// Block-Cursor an CursorCol
		if i == m.buffer.CursorRow && m.focus == "editor" {
			rendered = injectBlockCursor(rendered, line, m.buffer.CursorCol)
		}

		bodyLines = append(bodyLines, numFmt+" │ "+rendered)
	}
	body := strings.Join(bodyLines, "\n")

	// Scroll-Indicator rechts (1 Zeichen pro Block)
	var scrollArrow string
	if visibleStart > 0 {
		scrollArrow += "\x1b[33m↑\x1b[0m"
	}
	if visibleEnd < totalLines {
		scrollArrow += "\x1b[33m↓\x1b[0m"
	}
	scrollInfo := ""
	if scrollArrow != "" {
		scrollInfo = " " + scrollArrow + fmt.Sprintf(" (%d-%d/%d)", visibleStart+1, visibleEnd, totalLines)
	}

	// Status-Bar mit allen Editor-Infos
	var currentLineLen int
	if m.buffer.CursorRow < len(m.buffer.Lines) {
		currentLineLen = utf8.RuneCountInString(m.buffer.Lines[m.buffer.CursorRow])
	}
	wc := m.buffer.WordCount()
	cc := m.buffer.CharCount()
	cursorPos := fmt.Sprintf("\x1b[48;5;236m \x1b[1;33m Ln %d/%d \x1b[0m\x1b[48;5;236m \x1b[33m Col %d/%d \x1b[0m",
		m.buffer.CursorRow+1, totalLines, m.buffer.CursorCol+1, currentLineLen)
	counters := fmt.Sprintf("\x1b[2;37m Words %d · Chars %d \x1b[0m", wc, cc)
	hints := "\x1b[2;37m [Shift+Arrows] Sel · [Ctrl+D] Dup · [Ctrl+G] Go-to · [Ctrl+W] Wrap \x1b[0m"

	// Optional: Go-to-line Modal-Bar
	gotoBar := ""
	if m.goToLineMode {
		gotoBar = "\n\x1b[48;5;240m\x1b[1;37m Go to line: \x1b[0m\x1b[1;33m" + m.goToLineQuery + "_\x1b[0m"
	}

	return headerLine + "\n" + body + "\n" + cursorPos + " " + counters + " " + scrollInfo + "\n" + hints + gotoBar
}

// editorViewportHeight returns the available height for the editor content.
func (m Model) editorViewportHeight() int {
	return m.height - 8
}

// editorScrollOffset tracks the first visible line.
func (m *Model) clampScrollOffset() {
	if m.buffer == nil {
		return
	}
	vp := m.editorViewportHeight()
	if m.editorScrollOffset < 0 {
		m.editorScrollOffset = 0
	}
	maxOff := len(m.buffer.Lines) - vp
	if maxOff < 0 {
		maxOff = 0
	}
	if m.editorScrollOffset > maxOff {
		m.editorScrollOffset = maxOff
	}
	if m.buffer.CursorRow < m.editorScrollOffset {
		m.editorScrollOffset = m.buffer.CursorRow
	}
	if m.buffer.CursorRow >= m.editorScrollOffset+vp {
		m.editorScrollOffset = m.buffer.CursorRow - vp + 1
	}
}

// applySelectionBG wandelt den gerenderten String so um, dass der markierte Bereich in Selection-Farbe erscheint.
func applySelectionBG(rendered string, originalLine string, lineIdx, sr, sc, er, ec int) string {
	startCol := 0
	endCol := len(originalLine)
	if lineIdx == sr {
		startCol = sc
	}
	if lineIdx == er {
		endCol = ec
	}
	if startCol >= endCol {
		return rendered
	}
	// Wrap the substring in selection highlight
	before := string([]rune(originalLine)[:startCol])
	selText := string([]rune(originalLine)[startCol:endCol])
	after := string([]rune(originalLine)[endCol:])
	// Note: für selected text rendern wir plain — highlighting wird durch selection BG ersetzt
	return before + "\x1b[48;5;57m" + selText + "\x1b[0m" + after
}

// injectBlockCursor fügt einen Block-Cursor (reverse-video) an col in line ein.
func injectBlockCursor(rendered string, originalLine string, col int) string {
	runes := []rune(originalLine)
	if col > len(runes) {
		col = len(runes)
	}
	before := string(runes[:col])
	if col < len(runes) {
		char := runes[col]
		after := string(runes[col+1:])
		return before + "\x1b[7m" + string(char) + "\x1b[0m" + after
	}
	return rendered + "\x1b[7m \x1b[0m"
}

func (m Model) tocHeadings() []markdown.Heading {
	if m.buffer == nil {
		return nil
	}
	return markdown.Headings(m.buffer.ToString())
}

func (m Model) renderTOC() string {
	hs := m.tocHeadings()
	if len(hs) == 0 {
		return "INHALT\n──────\n▸ (Datei öffnen)"
	}
	tocLines := []string{"INHALT", "──────"}
	for _, h := range hs {
		indent := strings.Repeat("  ", h.Level-1)
		bullet := "▸"
		if h.Level == 1 {
			bullet = "▾"
		}
		tocLines = append(tocLines, fmt.Sprintf("%s%s %s", indent, bullet, h.Text))
	}
	return strings.Join(tocLines, "\n")
}

func trimPath(path string, max int) string {
	if len(path) <= max {
		return path
	}
	return "…" + path[len(path)-max+1:]
}

// Quitting returns true wenn Ctrl+Q gedrückt wurde.
func (m Model) Quitting() bool { return m.quitting }

// FlatList returns the FlatList for testing.
func (m Model) FlatList() []*workspace.FileNode { return m.flatList }

// WithCursor returns a copy with cursorIdx set.
func (m Model) WithCursor(idx int) Model {
	if idx >= 0 {
		m.cursorIdx = idx
	}
	return m
}

// CurrentFile returns the current file path.
func (m Model) CurrentFile() string { return m.currentFile }

// Cursor returns current cursor position.
func (m Model) Cursor() int { return m.cursorIdx }

func (m Model) Buffer() *editor.Buffer { return m.buffer }

// IsBufferOpen returns true if the buffer is loaded.
func (m Model) IsBufferOpen() bool { return m.buffer != nil }

// HighlightedLines returns the buffer content with ANSI-highlight applied per line.
// Public for testing — bypasses lipgloss width constraint.
func (m Model) HighlightedLines() []string {
	if m.buffer == nil {
		return nil
	}
	var hl []string
	p := markdown.HighlightParams{CurrentLine: m.buffer.CursorRow}
	for _, line := range m.buffer.Lines {
		hl = append(hl, markdown.HighlightLine(line, p))
	}
	return hl
}

// TocText returns the TOC pane content (hierarchical headings).
func (m Model) TocText() string {
	hs := m.tocHeadings()
	if len(hs) == 0 {
		return "INHALT\n──────\n▸ (Datei öffnen)"
	}
	tocLines := []string{"INHALT", "──────"}
	for _, h := range hs {
		indent := strings.Repeat("  ", h.Level-1)
		bullet := "▸"
		if h.Level == 1 {
			bullet = "▾"
		}
		tocLines = append(tocLines, fmt.Sprintf("%s%s %s", indent, bullet, h.Text))
	}
	return strings.Join(tocLines, "\n")
}

// HeadingsCount returns number of headings in current buffer.
func (m Model) HeadingsCount() int {
	return len(m.tocHeadings())
}
