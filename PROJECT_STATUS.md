# PROJECT_STATUS — mdskim2

**Intent**: Go + Bubble-Tea markdown-Workspace (modernes Terminal-Tool für strukturierte Markdown-Dokumente).

**Status**: shipped (MVP complete)

**Maturity**: production

**Description**: mdskim2 ist ein TUI-Markdown-Workspace — Workspace-Tree, Live-Preview, Suche, Multi-Tabs, Command Palette, Quick Open. Inspiriert von Obsidian + VS-Code, gebaut als Single-Binary in Go.

**URL**: https://github.com/dennis605/mdskim2

**Start command**: `cd /Users/dennisschonig/projects/mdskim2 && go run ./cmd/mdskim2` (oder mit Workspace-Pfad)

**Last action**: Sprint U6 — UI-Revamp (Python-mdskim-Look). Keybinding-Toolbar oben, TabbedContent rechts (Preview | TOC | Backlinks) statt immer-TOC. Neues `internal/backlinks` Package mit 5 Tests. 4 U6-Tests grün.

**Next step**: Optionale UI-Polish-Sprints U7+: Sidebar-Resize via Ctrl+Left/Right, File-Watcher (fsnotify), Markdown-Lint-Markers, Splitter-Drag. Aktueller Look entspricht jetzt dem Python-mdskim.

**Updated**: 2026-09-27 23:36 UTC+02:00

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

### UX-Polish Sprints (post-MVP)

**Sprint U1 — UX Polish (Auto-Open + Visible Cursor + Edit-Badge)** ✓  
Commit `3a04381` — Tree-Pfeile öffnen automatisch Datei, sichtbarer ▶-Marker an Cursor-Zeile, lila EDIT-Header.

**Sprint U2 — Focus-Modell + Editor-Look (Line-Numbers + Block-Cursor)** ✓  
Commit `6ce50fb` — `m.focus` "tree"/"editor" mit Routing der Pfeiltasten. Escape/Enter steuern Focus. Tree-Pfeile bleiben persistent nach auto-open. Editor zeigt Zeilennummern, Focus-Badge `[EDIT]`/`[FILES]`, aktive Zeile gelb.

**Sprint U3 — Advanced Editor Features (Selection · Word-Nav · Auto-Indent · Line-Ops · Go-to-Line · Wrap · Scroll)** ✓  
Commit `d4c6c73` — 10 echte Editor-Features:
- Block-Cursor an exakter Cursor-Spalte (`\x1b[7m` reverse-video char overwrite)
- Shift+↑↓←→ Selection mit lila Background-Highlight (48;5;57)
- Ctrl+← / Ctrl+→ wortweise Navigation (VS Code-style)
- Alt+↑ / Alt+↓ Zeile verschieben
- Ctrl+D Zeile duplizieren
- Ctrl+W Soft-Wrap Toggle (Header zeigt [WRAP])
- Ctrl+G Go-to-Line Modal (1-indexed, clamp 1..N)
- Auto-Indent: Enter übernimmt führende Whitespace der aktuellen Zeile
- Status-Bar mit Words N · Chars N · Scroll-Indicator (↑↓ + Position)
- Editor-Viewport mit clampScrollOffset (Cursor bleibt im Viewport)

## UX-Sprint-Status (current = U3 done)

| Sprint | Status | Commit |
|---|---|---|
| U1 | ✓ done | `3a04381` |
| U2 | ✓ done | `6ce50fb` |
| U3 | ✓ done | `d4c6c73` |
| U4+ | offen — bereit für nächste User-Anforderung |

## Letzte Verifikation
- 7-Gate alle ✓ (Build · Vet · Test · Help · Boot · Ctrl+Q Quit · Snapshot vs Golden)
- 60+ Tests grün (`go test ./...`)
- Modul: `github.com/dennis605/mdskim2` v0.3.0 (post-MVP UX-Polish)
