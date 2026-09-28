// Command mdskim2 ist der Entry-Point für die Terminal-Markdown-Anwendung.
//
// Verwendung (Usage):
//
//	mdskim2 [--help] [--version] [WORKSPACE-PFAD]
//
// Wenn kein WORKSPACE-PFAD angegeben ist, wird das aktuelle Verzeichnis genutzt.
// Drücke Ctrl+Q zum Beenden, F1 für die vollständige Tastatur-Übersicht.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dennis605/mdskim2/internal/app"
)

// version wird zur Build-Zeit via -ldflags gesetzt (z.B. -ldflags "-X main.version=0.2.0").
// Ohne gesetzte Variable wird "dev" angezeigt.
var version = "0.2.0"

const usage = `mdskim2 — Modern Markdown Workspace for the Terminal
Version: %s

Verwendung (Usage):
  mdskim2 [--help] [--version] [WORKSPACE-PFAD]

Wenn kein WORKSPACE-PFAD angegeben ist, wird das aktuelle Verzeichnis genutzt.

Bedienung (alle Tastatur-Shortcuts sind Windows-/Obsidian-artig — KEINE Vim-Modi).
Hinweis: Manche Windows-Shortcuts funktionieren im Terminal nicht
(z.B. Ctrl+Shift+P, Ctrl+Tab). Wir nutzen App-seitige Workarounds
(siehe "Besonderheiten" am Ende).

Navigation & Fokus
  ↑↓ im Tree          Datei wird automatisch geladen + Vorschau
  Enter auf File      Wechselt Focus in den Editor
  Esc                 Zurück zum Tree (verlässt Editor)
  ↑↓ im Editor        Cursor im Buffer bewegen
  F1                  Vollständige Hilfe anzeigen
  Ctrl+Q              Beenden

Pane-Fokus (U5/U6.1)
  F6                  Next Pane (Tree → Editor → Right → TOC → Tree)
  Shift+F6            Previous Pane
  Alt+1 / Alt+2       Focus Tree / Editor
  Alt+3 / Alt+4       Focus Right Pane / Right-Tab-Bar
  Ctrl+Shift+←/→      Pane-Resize / -Switch (Right-Pane)
  Alt+5..Alt+8        Right-Tab: Preview / TOC / Backlinks / Tags

Datei-Operationen (R4, U8.1, U8.3)
  Ctrl+S              Speichern
  Ctrl+T              Neuer Tab (Datei-Picker)
  Ctrl+R              Refresh Workspace (Tree neu laden)
  Ctrl+W              Aktuelle Datei schließen
  Ctrl+Shift+W        Aktuellen Buffer schließen
  Ctrl+P              Quick Open (Datei suchen)
  Ctrl+O              File-Dialog
  Ctrl+N              Neue Datei
  Ctrl+K              Command Palette
  F5                  Refresh Tree
  /                   Tree-Filter (Live-Suche im Sidebar)
  a / Shift+A         Neue Datei / Neuer Ordner (im Tree)
  d / r               Löschen / Umbenennen (im Tree, mit Confirm)

Editor (R4, U3, U4)
  Ctrl+C/X/V          Copy / Cut / Paste
  Ctrl+Z / Y          Undo / Redo
  Ctrl+A              Alles auswählen
  Ctrl+F              Suchen (in Datei)
  Ctrl+H              Suchen und Ersetzen
  Ctrl+B / I          Markdown Bold / Italic
  Ctrl+L / G          Go-to-Line (Direkt-Sprung)
  Ctrl+Shift+R        Read-Only Toggle (Header zeigt [RO])
  Ctrl+Enter          Markdown-Task toggeln ([ ] ↔ [x])

Find / Workspace-Suche (R6, U9.2, U9.3)
  Ctrl+Shift+F        Find in Files (Workspace-Grep)
  Ctrl+Shift+O        Recent Files (letzte 10 Dateien)

Obsidian-Features (U9)
  Ctrl+Shift+D        Daily Notes (heutiges Datum anlegen/öffnen)
  Ctrl+Enter          Task-List-Toggle auf - [ ] / - [x]
  Alt+8               Tag-Panel im RightPane (4. Tab)
  [[                  Wiki-Link-Autocomplete (Popup mit allen .md)

Mouse (U7)
  Klick Sidebar       Datei laden
  Klick Editor        Cursor positionieren
  Klick Right-Pane    Tab wechseln / TOC springen
  Klick Tab-Bar       Tab wechseln

Besonderheiten (Spec-Hard-Verbot)
  Ctrl+Shift+P        Funktioniert in vielen Terminals NICHT
                      (wird als Modifier-only-Sequenz geschluckt).
                      Workaround: Ctrl+K öffnet die Command Palette.
  Ctrl+Tab            Wird vom Terminal als Tab-Sequenz verarbeitet
                      und kann nicht abgefangen werden.
                      Workaround: F6 cycled durch die Panes.
  Vim-Modi            Bewusst NICHT implementiert. Wir nutzen
                      Standard-Shortcuts (Ctrl+S, Ctrl+O, etc.).

Weitere Info: https://github.com/dennis605/mdskim2
`

func main() {
	helpFlag := flag.Bool("help", false, "Zeigt diese Hilfe und beendet das Programm.")
	versionFlag := flag.Bool("version", false, "Zeigt die Version und beendet das Programm.")
	flag.Usage = func() { fmt.Printf(usage, version) }
	flag.Parse()

	if *helpFlag {
		fmt.Printf(usage, version)
		os.Exit(0)
	}
	if *versionFlag {
		fmt.Printf("mdskim2 %s\n", version)
		os.Exit(0)
	}

	workspace := "."
	args := flag.Args()
	if len(args) > 0 {
		workspace = args[0]
	}

	absWorkspace, err := filepath.Abs(workspace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fehler beim Auflösen des Workspace-Pfads %q: %v\n", workspace, err)
		os.Exit(1)
	}

	// Fail fast: Workspace muss existieren UND ein Verzeichnis sein.
	info, err := os.Stat(absWorkspace)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Fehler: Workspace-Pfad existiert nicht: %s\n", absWorkspace)
		} else {
			fmt.Fprintf(os.Stderr, "Fehler beim Prüfen des Workspace-Pfads %q: %v\n", absWorkspace, err)
		}
		fmt.Fprintf(os.Stderr, "Tipp: Lege das Verzeichnis an oder übergebe einen existierenden Pfad.\n")
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Fehler: %q ist eine Datei, kein Verzeichnis.\n", absWorkspace)
		os.Exit(1)
	}

	model := app.New(absWorkspace)
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Fehler: %v\n", err)
		os.Exit(1)
	}
}
