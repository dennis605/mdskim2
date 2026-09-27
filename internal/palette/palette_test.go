package palette

import (
	"testing"
)

func TestRegisterUndRun(t *testing.T) {
	r := NewRegistry()
	r.Register(Command{Name: "save", Description: "Save current file", Keywords: []string{"write", "store"}})
	r.Register(Command{Name: "open", Description: "Open a file", Keywords: []string{"load", "read"}})

	_, ok := r.Run("save")
	if !ok {
		t.Error("save sollte gefunden werden")
	}
	_, ok = r.Run("nonexistent")
	if ok {
		t.Error("nonexistent sollte nicht gefunden werden")
	}
}

func TestSearch(t *testing.T) {
	r := NewRegistry()
	r.Register(Command{Name: "save", Description: "Save", Keywords: []string{"write"}})
	r.Register(Command{Name: "save-as", Description: "Save As", Keywords: []string{"write"}})
	r.Register(Command{Name: "open", Description: "Open", Keywords: []string{"load"}})

	results := r.Search("save")
	if len(results) != 2 {
		t.Errorf("Search 'save' sollte 2 finden, got %d: %v", len(results), results)
	}

	results = r.Search("wri")
	if len(results) != 2 {
		t.Errorf("Search 'wri' sollte 2 finden, got %d", len(results))
	}
}
