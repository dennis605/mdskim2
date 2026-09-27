package workspace

import (
	"strings"
)

// TreeRenderer rendert einen Tree als String mit Indentation und Glyphs.
type TreeRenderer struct {
	CollapsedDirs map[string]bool
}

// NewTreeRenderer erzeugt einen neuen Renderer (alle Dirs expanded by default).
func NewTreeRenderer() *TreeRenderer {
	return &TreeRenderer{CollapsedDirs: make(map[string]bool)}
}

// Render rendert die FlatList in einen String.
func (r *TreeRenderer) Render(ws *Workspace) string {
	var lines []string
	for _, node := range ws.FlatList(r.CollapsedDirs) {
		lines = append(lines, r.formatNode(node))
	}
	return strings.Join(lines, "\n")
}

// formatNode formatiert eine einzelne Zeile.
func (r *TreeRenderer) formatNode(node *FileNode) string {
	indent := strings.Repeat("  ", node.Depth)

	if node.IsDir {
		glyph := "▾ "
		if r.CollapsedDirs[node.Path] {
			glyph = "▸ "
		}
		return indent + glyph + node.Name
	}

	// Datei: Markdown-Files bekommen ein Glyph, andere einfach indent + name
	if strings.HasSuffix(node.Name, ".md") {
		return indent + "• " + node.Name
	}
	return indent + node.Name
}

// ToggleDir toggelt den Collapsed-Status eines Verzeichnisses.
func (r *TreeRenderer) ToggleDir(path string) {
	if r.CollapsedDirs[path] {
		delete(r.CollapsedDirs, path)
	} else {
		r.CollapsedDirs[path] = true
	}
}

// FormatNode formatiert eine einzelne Node-Zeile (exportiert für App-Layer).
func (r *TreeRenderer) FormatNode(node *FileNode) string {
	return r.formatNode(node)
}

// IsCollapsed prüft, ob ein Verzeichnis kollabiert ist.
func (r *TreeRenderer) IsCollapsed(path string) bool {
	return r.CollapsedDirs[path]
}
