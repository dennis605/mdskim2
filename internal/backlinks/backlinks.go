// Package backlinks collects [[Wiki-Links]] that point to a given file
// from anywhere in the workspace. Used by the right-pane "Backlinks" tab
// in mdskim2 (matches Python mdskim's BacklinksPanel feature).
package backlinks

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Entry describes one backlink: the file that contains the link,
// the line number (1-indexed) where it appears, and the link target.
type Entry struct {
	SourceFile string // absolute path of file containing the link
	SourceLine int    // 1-indexed line number
	LinkTarget string // the [[link]] text
	LinkAlias  string // the [[link|alias]] alias, if any
}

// Collect scans the workspace rooted at root for files containing
// [[wiki-links]] that point to target. Returns sorted entries
// (by SourceFile then SourceLine).
func Collect(root, target string) ([]Entry, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	// Determine the basename and full-relative form of the target.
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		targetAbs = target
	}
	targetBase := filepath.Base(targetAbs)
	targetNoExt := strings.TrimSuffix(targetBase, filepath.Ext(targetBase))

	linkRe := regexp.MustCompile(`\[\[([^\]\|]+)(?:\|([^\]]+))?\]\]`)

	var out []Entry
	err = filepath.WalkDir(absRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil // skip broken paths
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "node_modules" || strings.HasPrefix(name, ".") {
				if path != absRoot {
					return filepath.SkipDir
				}
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".md" && ext != ".markdown" {
			return nil
		}
		// Skip the target file itself
		pa, _ := filepath.Abs(path)
		if pa == targetAbs {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			matches := linkRe.FindAllStringSubmatch(line, -1)
			for _, m := range matches {
				link := strings.TrimSpace(m[1])
				linkBase := filepath.Base(link)
				linkNoExt := strings.TrimSuffix(linkBase, filepath.Ext(linkBase))
				if linkBase == targetBase || linkNoExt == targetNoExt || link == targetNoExt {
					entry := Entry{
						SourceFile: path,
						SourceLine: i + 1,
						LinkTarget: link,
						LinkAlias:  strings.TrimSpace(m[2]),
					}
					out = append(out, entry)
				}
			}
		}
		return nil
	})
	if err != nil {
		return out, err
	}
	// Sort: by SourceFile then SourceLine
	sort.Slice(out, func(i, j int) bool {
		if out[i].SourceFile != out[j].SourceFile {
			return out[i].SourceFile < out[j].SourceFile
		}
		return out[i].SourceLine < out[j].SourceLine
	})
	return out, nil
}
