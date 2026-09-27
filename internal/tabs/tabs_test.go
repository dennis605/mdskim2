package tabs

import (
	"testing"

	"github.com/dennis605/mdskim2/internal/editor"
)

func TestOpenUndActive(t *testing.T) {
	m := NewManager()
	if m.Active() != nil {
		t.Error("Empty manager sollte nil active")
	}

	buf := editor.NewEmpty()
	m.Open("/tmp/a.md", buf)
	if m.Active() == nil || m.Active().Path != "/tmp/a.md" {
		t.Errorf("Active sollte /tmp/a.md sein, got %v", m.Active())
	}
}

func TestOpenDeduplicates(t *testing.T) {
	m := NewManager()
	m.Open("/tmp/a.md", editor.NewEmpty())
	m.Open("/tmp/a.md", editor.NewEmpty())
	if m.Len() != 1 {
		t.Errorf("Deduplicate sollte 1 Tab sein, got %d", m.Len())
	}
}

func TestClose(t *testing.T) {
	m := NewManager()
	m.Open("/tmp/a.md", editor.NewEmpty())
	m.Open("/tmp/b.md", editor.NewEmpty())
	m.Open("/tmp/c.md", editor.NewEmpty())
	m.SetActive(0) // a
	m.Close()
	if m.Len() != 2 {
		t.Errorf("Close sollte 2 Tabs lassen, got %d", m.Len())
	}
	if m.Active().Path != "/tmp/b.md" {
		t.Errorf("Active sollte sich verschoben haben, got %s", m.Active().Path)
	}
}

func TestNextPrev(t *testing.T) {
	m := NewManager()
	m.Open("/tmp/a.md", editor.NewEmpty())
	m.Open("/tmp/b.md", editor.NewEmpty())
	m.SetActive(0)
	m.Next()
	if m.Active().Path != "/tmp/b.md" {
		t.Errorf("Next sollte /tmp/b.md aktivieren, got %s", m.Active().Path)
	}
	m.Next()
	if m.Active().Path != "/tmp/a.md" {
		t.Errorf("Next (wrap) sollte /tmp/a.md, got %s", m.Active().Path)
	}
	m.Prev()
	if m.Active().Path != "/tmp/b.md" {
		t.Errorf("Prev sollte /tmp/b.md, got %s", m.Active().Path)
	}
}

func TestSearch(t *testing.T) {
	m := NewManager()
	m.Open("/tmp/foo.md", editor.NewEmpty())
	m.Open("/tmp/bar.md", editor.NewEmpty())
	m.Open("/tmp/baz.md", editor.NewEmpty())

	results := m.Search("ba")
	if len(results) != 2 {
		t.Errorf("Search 'ba' sollte 2 finden, got %d", len(results))
	}
}
