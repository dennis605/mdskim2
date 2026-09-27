# mdskim2 — Project Status

## Aktuell

- **Intent**: Phase 2 / Sprint R1 (Bootstrap) liefert einen lauffähigen TUI-Skelett
  mit 3-Pane-Layout, Shortcut-Footer, Status-Bar und Demo-Vault.
- **Last action**: Sprint R1 abgeschlossen — 7/7 Gates grün, 4 Tests passend,
  Golden-Snapshot erzeugt. Working tree bereit zum Commit.
- **Next step**: Hermes-Self-Review → User-Commit → Sprint R2 (Workspace + FileTree).
- **Status**: active
- **Maturity**: prototype (R1)
- **Updated**: 2026-09-27 22:24 UTC+02:00
- **Confidence**: high
- **URL**: https://github.com/dennis605/mdskim2 (Repo angelegt, public)
- **Description**: Modern Markdown Workspace for the Terminal — Obsidian/VS-Code-feeling,
  Windows-artige Tastenkürzel, single static binary (Go+Bubble Tea).
- **Start command**: `cd ~/projects/mdskim2 && go run ./cmd/mdskim2 demo/`
  - Mit eigenem Vault: `cd ~/projects/mdskim2 && go run ./cmd/mdskim2 ~/mein-vault/`
  - Tests: `cd ~/projects/mdskim2 && go test ./...`
  - Snapshot-Update: `cd ~/projects/mdskim2 && go test ./test/smoke/... -update`

## Architektur (Stand R1)

```
mdskim2/
├── cmd/mdskim2/main.go          # CLI entry: flag, tea.NewProgram, Quit
├── internal/
│   ├── app/                     # Bubble-Tea Model + Update + View
│   │   ├── app.go               # Model, Init, Update, View, handleKey
│   │   └── keymap.go            # Spec-Shortcut-Map (16 Keys, 1 in R1)
│   └── ui/                      # Themes, Layout, Status, Footer
│       ├── theme.go             # Light + Dark Themes (lipgloss.Style)
│       ├── layout.go            # 3-Pane-Layout (Sidebar, Editor, TOC)
│       ├── status.go            # Status-Bar mit Workspace + Mode + Lines
│       └── footer.go            # Shortcut-Hint-Leiste
├── demo/Test Engineering/       # Demo-Vault (6 .md Files + README)
│   ├── Teststrategie.md
│   ├── Testplanung.md
│   ├── Testarten.md
│   ├── Testmethoden.md
│   ├── Testdesign.md
│   └── Projekte/Teststrategie-SAP.md
├── test/smoke/                  # Headless-Tests (golden snapshot)
│   ├── boot_test.go             # View-Content + Ctrl+Q
│   ├── boot_smoke_test.go       # Required-Markers Test
│   ├── snapshot_test.go         # Snapshot-Diff vs Golden
│   └── testdata/
│       ├── snapshot-boot.txt    # Golden-Snapshot
│       └── sample-workspace/    # Test-Fixture-Vault
├── go.mod                       # github.com/dennis605/mdskim2 (Go 1.26.2)
├── go.sum
├── PHASE_1_BRIEF.md             # 2-Zeilen-Mission
└── PHASE_1_REPORT.md            # Vollständige Spez-Analyse (47 KB)
```

## MVP-Sprint-Plan (Phase 2)

| Sprint | MVP-Item | Status |
|---|---|---|
| **R1** | **Bootstrap (3-Pane, Footer, Status, Demo-Vault, Ctrl+Q)** | **done** |
| R2 | Workspace/Ordner öffnen, Dateibaum (Maus+Kbd) | open |
| R3 | Markdown-Datei öffnen + bearbeiten | open |
| R4 | Speichern, Standard-Shortcuts, Clipboard | open |
| R5 | Live-TOC (H1-H4), Markdown-Highlighting | open |
| R6 | Suche (Ctrl+F), Sidebar-Resize, Replace (Ctrl+H) | open |
| R7 | Editor/Preview/Split-View (Ctrl+Backslash) | open |
| R8 | Quick Open (Ctrl+P), Tabs (Ctrl+J/K/W) | open |
| R9 | Command Palette (Ctrl+Shift+P), Workspace-Suche | open |
| R10 | Polish, 8 Obsidian-Architektur-Stubs (L1-L8) | open |

## Sprint R1 Verifikation (7/7 Gates grün)

| Gate | Status |
|---|---|
| 1. `go build ./...` | ✓ |
| 2. `go vet ./...` | ✓ |
| 3. `go test ./...` (4 Tests) | ✓ |
| 4. `--help` zeigt Usage | ✓ |
| 5. App-Boot-Smoke | ✓ |
| 6. Ctrl+Q Quit-Test | ✓ |
| 7. Snapshot-Diff vs Golden | ✓ |

**Implementiert R1 (1 von 16):**
- Ctrl+Q — Quit

**Vorbereitet (in R2-R9 zu implementieren):**
- Ctrl+S/O/N/C/X/V/Z/Y/A/B/I/F/H/P/Shift+P/W — in `internal/app/keymap.go`.

## Sprint-R1-Learnings

1. **Bubble Tea Ctrl+Q ohne Program-Run testen**: `m.Update(tea.KeyMsg{Type: tea.KeyCtrlQ})`
   returnt `(tea.Model, tea.Cmd)` — cmd ist `tea.Quit`.
2. **Snapshot-Tests ohne TTY**: `m.View()` direkt aufrufen, kein `tea.NewProgram`.
3. **Layout in `internal/ui/`** statt im Model: `lipgloss.NewStyle().Width().Height().Render(content)`.
4. **`go mod tidy` removed unused deps**: Erst bei tatsächlichem Import in R3-R9 kommen sie zurück.
5. **Python-escaped Strings → Go-Files**: `r"..."` in Python oder `\n` schreiben, sonst
   werden `
`-Sequenzen zu echten Newlines in den Go-Files (Compile-Error).
6. **lipgloss-Version für glamour**: `v1.1.1-0.20250404203927-76690c660834` ist mit
   `glamour@latest` kompatibel (Pseudo-Pre-Release).

## Offene Punkte für Sprint R2

- FileTree-Komponente: `internal/workspace/filetree.go` mit recursive ls.
- Workspace-Open: `os.ReadDir`-basiert mit `fsnotify` für Live-Reload.
- Mousemap: Sidebar-Klick auf File → öffnet im Editor.
- Tabs (R8) und Modal-Layer (R6/R8/R9).
- Pflicht-Stubs (R10): 8 Obsidian-Interfaces vorbereiten.
