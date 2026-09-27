// Package app implementiert die mdskim2-Anwendung als Bubble-Tea-Model.
package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/atotto/clipboard"

	"github.com/dennis605/mdskim2/internal/editor"
	"github.com/dennis605/mdskim2/internal/ui"
	"github.com/dennis605/mdskim2/internal/workspace"
)

// Model ist der zentrale Anwendungs-State.
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
	selection   string // für R5: aktuelle Selection
	mode        string
	currentFile string

	theme   ui.Theme
	layout  ui.Layout
	version string

	// Save-Status
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
		version:       "R4: Save + Clipboard",
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
		// Wenn Buffer aktiv: Copy line (Selection kommt in R5)
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
		// Markdown Bold
		if m.buffer != nil {
			m.buffer.InsertString("****")
			m.buffer.MoveCursor(0, -2) // Cursor zwischen die ** setzen
		}
		return m, nil

	case tea.KeyCtrlI:
		// Markdown Italic
		if m.buffer != nil {
			m.buffer.InsertString("**")
			m.buffer.MoveCursor(0, -1) // Cursor zwischen die * setzen
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
		if m.buffer != nil {
			m.buffer.InsertNewLine()
			return m, nil
		}
		return m.selectCurrent()
	case tea.KeyBackspace:
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
		if m.buffer != nil {
			for _, r := range msg.Runes {
				m.buffer.InsertChar(r)
			}
		}
		return m, nil
	}
	return m, nil
}

// saveBuffer speichert den Buffer auf Disk.
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
		return fmt.Sprintf("mdskim2 — Workspace: %s — R4 Save+Clipboard\n", m.workspacePath)
	}

	header := m.layout.Header(m.workspacePath, m.theme)
	sidebar := m.layout.Sidebar(m.renderSidebar(), m.theme, true)
	editor := m.layout.Editor(m.renderEditor(), m.theme, m.buffer != nil)
	toc := m.layout.TOC(m.renderTOC(), m.theme, false)
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
		return "# Willkommen bei mdskim2\n\n" +
			"↑↓ in Sidebar · Enter öffnet File\n\n" +
			"Workspace: " + m.workspacePath
	}
	heading := "# " + trimPath(m.currentFile, 40)
	body := m.buffer.ToString()
	return heading + "\n\n" + body
}

func (m Model) renderTOC() string {
	if m.buffer == nil {
		return "INHALT\n──────\n▸ (Datei öffnen)"
	}
	var tocLines []string
	tocLines = append(tocLines, "INHALT")
	tocLines = append(tocLines, "──────")
	for _, line := range m.buffer.Lines {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "# ") {
			tocLines = append(tocLines, fmt.Sprintf("▸ %s", strings.TrimPrefix(t, "# ")))
		}
	}
	if len(tocLines) == 2 {
		tocLines = append(tocLines, "(keine H1)")
	}
	return strings.Join(tocLines, "\n")
}

func trimPath(path string, max int) string {
	if len(path) <= max {
		return path
	}
	return "…" + path[len(path)-max+1:]
}
