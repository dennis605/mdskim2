# Sprint R2 — Workspace + FileTree + Mousemap

## Ziel

MVP-Item: "Workspace/Ordner öffnen + Dateibaum mit Maus+Tastatur"

## Spec-Konformität

User-Spec v1, Items:
- Dateibaum links (Sidebar)
- Workspace-Pfad oben links im Header
- Tastatur: ↑↓ navigieren, Enter öffnet, Backspace geht hoch
- Maus: Klick auf File öffnet ihn, Klick auf Verzeichnis expandet/kollabiert
- Sidebar-Label "FILES" oben
- Aktive Datei visuell hervorgehoben

## Implementation

### Workspace-Layer (NEU)
- `internal/workspace/workspace.go` — Workspace-Loader (Pfad → Tree)
- `internal/workspace/filetree.go` — Tree-Datenstruktur + Render
- `internal/workspace/watcher.go` — fsnotify-basierter Live-Reload

### Mousemap (NEU)
- Sidebar-Klick (X in Sidebar-Bereich) → tree.Select(row)
- Editor-Klick → Cursor im Editor (kommt in R3)

### Keymap (Updates in app.go)
- `↑/↓` — Navigation in Sidebar
- `Enter` — Datei öffnen
- `Backspace` — Parent-Verzeichnis
- `Ctrl+O` — Workspace/Datei-Dialog (R4-vorbereitung)
- `F2` — Rename (R4)
- `Del` — Delete (R4)

### R2-Verifikation

- Workspace öffnen via CLI: `go run ./cmd/mdskim2 demo/`
- FileTree zeigt 6 .md-Files + 1 README + 1 Subfolder
- Mousemap: Klick in Sidebar selektiert File (visualisiert in TreeActiveFile)
- ↑↓ navigiert durch Tree
- Enter öffnet ausgewählten File im Editor (R3 füllt Buffer)
- Backspace klappt Parent auf

## 7-Gates

1. `go build` ✓
2. `go vet` ✓
3. `go test ./...` ✓ (mind. 6 neue Tests: Tree-Build, Render, Mouse-Click, Navigation)
4. `./mdskim2 demo/` — Tree sichtbar
5. `--help` ✓
6. Ctrl+Q Quit ✓
7. Snapshot-Diff vs neues Golden

## Out-of-Scope für R2 (spätere Sprints)

- File-Editor mit Buffer (R3)
- Markdown-Highlighting (R5)
- Tabs (R8)
- Rename/Delete (R4)
