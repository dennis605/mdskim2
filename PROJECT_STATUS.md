# PROJECT_STATUS — mdskim2

**Intent**: Go + Bubble-Tea markdown-Workspace (modernes Terminal-Tool für strukturierte Markdown-Dokumente).

**Status**: shipped (MVP complete)

**Maturity**: production

**Description**: mdskim2 ist ein TUI-Markdown-Workspace — Workspace-Tree, Live-Preview, Suche, Multi-Tabs, Command Palette, Quick Open. Inspiriert von Obsidian + VS-Code, gebaut als Single-Binary in Go.

**URL**: https://github.com/dennis605/mdskim2

**Start command**: `cd /Users/dennisschonig/projects/mdskim2 && go run ./cmd/mdskim2` (oder mit Workspace-Pfad)

**Last action**: Sprint R10 — Final 7-Gate-Verifikation aller 16 MVP-Items + 8 Obsidian-Stubs L1-L8. Alle Gates grün, alle Tests grün, alle 16 MVP-Items abgedeckt, 12 Architecture-Components gebaut.

**Next step**: Ausliefern oder nächste Phase planen. Optionale post-MVP-Features: Workspace-Grep-UI, Tab-Bar-Visualisierung, Fuzzy-Search für Quick Open, Regex-Search, Sidebar-Resize, fsnotify Auto-Reload.

**Updated**: 2026-09-27

**Confidence**: high

## Sprint-Fortschritt

| Sprint | Inhalt | Status |
|--------|--------|--------|
| R1 | Bootstrap (TUI-Skelett + 3-Pane-Layout) | ✓ done |
| R2 | Workspace + FileTree + Mousemap | ✓ done |
| R3 | Markdown-Datei öffnen + Buffer + Markdown-Syntax-Editor | ✓ done |
| R4 | Save + Standard-Shortcuts + Clipboard | ✓ done |
| R5 | Markdown-Highlighting + Live-TOC | ✓ done |
| R6 | Search (Ctrl+F) + Replace (Ctrl+H) | ✓ done |
| R7 | Preview (HTML/glamour) + Split-View | ✓ done |
| R8 | Tabs + Quick-Open | ✓ done |
| R9 | Command Palette (Ctrl+K) + Workspace-Suche | ✓ done |
| R10 | Polish + 8 Obsidian-Stubs (L1-L8) + Final-Verifikation | ✓ done |

## 7-Gate-Verifikation (R10 Final)

- Gate 1: `go build ./...` ✓
- Gate 2: `go vet ./...` ✓
- Gate 3: `go test ./...` ✓
- Gate 4: `--help` ✓
- Gate 5: Workspace-Boot ✓
- Gate 6: Ctrl+Q Quit ✓
- Gate 7: Snapshot-Diff vs Golden ✓
