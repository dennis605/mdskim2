# mdskim2

> **Modern Markdown Workspace for the Terminal** — ein TUI-Workspace für strukturierte
> Markdown-Dokumente. Inspiriert von Obsidian + VS Code, gebaut als Single-Binary in Go.

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Bubble Tea](https://img.shields.io/badge/Bubble%20Tea-v1.3-FF69B4)](https://github.com/charmbracelet/bubbletea)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## Was ist mdskim2?

`mdskim2` ist ein Terminal-UI-Markdown-Workspace mit:

- **Workspace-Tree** mit Datei-Operationen (anlegen, umbenennen, löschen, filtern)
- **Markdown-Editor** mit Syntax-Highlighting, Multi-Line-Selection, Auto-Pair,
  Word-Navigation, Auto-Indent, Go-to-Line und Read-Only-Mode
- **Live-Preview**, **TOC** und **Backlinks-Panel** als Tabs im rechten Bereich
- **Tag-Panel** — alle `#tag` aus dem Workspace mit Häufigkeits-Sortierung
- **Wiki-Link-Autocomplete** beim Tippen von `[[`
- **Daily-Notes** mit `Ctrl+Shift+D` (`<workspace>/Daily Notes/YYYY-MM-DD.md`)
- **Find-in-Files** mit `Ctrl+Shift+F` (Workspace-Grep mit Result-Navigation)
- **Recent-Files** mit `Ctrl+Shift+O`
- **Quick-Open** mit `Ctrl+P`, **Command-Palette** mit `Ctrl+K`
- **Tree-Filter** mit `/` (Live-Suche im Sidebar)
- **Task-Toggle** mit `Ctrl+Enter` (Markdown `- [ ]` ↔ `- [x]`)
- **Auto-Pair** für `()`, `[]`, `{}`, `""`
- **Mouse-Support** für Klicks in Sidebar, Editor, RightPane und TOC
- **Auto-Save** alle 30 Sekunden (`[AUTO-SAVED HH:MM:SS]` im Header)

## Installation

### go install (empfohlen)

```bash
go install github.com/dennis605/mdskim2/cmd/mdskim2@latest
```

Das Binary landet in `$GOPATH/bin/mdskim2` (üblicherweise `~/go/bin/`).
Stelle sicher, dass `~/go/bin` in deinem `PATH` ist.

### go run (zum Testen)

```bash
cd /Users/dennisschonig/projects/mdskim2
go run ./cmd/mdskim2 /pfad/zum/workspace
```

### Aus Source bauen

```bash
git clone https://github.com/dennis605/mdskim2.git
cd mdskim2
go build -o mdskim2 ./cmd/mdskim2
./mdskim2 ./demo/Test\ Engineering
```

## Schnellstart

```bash
# Neuen Workspace anlegen und öffnen
mkdir ~/notes && ~/go/bin/mdskim2 ~/notes

# Demo-Workspace (im Repo) anschauen
cd /Users/dennisschonig/projects/mdskim2
go run ./cmd/mdskim2 "./demo/Test Engineering"
```

Beim Start:
1. **↑↓** im Tree navigieren — die Datei wird automatisch geladen und im Preview angezeigt.
2. **Enter** wechselt den Focus in den Editor.
3. **Ctrl+S** speichert die Änderungen.
4. **Ctrl+Q** beendet die App.

## Tastatur-Übersicht

Die vollständige Liste findest du mit `mdskim2 --help` oder `F1` in der App.

### Navigation & Pane-Focus

| Shortcut | Action |
|----------|--------|
| `↑↓` (Tree) | Datei automatisch laden + Preview |
| `Enter` (Tree) | Focus in den Editor |
| `Esc` | Zurück zum Tree |
| `F6` / `Shift+F6` | Next / Previous Pane |
| `Alt+1..4` | Tree / Editor / RightPane / TabBar |
| `Alt+5..8` | Right-Tab: Preview / TOC / Backlinks / Tags |
| `Ctrl+Shift+←/→` | Pane-Navigation (RightPane) |

### Datei-Operationen

| Shortcut | Action |
|----------|--------|
| `Ctrl+S` | Speichern |
| `Ctrl+T` | Neuer Tab |
| `Ctrl+W` / `Ctrl+Shift+W` | Datei / Buffer schließen |
| `Ctrl+P` | Quick Open (Datei suchen) |
| `Ctrl+K` | Command Palette |
| `Ctrl+R` / `F5` | Refresh Workspace |
| `/` | Tree-Filter |
| `a` / `Shift+A` | Neue Datei / Ordner (im Tree) |
| `d` / `r` | Löschen / Umbenennen (im Tree) |

### Editor

| Shortcut | Action |
|----------|--------|
| `Ctrl+C/X/V` | Copy / Cut / Paste |
| `Ctrl+Z/Y` | Undo / Redo |
| `Ctrl+A` | Alles auswählen |
| `Ctrl+F` / `Ctrl+H` | Suchen / Ersetzen |
| `Ctrl+B/I` | Markdown Bold / Italic |
| `Ctrl+L/G` | Go-to-Line |
| `Ctrl+Shift+R` | Read-Only Toggle |
| `Ctrl+Enter` | Task-List toggeln |

### Workspace-Suche & Obsidian-Features

| Shortcut | Action |
|----------|--------|
| `Ctrl+Shift+F` | Find in Files (Workspace-Grep) |
| `Ctrl+Shift+O` | Recent Files |
| `Ctrl+Shift+D` | Daily Notes |
| `[[` | Wiki-Link-Autocomplete |
| `Alt+8` | Tag-Panel |

## Besonderheiten — warum manche Shortcuts anders sind

### Spec-Hard-Verbot: keine stillen Vim-Substitute

Manche Standard-Shortcuts aus Desktop-Apps funktionieren in Terminals nicht,
weil das Terminal sie vor der Anwendung schluckt. Wir ersetzen sie **nie**
stillschweigend durch Vim-Keys, sondern durch dokumentierte App-seitige Workarounds:

| Gewünscht | Problem | Workaround in mdskim2 |
|-----------|---------|----------------------|
| `Ctrl+Shift+P` (VS Code) | Wird vom Terminal nicht als Modifier-Sequenz erkannt | `Ctrl+K` (Command Palette) |
| `Ctrl+Tab` (Tab-Switch) | Terminal verarbeitet das als Tabulator-Zeichen | `F6` cycled durch Panes |
| `Ctrl+Click` (Link-Open) | Bubble Tea kann Ctrl-Klicks nicht zuverlässig | `Enter` auf Wiki-Link im Editor |

### Kein Vim-Mode

`mdskim2` ist **kein** Vim. Wir nutzen Standard-Shortcuts (`Ctrl+S`, `Ctrl+O`, …)
analog zu Obsidian / VS Code / Sublime. `h/j/k/l`, `:w`, `:q` etc. funktionieren
nicht. Diese Entscheidung ist explizit und wird nicht geändert.

### Ctrl+Enter und Ctrl+J

`Ctrl+Enter` wird in den meisten Terminals als `Ctrl+J` (= LF, ASCII 10)
gesendet. Wir behandeln beide gleichwertig.

## Konfiguration

`mdskim2` legt seinen persistenten State in `~/.mdskim2/` ab:

```
~/.mdskim2/
└── recent.json     # MRU-Liste der letzten 10 geöffneten Dateien
```

Aktuell wird nur `recent.json` persistiert. Weitere Einstellungen (Theme,
Sidebar-Breite, Schriftgröße, Key-Rebindings) sind für eine spätere Version
geplant und werden in `~/.mdskim2/config.toml` landen.

## CLI-Optionen

```bash
mdskim2 [--help] [--version] [WORKSPACE-PFAD]
```

| Flag | Effect |
|------|--------|
| `--help`, `-h` | Zeigt die vollständige Hilfe |
| `--version` | Zeigt die Version und beendet |
| `WORKSPACE-PFAD` | Optional. Default: aktuelles Verzeichnis |

Wenn `WORKSPACE-PFAD` nicht existiert oder kein Verzeichnis ist, bricht
`mdskim2` mit einer freundlichen Fehlermeldung und Exit-Code 1 ab, statt
später mit `could not open /dev/tty` zu scheitern.

## Entwicklung

### Tests

```bash
go test ./...                       # alle Tests
go test ./test/smoke/... -v         # Smoke-Tests mit verbose Output
go test -run TestU9 ./test/smoke/   # Nur U9-Sprints
go test -coverprofile=c.out ./...   # Coverage-Profiling
go tool cover -html=c.out           # Coverage als HTML
```

### 7-Gate-Verifikation

Vor jedem Commit läuft das volle 7-Gate-Setup:

| Gate | Command |
|------|---------|
| Build | `go build ./...` |
| Tests | `go test ./... -count=1` |
| Vet | `go vet ./...` |
| Format | `gofmt -l .` (leer = OK) |
| debug-view | `go run ./cmd/debug-view <workspace>` (nicht-interaktive Frame-Ausgabe) |
| Snapshot | `go test ./test/smoke/... -run TestSnapshot` |
| Binary | `go build -o /tmp/mdskim2 ./cmd/mdskim2` |

### Architecture

```
cmd/
  mdskim2/         Entry-Point (CLI, --help, --version, Fail-Fast-Validation)
  debug-view/      Nicht-interaktive Frame-Dump für CI / Snapshots

internal/
  app/             Bubble Tea Root-Model (Init/Update/View + Keymap)
  editor/          Buffer + Undo/Redo + Selection + History
  workspace/       File-Tree-Walk + Flat-List für Sidebar
  markdown/        Highlight, Tag-Extraktion, TOC
  preview/         Glamour-basiertes Markdown-Rendering
  grep/            Workspace-weite Suche
  search/          In-Datei Suche + Replace
  tabs/            Tab-Manager
  recent/          MRU-Liste auf Disk
  backlinks/       [[Wiki-Link]]-Scanner
  ui/              Layout-Primitiven, Theme, Toolbar

test/
  smoke/           End-to-End-Tests via Bubble-Tea Update-Dispatching
```

### Sprint-Historie

Die Entwicklung lief in zwei Phasen:

**MVP (Sprints R1–R10):** Workspace, Editor, Save, Search, Preview, Highlight,
Quick-Open, Command-Palette, Tabs.

**UX-Polish (Sprints U1–U9.5):**
- **U1–U2:** Auto-Open, Cursor-Visibility, Focus-Modell, Zeilennummern
- **U3:** Selection, Word-Nav, Auto-Indent, Line-Ops, Go-to-Line, Wrap, Scroll
- **U4:** Read-Only, Whitespace-Indicator, Auto-Save, Recent-Files
- **U5:** Pane-Focus-Switching (F6, Alt+1..4)
- **U6:** UI-Revamp (Python-mdskim-Look), Backlinks, RightPane-Tabs
- **U6.1:** Ctrl+Shift+Arrow Pane-Navigation
- **U7:** Mouse-Handling
- **U8.1–U8.5:** Tree-File-Ops, Auto-Pair, Tree-Filter, Wiki-Link-Auto, Tag-Highlight
- **U9.1–U9.5:** Daily-Notes, Find-in-Files, Recent-Files-Menü, Tag-Panel, Task-Toggle

Siehe `PROJECT_STATUS.md` für Details.

## Lizenz

MIT — siehe [LICENSE](LICENSE).

## Danksagungen

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI-Framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Styling
- [Glamour](https://github.com/charmbracelet/glamour) — Markdown-Rendering
- Inspiriert von [Obsidian](https://obsidian.md/) und dem Python-Original
  [mdskim](https://github.com/dennis605/mdskim) (Textual).
