// Package ui definiert Themes, Layout-Primitive und Status-/Footer-Rendering.
package ui

import "github.com/charmbracelet/lipgloss"

// Theme hält alle Style-Definitionen für eine Light- oder Dark-Variante.
type Theme struct {
	Name string

	Border         lipgloss.Style
	ActiveBorder   lipgloss.Style
	InactiveBorder lipgloss.Style

	Header lipgloss.Style

	SidebarLabel   lipgloss.Style
	TreeFile       lipgloss.Style
	TreeDir        lipgloss.Style
	TreeActiveFile lipgloss.Style

	EditorTitle    lipgloss.Style
	EditorBody     lipgloss.Style
	EditorMarkdown lipgloss.Style

	TOCHeading     lipgloss.Style
	TOCActiveEntry lipgloss.Style

	Status    lipgloss.Style
	Footer    lipgloss.Style
	Shortcut  lipgloss.Style
	Separator lipgloss.Style
}

// Light ist das Standard-Theme (für helle Terminals).
func Light() Theme {
	return Theme{
		Name:           "light",
		Border:         lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")),
		ActiveBorder:   lipgloss.NewStyle().Foreground(lipgloss.Color("#007ACC")),
		InactiveBorder: lipgloss.NewStyle().Foreground(lipgloss.Color("#DDDDDD")),
		Header:         lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#007ACC")),
		SidebarLabel:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#666666")),
		TreeFile:       lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")),
		TreeDir:        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#666666")),
		TreeActiveFile: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#007ACC")).Background(lipgloss.Color("#E8F4FD")),
		EditorTitle:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#007ACC")),
		EditorBody:     lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")),
		EditorMarkdown: lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")),
		TOCHeading:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#666666")),
		TOCActiveEntry: lipgloss.NewStyle().Foreground(lipgloss.Color("#007ACC")).Bold(true),
		Status:         lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")),
		Footer:         lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")),
		Shortcut:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#007ACC")),
		Separator:      lipgloss.NewStyle().Foreground(lipgloss.Color("#DDDDDD")),
	}
}

// Dark ist das dunkle Theme.
func Dark() Theme {
	return Theme{
		Name:           "dark",
		Border:         lipgloss.NewStyle().Foreground(lipgloss.Color("#444444")),
		ActiveBorder:   lipgloss.NewStyle().Foreground(lipgloss.Color("#5B9BD5")),
		InactiveBorder: lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")),
		Header:         lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#5B9BD5")),
		SidebarLabel:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#AAAAAA")),
		TreeFile:       lipgloss.NewStyle().Foreground(lipgloss.Color("#EEEEEE")),
		TreeDir:        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#AAAAAA")),
		TreeActiveFile: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#5B9BD5")).Background(lipgloss.Color("#1E2A38")),
		EditorTitle:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#5B9BD5")),
		EditorBody:     lipgloss.NewStyle().Foreground(lipgloss.Color("#EEEEEE")),
		EditorMarkdown: lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")),
		TOCHeading:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#AAAAAA")),
		TOCActiveEntry: lipgloss.NewStyle().Foreground(lipgloss.Color("#5B9BD5")).Bold(true),
		Status:         lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")),
		Footer:         lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")),
		Shortcut:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#5B9BD5")),
		Separator:      lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")),
	}
}
