// Package app implementiert die mdskim2-Anwendung als Bubble-Tea-Model.
//
// Architektur: Elm-Pattern. Update() ist ein Switch auf tea.Msg, View()
// rendert via internal/ui einen 3-Pane-Layout-Snapshot.
package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dennis605/mdskim2/internal/ui"
)

// Model ist der zentrale Anwendungs-State.
type Model struct {
	// Input
	workspace string
	width     int
	height    int
	quitting  bool

	// Status-Anzeige
	currentFile string
	mode        string

	// Theme
	theme ui.Theme

	// Layout
	layout ui.Layout

	// Sprint-Identifikation (für Status-Bar)
	version string
}

// New erzeugt einen frischen Anwendungs-Model.
func New(workspace string) Model {
	return Model{
		workspace: workspace,
		width:     120,
		height:    40,
		mode:      "BOOT",
		version:   "R1: Bootstrap",
		theme:     ui.Light(),
		layout:    ui.DefaultLayout(),
	}
}

// Init ist beim Bubble-Tea-Start erforderlich (siehe tea.Model interface).
func (m Model) Init() tea.Cmd {
	return nil
}

// Update verarbeitet eingehende Messages und gibt neuen State zurück.
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
		return m, nil

	default:
		return m, nil
	}
}

// handleKey verarbeitet Tastatur-Events gemäß User-Spec-Shortcuts.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Ctrl+Q: Beenden (Spec-Hard-Shortcut, in R1 schon implementiert)
	if msg.Type == tea.KeyCtrlQ {
		m.quitting = true
		return m, tea.Quit
	}

	// Esc abfangen, falls nötig
	if msg.Type == tea.KeyEsc {
		return m, nil
	}

	return m, nil
}

// View rendert den aktuellen State als String.
func (m Model) View() string {
	if m.quitting {
		return "mdskim2 — beendet.\n"
	}

	if m.width == 0 || m.height == 0 {
		// Bubble Tea hat uns noch keine Window-Size geschickt → Mini-View
		return fmt.Sprintf("mdskim2 — Workspace: %s — R1 Bootstrap\n", m.workspace)
	}

	header := m.layout.Header(m.workspace, m.theme)

	// 3 vertikale Panes nebeneinander
	sidebar := m.layout.Sidebar(sidebarContent(), m.theme, false)
	editor := m.layout.Editor(editorContent(m.workspace, m.version), m.theme, true)
	toc := m.layout.TOC(tocContent(), m.theme, false)
	body := m.layout.Compose(sidebar, editor, toc)

	statusInfo := ui.StatusInfo{
		Workspace: m.workspace,
		File:      m.currentFile,
		Encoding:  "UTF-8",
		Mode:      m.mode,
		Version:   m.version,
	}
	status := m.layout.RenderStatus(statusInfo, m.theme)
	footer := m.layout.RenderFooter(ui.DefaultShortcuts(), m.theme)

	return strings.Join([]string{header, body, status, footer}, "\n")
}

// sidebarContent rendert den Beispiel-Tree (R2 ersetzt dies).
func sidebarContent() string {
	return "FILES\n\n▾ Test Engineering\n  • Tests…\n  • Testpl…\n  • Testar…\n  • Testmet…\n\nTAGS\n\n#obsidian\n#markdown"
}

// editorContent rendert den Editor (R3-R5 füllen Buffer + Markdown-Highlight).
func editorContent(workspace, version string) string {
	heading := "# Welcome to mdskim2"
	para := "Modern Markdown Workspace for the Terminal. " +
		"Obsidian/VS Code feel, no vim modes, single binary."
	return heading + "\n\n" + para + "\n\nSprint " + version + "\n\nWorkspace: " + workspace
}

// tocContent rendert das TOC (R5 füllt es live mit Headings).
func tocContent() string {
	return "INHALT\n──────\n▸ Welcome\n▸ Sprint R1"
}
