package ui

import "github.com/charmbracelet/lipgloss"

// Layout hält die berechneten Größen des Layouts für ein Window.
type Layout struct {
	Width, Height int
	HeaderHeight  int
	StatusHeight  int
	FooterHeight  int

	// Sidebar (Links)
	SidebarWidth int

	// TOC (Rechts)
	TOCWidth int

	// Editor (Mitte) — wird aus Rest berechnet.
	EditorWidth  int
	EditorHeight int
}

// Defaults für Layout (Go erlaubt keine Default-Werte in Struct-Defs,
// daher zentral hier).
func DefaultLayout() Layout {
	return Layout{
		HeaderHeight: 1,
		StatusHeight: 1,
		FooterHeight: 1,
		SidebarWidth: 24,
		TOCWidth:     28,
	}
}

// Compute berechnet die Pane-Größen aus Terminal-Dimensionen.
func (l *Layout) Compute(width, height int) {
	// Defaults sicherstellen (falls leer)
	if l.HeaderHeight == 0 { l.HeaderHeight = 1 }
	if l.StatusHeight == 0 { l.StatusHeight = 1 }
	if l.FooterHeight == 0 { l.FooterHeight = 1 }
	if l.SidebarWidth == 0 { l.SidebarWidth = 24 }
	if l.TOCWidth == 0 { l.TOCWidth = 28 }
	l.Width = width
	l.Height = height

	contentHeight := height - l.HeaderHeight - l.StatusHeight - l.FooterHeight
	if contentHeight < 1 {
		contentHeight = 1
	}
	l.EditorHeight = contentHeight

	contentWidth := width - l.SidebarWidth - l.TOCWidth
	if contentWidth < 1 {
		contentWidth = 1
	}
	l.EditorWidth = contentWidth
}

// Header rendert die obere Header-Zeile mit Workspace-Pfad.
func (l *Layout) Header(workspace string, theme Theme) string {
	style := theme.Header.
		Width(l.Width).
		Align(lipgloss.Center)
	text := " mdskim2 — " + workspace + " "
	return style.Render(text)
}

// Sidebar rendert die linke Pane (Datei-Tree).
func (l *Layout) Sidebar(content string, theme Theme, isActive bool) string {
	border := theme.InactiveBorder
	if isActive {
		border = theme.ActiveBorder
	}

	innerWidth := l.SidebarWidth - 2
	if innerWidth < 1 {
		innerWidth = 1
	}
	innerHeight := l.EditorHeight - 2
	if innerHeight < 1 {
		innerHeight = 1
	}
	rendered := lipgloss.NewStyle().Width(innerWidth).Height(innerHeight).Render(content)
	return border.
		Width(l.SidebarWidth).
		Height(l.EditorHeight).
		Render(rendered)
}

// Editor rendert die mittlere Pane.
func (l *Layout) Editor(content string, theme Theme, isActive bool) string {
	border := theme.InactiveBorder
	if isActive {
		border = theme.ActiveBorder
	}

	innerWidth := l.EditorWidth - 2
	if innerWidth < 1 {
		innerWidth = 1
	}
	innerHeight := l.EditorHeight - 2
	if innerHeight < 1 {
		innerHeight = 1
	}
	rendered := lipgloss.NewStyle().Width(innerWidth).Height(innerHeight).Render(content)
	return border.
		Width(l.EditorWidth).
		Height(l.EditorHeight).
		Render(rendered)
}

// TOC rendert die rechte Pane (Inhaltsverzeichnis).
func (l *Layout) TOC(content string, theme Theme, isActive bool) string {
	border := theme.InactiveBorder
	if isActive {
		border = theme.ActiveBorder
	}

	innerWidth := l.TOCWidth - 2
	if innerWidth < 1 {
		innerWidth = 1
	}
	innerHeight := l.EditorHeight - 2
	if innerHeight < 1 {
		innerHeight = 1
	}
	rendered := lipgloss.NewStyle().Width(innerWidth).Height(innerHeight).Render(content)
	return border.
		Width(l.TOCWidth).
		Height(l.EditorHeight).
		Render(rendered)
}

// Compose setzt die 3 vertikalen Panes nebeneinander.
func (l *Layout) Compose(sidebar, editor, toc string) string {
	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, editor, toc)
}
