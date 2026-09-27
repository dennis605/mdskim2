package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Shortcut repräsentiert einen anzuzeigenden Tastenkürzel im Footer.
type Shortcut struct {
	Key         string // "Ctrl+S"
	Description string // "Save"
}

// DefaultShortcuts ist die initiale Liste globaler Shortcuts (R4-Stand).
func DefaultShortcuts() []Shortcut {
	return []Shortcut{
		{Key: "Ctrl+S", Description: "Save"},
		{Key: "Ctrl+Z", Description: "Undo"},
		{Key: "Ctrl+Y", Description: "Redo"},
		{Key: "Ctrl+B", Description: "Bold"},
		{Key: "Ctrl+T", Description: "Quick Open"},
		{Key: "Ctrl+P", Description: "Preview"},
		{Key: "Ctrl+K", Description: "Palette"},
		{Key: "Ctrl+Q", Description: "Quit"},
		{Key: "F6", Description: "Focus next pane"},
		{Key: "Alt+1..4", Description: "Jump to pane"},
	}
}

// RenderFooter erzeugt die untere Shortcut-Leiste.
func (l *Layout) RenderFooter(shortcuts []Shortcut, theme Theme) string {
	parts := make([]string, 0, len(shortcuts))
	for _, s := range shortcuts {
		key := theme.Shortcut.Render(s.Key)
		desc := theme.Footer.Render(" " + s.Description)
		parts = append(parts, key+desc)
	}

	content := lipgloss.JoinHorizontal(lipgloss.Top, joinWithSep(parts, " │ "))
	padding := l.Width - lipgloss.Width(content)
	if padding < 0 {
		padding = 0
	}

	return strings.Repeat(" ", padding) + content
}

func joinWithSep(parts []string, sep string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += sep
		}
		out += p
	}
	return out
}
