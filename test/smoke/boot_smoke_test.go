package smoke

import (
	"strings"
	"testing"

	"github.com/dennis605/mdskim2/internal/app"
)

func TestBootViewContainsExpectedMarkers(t *testing.T) {
	m := app.New("/tmp/test-workspace")
	mode := m.View()

	mustContain := []string{
		"mdskim2",
		"FILES",
		"Preview",
		"TOC",
		"Backlinks",
		"Ctrl+S",
		"Ctrl+Q",
		"/tmp/test-workspace",
	}
	for _, m := range mustContain {
		if !strings.Contains(mode, m) {
			t.Errorf("erwartet %q in View, got: %s", m, mode)
		}
	}
}
