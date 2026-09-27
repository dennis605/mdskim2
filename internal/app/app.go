// Package app implementiert die mdskim2-Anwendung als Bubble-Tea-Model.
package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/atotto/clipboard"

	"github.com/dennis605/mdskim2/internal/editor"
	"github.com/dennis605/mdskim2/internal/markdown"
	"github.com/dennis605/mdskim2/internal/preview"
	"github.com/dennis605/mdskim2/internal/search"
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

	theme   ui.Theme
	layout  ui.Layout
	version string

	saveError string
}

func New(workspacePath string) Model {
	ws, _ := workspace.Load(workspacePath)
	r := workspace.NewTreeRenderer()
	flat := ws.FlatList(r.CollapsedDirs)

	return Model{
		workspacePath: workspacePath,
		width:         120,
		height:        40,
		mode:          "EDIT",
		version:       "R7: Preview + Split",
		theme:         ui.Light(),
		layout:        ui.DefaultLayout(),
		workspace:     ws,
		treeRender:    r,
		flatList:      flat,
	}
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
		return m, nil
	case tea.KeyUp:
		if m.buffer != nil {
			m.buffer.MoveCursor(-1, 0)
		} else if m.cursorIdx > 0 {
			m.cursorIdx--
		}
		return m, nil
	case tea.KeyDown:
		if m.buffer != nil {
			m.buffer.MoveCursor(1, 0)
		} else if m.cursorIdx < len(m.flatList)-1 {
			m.cursorIdx++
		}
		return m, nil
	case tea.KeyLeft:
		if m.buffer != nil {
			m.buffer.MoveCursor(0, -1)
		}
		return m, nil
	case tea.KeyRight:
		if m.buffer != nil {
			m.buffer.MoveCursor(0, 1)
		}
		return m, nil
	case tea.KeyHome:
		if m.buffer != nil {
			m.buffer.Home()
		}
		return m, nil
	case tea.KeyEnd:
		if m.buffer != nil {
			m.buffer.End()
		}
		return m, nil
	case tea.KeyEnter:
		if m.searchActive && len(m.searchResults) > 0 {
			m.searchIdx++
			if m.searchIdx >= len(m.searchResults) {
				m.searchIdx = 0
			}
			match := m.searchResults[m.searchIdx]
			if m.buffer != nil {
				m.buffer.CursorRow = match.Line
				m.buffer.CursorCol = match.Col
			}
			return m, nil
		}
		if m.buffer != nil {
			m.buffer.InsertNewLine()
			return m, nil
		}
		return m.selectCurrent()
	case tea.KeyBackspace:
		if m.searchActive && m.searchQuery != "" {
			if len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
			}
			if m.buffer != nil {
				m.searchResults = search.Find(m.buffer.Lines, m.searchQuery, search.Options{})
			}
			return m, nil
		}
		if m.buffer != nil {
			m.buffer.DeleteChar()
			return m, nil
		}
		return m.toggleCurrentDir()
	case tea.KeyDelete:
		if m.buffer != nil {
			m.buffer.DeleteCharForward()
		}
		return m, nil
	}

	if msg.Type == tea.KeyRunes {
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
	}
	return m, nil
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
	if m.quitting {
		return "mdskim2 — beendet.\n"
	}
	if m.width == 0 || m.height == 0 {
		return fmt.Sprintf("mdskim2 — Workspace: %s — R5 MD-Highlight\n", m.workspacePath)
	}

	header := m.layout.Header(m.workspacePath, m.theme)
	sidebar := m.layout.Sidebar(m.renderSidebar(), m.theme, true)
	editorContent := m.renderEditor()
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
	var lines []string
	lines = append(lines, "FILES")
	for i, node := range m.flatList {
		line := m.treeRender.FormatNode(node)
		if i == m.cursorIdx {
			line = "▶ " + strings.TrimPrefix(line, " ")
		}
		lines = append(lines, line)
	}
	if m.workspace.TotalFiles() == 0 {
		lines = append(lines, "(leer)")
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderEditor() string {
	if m.buffer == nil {
		return "[NO-BUFFER] # Willkommen bei mdskim2\n\n" +
			"↑↓ in Sidebar · Enter öffnet File\n\n" +
			"Workspace: " + m.workspacePath
	}
	var highlighted []string
	p := markdown.HighlightParams{CurrentLine: m.buffer.CursorRow}
	for _, line := range m.buffer.Lines {
		highlighted = append(highlighted, markdown.HighlightLine(line, p))
	}
	heading := "[BUFFER-OPEN] # " + trimPath(m.currentFile, 40)
	highlightedHeading := markdown.HighlightLine(heading, p)
	return strings.Join([]string{highlightedHeading, "", strings.Join(highlighted, "\n")}, "\n")
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
