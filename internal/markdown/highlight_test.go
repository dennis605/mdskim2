package markdown

import (
	"strings"
	"testing"
)

func TestHeadings(t *testing.T) {
	text := "# H1\n\nPara\n\n## H2\n\nMehr\n\n### H3\n"
	hs := Headings(text)
	if len(hs) != 3 {
		t.Fatalf("erwartet 3 Headings, got %d", len(hs))
	}
	if hs[0].Level != 1 || hs[0].Text != "H1" {
		t.Errorf("Heading 0 falsch: %+v", hs[0])
	}
	if hs[1].Level != 2 || hs[1].Text != "H2" {
		t.Errorf("Heading 1 falsch: %+v", hs[1])
	}
	if hs[2].Level != 3 || hs[2].Text != "H3" {
		t.Errorf("Heading 2 falsch: %+v", hs[2])
	}
}

func TestHeadingsIgnoriertFalschesHeading(t *testing.T) {
	text := "#H1 ohne Leerzeichen\n# H2 mit"
	hs := Headings(text)
	if len(hs) != 1 || hs[0].Text != "H2 mit" {
		t.Errorf("erwartet nur H2, got %v", hs)
	}
}

func TestHeadingsLeererText(t *testing.T) {
	hs := Headings("")
	if len(hs) != 0 {
		t.Errorf("erwartet 0 Headings, got %d", len(hs))
	}
}

func TestHighlightLineHeading(t *testing.T) {
	p := HighlightParams{}
	line := HighlightLine("# Title", p)
	if !strings.Contains(line, "Title") {
		t.Errorf("HighlightLine sollte 'Title' enthalten, got %q", line)
	}
	if !strings.HasPrefix(line, "\x1b[") {
		t.Errorf("HighlightLine sollte mit ANSI starten, got %q", line[:min(20, len(line))])
	}
}

func TestHighlightLineBold(t *testing.T) {
	p := HighlightParams{}
	line := HighlightLine("Some **bold** text", p)
	if !strings.Contains(line, "\x1b[1mbold\x1b[22m") {
		t.Errorf("Bold-Span fehlt: %q", line)
	}
}

func TestHighlightLineItalic(t *testing.T) {
	p := HighlightParams{}
	line := HighlightLine("Some *italic* text", p)
	if !strings.Contains(line, "\x1b[3mitalic\x1b[23m") {
		t.Errorf("Italic-Span fehlt: %q", line)
	}
}

func TestHighlightLineCode(t *testing.T) {
	p := HighlightParams{}
	line := HighlightLine("`inline code`", p)
	if !strings.Contains(line, "inline code") {
		t.Errorf("Code-Span fehlt: %q", line)
	}
}

func TestHighlightLineLink(t *testing.T) {
	p := HighlightParams{}
	line := HighlightLine("[click](https://example.com)", p)
	if !strings.Contains(line, "click") || !strings.Contains(line, "https://example.com") {
		t.Errorf("Link fehlt: %q", line)
	}
}

func TestHighlightLineList(t *testing.T) {
	p := HighlightParams{}
	line := HighlightLine("- Item", p)
	if !strings.Contains(line, "-") {
		t.Errorf("Bullet fehlt: %q", line)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestHighlightLinesCurrentLineMarker(t *testing.T) {
	lines := []string{"# Hello", "World", "**Bold**"}
	out := HighlightLines(lines, HighlightParams{CurrentLine: 1})
	if len(out) != 3 {
		t.Fatalf("erwartet 3 Zeilen, got %d", len(out))
	}
	// Line 1 (idx 0) should NOT have ▶ marker
	if strings.Contains(out[0], "▶") {
		t.Errorf("Line 0 sollte keinen ▶ Marker haben, got: %s", out[0])
	}
	// Line 2 (idx 1) should have ▶ marker
	if !strings.Contains(out[1], "▶") {
		t.Errorf("Line 1 sollte ▶ Marker haben, got: %s", out[1])
	}
}
