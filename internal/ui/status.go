package ui

import "github.com/charmbracelet/lipgloss"

// StatusInfo hält den Inhalt für die Status-Bar.
type StatusInfo struct {
	Workspace string
	File      string // aktuell geöffnete Datei (leer = keine)
	Lines     int    // Zeilen in aktueller Datei
	Words     int    // Wörter in aktueller Datei
	Encoding  string // UTF-8
	Modified  bool   // hat ungespeicherte Änderungen
	Mode      string // "EDIT" | "PREV" | "SPLIT" | "BOOT"
	Version   string // z.B. "R1: Bootstrap"
}

// Render erzeugt die Status-Bar-String.
func (l *Layout) RenderStatus(info StatusInfo, theme Theme) string {
	right := info.Mode + " · " + info.Version
	if info.File != "" {
		right = info.File
		if info.Modified {
			right += " · ◆ modified"
		}
		right += " · " + info.Encoding
		if info.Lines > 0 {
			right += " · " + itoa(info.Lines) + " lines"
		}
		if info.Words > 0 {
			right += " · " + itoa(info.Words) + " W"
		}
		right += " · " + info.Mode + " · " + info.Version
	}

	left := info.Workspace
	if left == "" {
		left = "(kein Workspace)"
	}

	width := l.Width - lipgloss.Width(right) - 2
	if width < 1 {
		width = 1
	}
	leftRendered := theme.Status.Width(width).Render(truncate(left, width))
	rightRendered := theme.Status.Render(right)

	separator := theme.Separator.Width(l.Width).Render("─")
	return lipgloss.JoinVertical(lipgloss.Left, separator,
		lipgloss.JoinHorizontal(lipgloss.Top, leftRendered, " ", rightRendered))
}

// itoa ersetzt strconv.Itoa, um Import-Footprint klein zu halten.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	negative := n < 0
	if negative {
		n = -n
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if negative {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}

// truncate kürzt einen String auf max n Zeichen.
func truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n-1]) + "…"
}
