// Package markdown implementiert Markdown-Parsing und Highlighting.
package markdown

import (
	"regexp"
	"strings"
)

// Heading repräsentiert eine Markdown-Überschrift.
type Heading struct {
	Level int
	Text  string
	Line  int
}

// HighlightParams sammelt Highlighting-Optionen.
type HighlightParams struct {
	CurrentLine int
}

// Headings extrahiert alle H1-H6-Überschriften aus Markdown-Text.
func Headings(text string) []Heading {
	var hs []Heading
	for i, line := range strings.Split(text, "\n") {
		t := strings.TrimRight(line, " \t\r")
		n := 0
		for _, c := range t {
			if c == '#' {
				n++
				continue
			}
			break
		}
		if n >= 1 && n <= 6 && n < len(t) && (t[n] == ' ' || t[n] == '	') {
			rest := strings.TrimSpace(t[n+1:])
			if rest != "" {
				hs = append(hs, Heading{Level: n, Text: rest, Line: i})
			}
		}
	}
	return hs
}

// HighlightLine formatiert eine einzelne Zeile mit ANSI-Codes.
func HighlightLine(line string, p HighlightParams) string {
	t := strings.TrimRight(line, " \t\r")

	headingLevel := 0
	for _, c := range t {
		if c == '#' {
			headingLevel++
			continue
		}
		break
	}
	if headingLevel >= 1 && headingLevel <= 6 && strings.HasPrefix(t, strings.Repeat("#", headingLevel)+" ") {
		text := strings.TrimSpace(t[headingLevel:])
		return "\x1b[1;3" + headingColor(headingLevel) + "m" +
			strings.Repeat("#", headingLevel) + " " +
			"\x1b[1m" + text + "\x1b[0m"
	}

	if strings.HasPrefix(t, "```") {
		return "\x1b[48;5;240m\x1b[37m" + line + "\x1b[0m"
	}

	if strings.HasPrefix(t, "- ") || strings.HasPrefix(t, "* ") {
		bullet := t[:2]
		rest := t[2:]
		return "\x1b[1;36m" + bullet + "\x1b[0m" + highlightInline(rest)
	}

	if strings.HasPrefix(t, "> ") {
		return "\x1b[2;3m" + line + "\x1b[0m"
	}

	return highlightInline(line)
}

func headingColor(level int) string {
	switch level {
	case 1:
		return "5"
	case 2:
		return "3"
	case 3:
		return "2"
	case 4:
		return "4"
	case 5:
		return "6"
	case 6:
		return "7"
	}
	return "7"
}

func highlightInline(s string) string {
	codeRe := regexp.MustCompile("`([^`]+)`")
	s = codeRe.ReplaceAllString(s, "\x1b[48;5;236m\x1b[37m$1\x1b[0m")

	boldRe := regexp.MustCompile(`\*\*([^*]+)\*\*`)
	s = boldRe.ReplaceAllString(s, "\x1b[1m$1\x1b[22m")

	italicRe := regexp.MustCompile(`\*([^*]+)\*`)
	s = italicRe.ReplaceAllString(s, "\x1b[3m$1\x1b[23m")

	linkRe := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	s = linkRe.ReplaceAllString(s, "\x1b[4;36m$1\x1b[0m (\x1b[2;36m$2\x1b[0m)")

	return s
}
