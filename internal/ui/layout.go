package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Layout hält die berechneten Größen des Layouts für ein Window.
type Layout struct {
	Width, Height int
	HeaderHeight  int
	ToolbarHeight int // U6: 1-line keybinding-hint bar
	StatusHeight  int
	FooterHeight  int

	// Sidebar (Links)
	SidebarWidth int

	// Right Pane (Tabs Preview | TOC | Backlinks)
	RightWidth int

	// Editor (Mitte) — wird aus Rest berechnet.
	EditorWidth  int
	EditorHeight int
}

// Defaults für Layout (Go erlaubt keine Default-Werte in Struct-Defs,
// daher zentral hier).
func DefaultLayout() Layout {
	return Layout{
		HeaderHeight:  1,
		ToolbarHeight: 1,
		StatusHeight:  1,
		FooterHeight:  1,
		SidebarWidth:  24,
		RightWidth:    32,
	}
}

// Compute berechnet die Pane-Größen aus Terminal-Dimensionen.
func (l *Layout) Compute(width, height int) {
	if l.HeaderHeight == 0 {
		l.HeaderHeight = 1
	}
	if l.ToolbarHeight == 0 {
		l.ToolbarHeight = 1
	}
	if l.StatusHeight == 0 {
		l.StatusHeight = 1
	}
	if l.FooterHeight == 0 {
		l.FooterHeight = 1
	}
	if l.SidebarWidth == 0 {
		l.SidebarWidth = 24
	}
	if l.RightWidth == 0 {
		l.RightWidth = 32
	}
	l.Width = width
	l.Height = height

	contentHeight := height - l.HeaderHeight - l.ToolbarHeight - l.StatusHeight - l.FooterHeight
	if contentHeight < 1 {
		contentHeight = 1
	}
	l.EditorHeight = contentHeight

	contentWidth := width - l.SidebarWidth - l.RightWidth
	if contentWidth < 1 {
		contentWidth = 1
	}
	l.EditorWidth = contentWidth
}

// Header rendert die obere Header-Zeile mit Workspace-Pfad (Titel-Bar).
func (l *Layout) Header(workspace string, theme Theme) string {
	style := theme.Header.
		Width(l.Width).
		Align(lipgloss.Center)
	text := " mdskim2 — " + workspace + " "
	return style.Render(text)
}

// Toolbar rendert die 1-zeilige Shortcut-Leiste (U6: angelehnt an Python-mdskim).
// Erwartet `shortcuts` als Liste von (key, desc)-Paaren. Compact dargestellt.
func (l *Layout) Toolbar(brand string, items []Shortcut, theme Theme) string {
	parts := make([]string, 0, len(items)+1)
	if brand != "" {
		parts = append(parts, theme.Shortcut.Render(brand))
	}
	for _, s := range items {
		parts = append(parts,
			theme.Shortcut.Render(s.Key)+" "+theme.Footer.Render(s.Description))
	}
	content := strings.Join(parts, "  ·  ")
	// truncate to width
	if lipgloss.Width(content) > l.Width {
		// strip middle items
		for len(parts) > 2 {
			parts = parts[:len(parts)-2]
			content = strings.Join(parts, "  ·  ") + "  ·  …"
			if lipgloss.Width(content) <= l.Width {
				break
			}
		}
	}
	pad := l.Width - lipgloss.Width(content)
	if pad < 0 {
		pad = 0
	}
	style := theme.Footer.Width(l.Width)
	return style.Render(strings.Repeat(" ", pad) + content)
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

// RightPane rendert die rechte Pane (Tabs Preview | TOC | Backlinks).
// `tabLabels` enthält alle Tab-Namen in Reihenfolge, `activeTab` ist der Index.
// `tabContents` enthält die Inhalte pro Tab. Aktiver Tab wird hervorgehoben.
func (l *Layout) RightPane(tabLabels []string, activeTab int, tabContents []string,
	theme Theme, isActive bool) string {
	border := theme.InactiveBorder
	if isActive {
		border = theme.ActiveBorder
	}

	innerWidth := l.RightWidth - 2
	if innerWidth < 1 {
		innerWidth = 1
	}
	innerHeight := l.EditorHeight - 3 // 1 line tab-bar + 1 line title + 1 line spacing
	if innerHeight < 1 {
		innerHeight = 1
	}

	// Tab-Bar: "Preview  TOC  Backlinks" mit active highlighted
	var tabParts []string
	for i, lbl := range tabLabels {
		if i == activeTab {
			tabParts = append(tabParts, theme.Shortcut.Render(" "+lbl+" "))
		} else {
			tabParts = append(tabParts, theme.Footer.Render(" "+lbl+" "))
		}
	}
	tabBar := strings.Join(tabParts, " ")

	// Body content (active tab)
	body := ""
	if activeTab >= 0 && activeTab < len(tabContents) {
		body = tabContents[activeTab]
	}

	rendered := lipgloss.JoinVertical(lipgloss.Left,
		tabBar,
		theme.Separator.Width(innerWidth).Render(strings.Repeat("─", innerWidth)),
		lipgloss.NewStyle().Width(innerWidth).Height(innerHeight).Render(body),
	)
	return border.
		Width(l.RightWidth).
		Height(l.EditorHeight).
		Render(rendered)
}

// Compose setzt die 3 vertikalen Panes nebeneinander.
func (l *Layout) Compose(sidebar, editor, right string) string {
	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, editor, right)
}
