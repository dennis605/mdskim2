// Package preview rendert Markdown als ANSI-Terminal via glamour.
package preview

import (
	"github.com/charmbracelet/glamour"
)

// Render rendert Markdown-Text als formatierten Terminal-String.
func Render(markdown string) (string, error) {
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("notty"),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return "", err
	}
	defer r.Close()
	return r.Render(markdown)
}

// RenderPlain rendert Markdown ohne Style (Fallback).
func RenderPlain(text string) string {
	return text
}
