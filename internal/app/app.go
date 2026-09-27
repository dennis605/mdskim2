// Package app implementiert die mdskim2-Anwendung als Bubble-Tea-Model.
package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dennis605/mdskim2/internal/ui"
	"github.com/dennis605/mdskim2/internal/workspace"
)

// Model ist der zentrale Anwendungs-State.
type Model struct {
	// Input
	workspacePath string
	width         int
	height        int
	quitting      bool

	// Workspace + Tree
	workspace  *workspace.Workspace
	treeRender *workspace.TreeRenderer
	cursorIdx  int // Index in der Tree-FlatList
	flatList   []*workspace.FileNode

	// Editor (R3 - Stub für R2)
	currentFile string

	// Status
	mode    string
	theme   ui.Theme
	layout  ui.Layout
	version string
}

// New erzeugt einen frischen Anwendungs-Model.
func New(workspacePath string) Model {
	ws, _ := workspace.Load(workspacePath)
	r := workspace.NewTreeRenderer()
	flat := ws.FlatList(r.CollapsedDirs)

	return Model{
		workspacePath: workspacePath,
		width:         120,
		height:        40,
		mode:          "EDIT",
		version:       "R2: Workspace + Tree",
		theme:         ui.Light(),
		layout:        ui.DefaultLayout(),
		workspace:     ws,
		treeRender:    r,
		flatList:      flat,
		cursorIdx:     0,
	}
}

// Init ist beim Bubble-Tea-Start erforderlich.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update verarbeitet eingehende Messages.
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

// handleKey verarbeitet Tastatur-Events.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlQ:
		m.quitting = true
		return m, tea.Quit

	case tea.KeyEsc:
		return m, nil

	case tea.KeyUp:
		if m.cursorIdx > 0 {
			m.cursorIdx--
		}
		return m, nil

	case tea.KeyDown:
		if m.cursorIdx < len(m.flatList)-1 {
			m.cursorIdx++
		}
		return m, nil

	case tea.KeyEnter:
		return m.selectCurrent()
	}

	switch msg.String() {
	case "backspace":
		// Parent-Verzeichnis: wenn cursor auf Sub-Dir, klappe zu
		if m.cursorIdx >= 0 && m.cursorIdx < len(m.flatList) {
			node := m.flatList[m.cursorIdx]
			if node.IsDir {
				m.treeRender.ToggleDir(node.Path)
				m.flatList = m.workspace.FlatList(m.treeRender.CollapsedDirs)
				if m.cursorIdx >= len(m.flatList) {
					m.cursorIdx = len(m.flatList) - 1
				}
			}
		}
		return m, nil
	}

	return m, nil
}

// selectCurrent öffnet das selektierte Item (Enter / Click).
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
	} else {
		// Datei: in Editor laden (R3 füllt Buffer, R2 zeigt nur Pfad)
		m.currentFile = node.Path
		m.mode = "EDIT"
	}
	return m, nil
}

// handleMouse verarbeitet Maus-Events.
func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Type != tea.MouseLeft {
		return m, nil
	}

	// Sidebar-Spalte (1-basiert: Spalte 0 bis SidebarWidth-1)
	if msg.X >= 1 && msg.X < m.layout.SidebarWidth-1 {
		// Konvertiere Y in Tree-Index
		idx := msg.Y - 2 // 1 = Header, 1 = Sidebar-Label, danach Tree
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

// View rendert den aktuellen State als String.
func (m Model) View() string {
	if m.quitting {
		return "mdskim2 — beendet.\n"
	}

	if m.width == 0 || m.height == 0 {
		return fmt.Sprintf("mdskim2 — Workspace: %s — R2 Workspace+Tree\n", m.workspacePath)
	}

	header := m.layout.Header(m.workspacePath, m.theme)

	sidebarContent := m.renderSidebar()
	editorContent := m.renderEditor()
	tocContent := m.renderTOC()

	sidebar := m.layout.Sidebar(sidebarContent, m.theme, true)
	editor := m.layout.Editor(editorContent, m.theme, false)
	toc := m.layout.TOC(tocContent, m.theme, false)
	body := m.layout.Compose(sidebar, editor, toc)

	statusInfo := ui.StatusInfo{
		Workspace: m.workspacePath,
		File:      m.currentFile,
		Encoding:  "UTF-8",
		Mode:      m.mode,
		Version:   m.version,
	}
	status := m.layout.RenderStatus(statusInfo, m.theme)
	footer := m.layout.RenderFooter(ui.DefaultShortcuts(), m.theme)

	return strings.Join([]string{header, body, status, footer}, "\n")
}

// renderSidebar rendert den FileTree mit Selection.
func (m Model) renderSidebar() string {
	if m.workspace == nil {
		return "FILES\n(kein Workspace)"
	}

	var lines []string
	lines = append(lines, "FILES")

	for i, node := range m.flatList {
		line := m.treeRender.FormatNode(node)
		if i == m.cursorIdx {
			line = "▶ " + strings.TrimPrefix(line, "")
		}
		lines = append(lines, line)
	}

	if m.workspace.TotalFiles() == 0 {
		lines = append(lines, "(leer)")
	}

	return strings.Join(lines, "\n")
}

// renderEditor rendert den Editor-Inhalt.
func (m Model) renderEditor() string {
	if m.currentFile == "" {
		heading := "# Welcome to mdskim2"
		para := "Modern Markdown Workspace for the Terminal.\n" +
			"Obsidian/VS Code feel, no vim modes, single binary.\n\n" +
			"Sprint " + m.version + "\n\n" +
			"Workspace: " + m.workspacePath + "\n\n" +
			"↑↓ navigieren · Enter öffnet · Backspace klappt zu"
		return heading + "\n\n" + para
	}

	heading := "# " + trimPath(m.currentFile, 40)
	body := "(Editor-Buffer kommt in R3)\n\nAktuelle Datei:\n" + m.currentFile
	return heading + "\n\n" + body
}

// renderTOC rendert das Inhaltsverzeichnis (R5 füllt es live).
func (m Model) renderTOC() string {
	if m.currentFile == "" {
		return "INHALT\n──────\n▸ (Datei öffnen)"
	}
	return "INHALT\n──────\n(TOC kommt in R5)"
}

// trimPath kürzt einen langen Pfad für die Anzeige.
func trimPath(path string, max int) string {
	if len(path) <= max {
		return path
	}
	return "…" + path[len(path)-max+1:]
}
