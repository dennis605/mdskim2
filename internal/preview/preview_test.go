package preview

import (
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	md := "# Title\n\nSome **bold** text.\n"
	out, err := Render(md)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(out, "Title") {
		t.Errorf("Render sollte 'Title' enthalten: %q", out)
	}
}

func TestRenderPlain(t *testing.T) {
	out := RenderPlain("# Hello")
	if !strings.Contains(out, "Hello") {
		t.Errorf("RenderPlain: %q", out)
	}
}
