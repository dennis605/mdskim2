// Package tabs implementiert Multi-File Tab-Management.
package tabs

import (
	"path/filepath"
	"sort"
	"sync"

	"github.com/dennis605/mdskim2/internal/editor"
)

// Tab repräsentiert einen geöffneten Buffer.
type Tab struct {
	Path   string
	Name   string
	Buffer *editor.Buffer
}

// Manager hält alle offenen Tabs und den aktiven Tab.
type Manager struct {
	mu     sync.RWMutex
	all    []*Tab
	active int
}

// NewManager erzeugt einen leeren Tab-Manager.
func NewManager() *Manager {
	return &Manager{active: -1}
}

// Open öffnet eine Datei in einem Tab (oder fokussiert bestehenden Tab).
func (m *Manager) Open(path string, buf *editor.Buffer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Suche bestehenden Tab
	for i, t := range m.all {
		if t.Path == path {
			m.active = i
			return
		}
	}
	name := filepath.Base(path)
	t := &Tab{Path: path, Name: name, Buffer: buf}
	m.all = append(m.all, t)
	m.active = len(m.all) - 1
}

// Active returns the active tab.
func (m *Manager) Active() *Tab {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.active < 0 || m.active >= len(m.all) {
		return nil
	}
	return m.all[m.active]
}

// Close schließt den aktiven Tab.
func (m *Manager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.active < 0 || m.active >= len(m.all) {
		return
	}
	m.all = append(m.all[:m.active], m.all[m.active+1:]...)
	if len(m.all) == 0 {
		m.active = -1
	} else if m.active >= len(m.all) {
		m.active = len(m.all) - 1
	}
}

// Next wechselt zum nächsten Tab.
func (m *Manager) Next() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.all) == 0 {
		return
	}
	m.active = (m.active + 1) % len(m.all)
}

// Prev wechselt zum vorherigen Tab.
func (m *Manager) Prev() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.all) == 0 {
		return
	}
	m.active = (m.active - 1 + len(m.all)) % len(m.all)
}

// Len returns die Anzahl der Tabs.
func (m *Manager) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.all)
}

// All returns alle Tabs (read-only copy).
func (m *Manager) All() []*Tab {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*Tab, len(m.all))
	copy(out, m.all)
	return out
}

// SetActive setzt den aktiven Tab nach Index.
func (m *Manager) SetActive(idx int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if idx >= 0 && idx < len(m.all) {
		m.active = idx
	}
}

// Search sucht Tabs nach Pfad-Substring.
func (m *Manager) Search(query string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	sort.Slice(m.all, func(i, j int) bool { return m.all[i].Name < m.all[j].Name })
	var paths []string
	for _, t := range m.all {
		if query == "" {
			paths = append(paths, t.Path)
		} else if contains(t.Name, query) || contains(t.Path, query) {
			paths = append(paths, t.Path)
		}
	}
	return paths
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
