// Package workspace verwaltet das geöffnete Verzeichnis, scannt es
// und baut den FileTree. fsnotify-Watcher kommt in R3.
package workspace

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FileNode repräsentiert einen Eintrag im Tree.
type FileNode struct {
	Name     string
	Path     string
	IsDir    bool
	Children []*FileNode
	Parent   *FileNode
	Depth    int
}

// Workspace kapselt den Root-Pfad und den Root-Tree.
type Workspace struct {
	RootPath string
	Root     *FileNode
}

// Load scannt ein Verzeichnis rekursiv (ohne .git, node_modules, etc.)
// und gibt einen Workspace mit Tree zurück.
func Load(rootPath string) (*Workspace, error) {
	abs, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}

	root := &FileNode{
		Name:  filepath.Base(abs),
		Path:  abs,
		IsDir: true,
		Depth: 0,
	}

	if err := walk(abs, root, 0); err != nil {
		return nil, err
	}

	return &Workspace{RootPath: abs, Root: root}, nil
}

func walk(path string, parent *FileNode, depth int) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		// Verzeichnis nicht lesbar — überspringen, kein Hard-Fail
		return nil
	}

	// Sortieren: Verzeichnisse zuerst, dann alphabetisch (case-insensitive)
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name())
	})

	for _, entry := range entries {
		name := entry.Name()
		if name == ".git" || name == "node_modules" || name == ".DS_Store" {
			continue
		}
		if strings.HasPrefix(name, ".") {
			continue
		}

		child := &FileNode{
			Name:   name,
			Path:   filepath.Join(path, name),
			IsDir:  entry.IsDir(),
			Parent: parent,
			Depth:  depth + 1,
		}
		parent.Children = append(parent.Children, child)

		if entry.IsDir() {
			if err := walk(child.Path, child, depth+1); err != nil {
				return err
			}
		}
	}
	return nil
}

// FlatList gibt den Tree als flache Liste zurück. collapsedDirs hält die
// Pfade der kollabierten Verzeichnisse (relativ/absolut zum Root).
func (w *Workspace) FlatList(collapsedDirs map[string]bool) []*FileNode {
	if w == nil || w.Root == nil {
		return nil
	}
	var list []*FileNode
	flattenInto(w.Root, &list, collapsedDirs, true)
	return list
}

func flattenInto(node *FileNode, list *[]*FileNode, collapsedDirs map[string]bool, includeRoot bool) {
	if includeRoot {
		*list = append(*list, node)
	}
	if !node.IsDir {
		return
	}
	if collapsedDirs[node.Path] {
		return
	}
	for _, child := range node.Children {
		flattenInto(child, list, collapsedDirs, true)
	}
}

// FindByPath findet einen Node anhand seines vollen Pfads.
func (w *Workspace) FindByPath(path string) *FileNode {
	if w == nil || w.Root == nil {
		return nil
	}
	return findInTree(w.Root, path)
}

func findInTree(node *FileNode, path string) *FileNode {
	if node.Path == path {
		return node
	}
	for _, child := range node.Children {
		if found := findInTree(child, path); found != nil {
			return found
		}
	}
	return nil
}

// TotalFiles zählt alle Dateien (nicht Verzeichnisse).
func (w *Workspace) TotalFiles() int {
	if w == nil || w.Root == nil {
		return 0
	}
	return countFiles(w.Root)
}

func countFiles(node *FileNode) int {
	if !node.IsDir {
		return 1
	}
	n := 0
	for _, child := range node.Children {
		n += countFiles(child)
	}
	return n
}
