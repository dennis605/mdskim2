// Package search implementiert Text-Suche und Ersetzen im Buffer.
package search

import "strings"

// Match repräsentiert eine gefundene Stelle.
type Match struct {
	Line int
	Col  int
	End  int
	Text string
}

// Options konfiguriert eine Suche.
type Options struct {
	CaseSensitive bool
	WholeWord     bool
	Regex         bool
}

// Find sucht `pattern` in `lines` mit den Optionen.
// Wenn `pattern` leer ist, gibt nil zurück (keine Suche).
func Find(lines []string, pattern string, opt Options) []Match {
	if pattern == "" {
		return nil
	}
	var ms []Match
	for i, line := range lines {
		var text string
		var pat string
		if !opt.CaseSensitive {
			text = strings.ToLower(line)
			pat = strings.ToLower(pattern)
		} else {
			text = line
			pat = pattern
		}
		from := 0
		for {
			idx := strings.Index(text[from:], pat)
			if idx < 0 {
				break
			}
			matchCol := from + idx
			matchEnd := matchCol + len(pattern)
			// WholeWord check
			if opt.WholeWord {
				if (matchCol > 0 && isWordChar(rune(line[matchCol-1]))) ||
					(matchEnd < len(line) && isWordChar(rune(line[matchEnd]))) {
					from = matchCol + 1
					continue
				}
			}
			ms = append(ms, Match{
				Line: i,
				Col:  matchCol,
				End:  matchEnd,
				Text: line[matchCol:matchEnd],
			})
			from = matchCol + 1
		}
	}
	return ms
}

func isWordChar(r rune) bool {
	if r >= 'a' && r <= 'z' {
		return true
	}
	if r >= 'A' && r <= 'Z' {
		return true
	}
	if r >= '0' && r <= '9' {
		return true
	}
	return r == '_'
}

// ReplaceAll ersetzt alle Vorkommen in lines.
// Gibt (neue-lines, anzahl-ersetzungen) zurück.
func ReplaceAll(lines []string, pattern, replacement string, opt Options) ([]string, int) {
	if pattern == "" {
		return lines, 0
	}
	count := 0
	out := make([]string, len(lines))
	for i, line := range lines {
		var text, pat string
		if !opt.CaseSensitive {
			text = strings.ToLower(line)
			pat = strings.ToLower(pattern)
		} else {
			text = line
			pat = pattern
		}
		result := ""
		from := 0
		for {
			idx := strings.Index(text[from:], pat)
			if idx < 0 {
				result += line[from:]
				break
			}
			matchStart := from + idx
			matchEnd := matchStart + len(pattern)
			if opt.WholeWord {
				if (matchStart > 0 && isWordChar(rune(line[matchStart-1]))) ||
					(matchEnd < len(line) && isWordChar(rune(line[matchEnd]))) {
					result += line[from:matchEnd]
					from = matchEnd
					continue
				}
			}
			result += line[from:matchStart] + replacement
			count++
			from = matchEnd
		}
		out[i] = result
	}
	return out, count
}
