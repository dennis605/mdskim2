package smoke

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dennis605/mdskim2/internal/app"
)

var updateGolden = flag.Bool("update", false, "update golden snapshot files")

func TestSnapshotMatchesGolden(t *testing.T) {
	goldenPath := filepath.Join("testdata", "snapshot-boot.txt")

	if *updateGolden {
		m := app.New("/tmp/sample-workspace")
		mode := m.View()
		if err := os.WriteFile(goldenPath, []byte(mode), 0644); err != nil {
			t.Fatalf("golden write: %v", err)
		}
		t.Logf("Golden file aktualisiert: %s", goldenPath)
		return
	}

	// Fixture anlegen
	if err := os.MkdirAll("testdata/sample-workspace", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	dummyMD := filepath.Join("testdata", "sample-workspace", "README.md")
	if _, err := os.Stat(dummyMD); os.IsNotExist(err) {
		if err := os.WriteFile(dummyMD, []byte("# Test\n"), 0644); err != nil {
			t.Fatalf("write dummy: %v", err)
		}
	}

	wantBytes, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden: %v — Run with -update to create", err)
	}

	m := app.New("/tmp/sample-workspace")
	got := m.View()

	// Schnelle Vergleich: View enthält alle Golden-Lines als Substrings
	for _, line := range strings.Split(string(wantBytes), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if !strings.Contains(got, line) {
			t.Errorf("Golden-Line fehlt in View: %q", line)
		}
	}
}
