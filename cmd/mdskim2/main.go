// Command mdskim2 ist der Entry-Point für die Terminal-Markdown-Anwendung.
//
// Verwendung (Usage):
//
//	mdskim2 [WORKSPACE-PFAD]
//
// Wenn kein WORKSPACE-PFAD angegeben ist, wird das aktuelle Verzeichnis genutzt.
// Drücke Ctrl+Q zum Beenden, Ctrl+P für Quick Open (kommt in Sprint R8).
package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dennis605/mdskim2/internal/app"
)

const usage = `mdskim2 — Modern Markdown Workspace for the Terminal

Verwendung (Usage):
  mdskim2 [WORKSPACE-PFAD]

Wenn kein WORKSPACE-PFAD angegeben ist, wird das aktuelle Verzeichnis benutzt.

Bedienung (alle Windows-artig, keine Vim-Modi):
  Ctrl+S        Speichern                  (Sprint R4)
  Ctrl+C/X/V    Copy / Cut / Paste         (Sprint R4)
  Ctrl+Z / Y    Undo / Redo                (Sprint R4)
  Ctrl+A        Alles auswählen            (Sprint R4)
  Ctrl+F        Suchen (in Datei)          (Sprint R6)
  Ctrl+H        Suchen und Ersetzen        (Sprint R6)
  Ctrl+P        Quick Open (Datei suchen)  (Sprint R8)
  Ctrl+K        Command Palette            (Sprint R9)
  Ctrl+B / I    Markdown Bold / Italic     (Sprint R4)
  Ctrl+N        Neue Datei                 (Sprint R4)
  Ctrl+W        Aktuelle Datei schließen   (Sprint R8)
  Ctrl+Q        Beenden                    (Sprint R1)

Drücke F1 für Hilfe (Sprint R10).

Weitere Info: https://github.com/dennis605/mdskim2
`

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		fmt.Print(usage)
		os.Exit(0)
	}

	workspace := "."
	if len(os.Args) > 1 {
		workspace = os.Args[1]
	}

	absWorkspace, err := filepath.Abs(workspace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fehler beim Auflösen des Workspace-Pfads: %v\n", err)
		os.Exit(1)
	}

	model := app.New(absWorkspace)
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Fehler: %v\n", err)
		os.Exit(1)
	}
}
