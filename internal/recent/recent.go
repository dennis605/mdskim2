package recent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const MaxEntries = 10

// File repräsentiert einen Eintrag in der Recent-Liste.
type File struct {
	Path    string `json:"path"`
	When    int64  `json:"when"`
	Display string `json:"display"`
}

// List persistiert die letzten N Dateien.
type List struct {
	mu       sync.Mutex
	Files    []File `json:"files"`
	filePath string
}

// New lädt oder erstellt die Recent-Liste am Standardspeicherort.
func New() *List {
	dir := configDir()
	path := filepath.Join(dir, "recent.json")
	l := &List{filePath: path}
	l.load()
	return l
}

func configDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".mdskim2")
}

// Add fügt einen neuen Eintrag oben in die Liste ein (most-recent-first).
func (l *List) Add(path string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now().Unix()
	// Dedup: remove existing
	out := []File{{Path: path, When: now, Display: filepath.Base(path)}}
	seen := map[string]bool{path: true}
	for _, f := range l.Files {
		if f.Path == path {
			continue
		}
		if !seen[f.Path] {
			seen[f.Path] = true
			out = append(out, f)
		}
		if len(out) >= MaxEntries {
			break
		}
	}
	l.Files = out
	l.save()
	_ = filepath.Base
}

func (l *List) All() []File {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]File, len(l.Files))
	copy(out, l.Files)
	return out
}

func (l *List) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Files = nil
	l.save()
}

func (l *List) load() {
	data, err := os.ReadFile(l.filePath)
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, l)
}

func (l *List) save() {
	if err := os.MkdirAll(filepath.Dir(l.filePath), 0o755); err != nil {
		return
	}
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(l.filePath, data, 0o644)
}
