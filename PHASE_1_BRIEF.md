# mdskim2 — Moderner Markdown-Workspace für das Terminal

## Mission (verbatim aus User-Spec)

Konzeptionell Obsidian/VS Code/klassischer Windows-Editor.
**Kein Vim-Klon.** Keine Modi. Maus + Standard-Shortcuts (Ctrl+C/V/X/S/Z/Y/A/F/H/P/B/I/N/W/Tab/Home/End). Editor, Preview, Split, Live-TOC H1–H4, Quick-Open Fuzzy, Command Palette, Tabs, Workspace-Search, Sidebar-resize.

Stack-Kandidaten: Go + Bubble Tea/Lip Gloss oder Rust + Ratatui.
**Bevorzugt Go** wegen `bubbles/textarea` (reifer Editor), `goldmark` (Markdown), `chroma` (Highlighting), `lipgloss` (Styling), Single-Binary-Delivery.

## Repo

- Pfad: `~/projects/mdskim2`
- Branch: `main`, leer (git init)
- Sprache: **Go** (per User 2026-09-27, Schritt-3-Entscheidung im PHASE_1_REPORT)
- Existierendes `~/projects/mdskim` (Textual-Python) wird **NICHT** angefasst.

## User-Workflow — du musst der einzige Entscheider sein

Dennis hat Memory-Constraints: "Eine Antwort, kein Menü" und "Fahre fort und treffe eigenmächtig alle Entscheidungen" für die Routinearbeit. **Aber dieser Brief ist ein 0-Setup-Projekt, wo Dennis sehr wohl Mitsprache will** — die Spec ist detailliert, der Stack ist eine Architektur-Entscheidung.

**Hard-Regel für Prime Agent (dieses Sprints):**

- **Triviale Stack-/Modul-Entscheidungen**: selbst entscheiden, kurz im Report begründen.
- **Architektur-Wahl (Go vs Rust), MVP-Scope, ASCII-Wireframe-Layout**: STOP, an User.
- **Triviale Implementierungs-Details (Variablennamen, helper-Funktionen, error-handling)**: entscheiden.
- **Bei Unsicherheit zwischen mehren plausiblen Wegen**: STOP, an User.

## Vorgehensweise — Phase 1 (dieser Sprint, **KEIN Code schreiben**)

Dennis' Spec sagt explizit: "Beginne **nicht sofort mit der vollständigen Implementierung**." Phase 1 ist die 8-Schritt-Analyse aus der Spec.

### Schritt 1 — Anforderungs-Matrix

Erzeuge eine Tabelle: Welche Spec-Anforderung wird von welchem Library-Stack erfüllt. Mindestens:

| Anforderung | Go/Bubble Tea | Rust/Ratatui |
|---|---|---|
| Texteditor (Cursor, Selection, Undo) | bubbles/textarea? | tui-textarea? |
| Markdown Parser | goldmark | pulldown-cmark |
| Syntax Highlighting | chroma | syntect |
| Fuzzy Search | fuzzysearch oder eigene | nucleo oder skim |
| Clipboard | golang.design/x/clipboard | arboard |
| Mouse | Bubble Tea native | ratatui (crossterm) |
| Single Binary | ✓ trivial | ✓ trivial |
| Windows-Terminal | ✓ (via bubbles) | ✓ |
| Unicode/Umlaute | ✓ | ✓ |
| Performance große Files | ??? | ??? |

### Schritt 2 — Bestehende OSS-Komponenten prüfen

Liste konkrete Libraries die wir nutzen werden mit Repo/Stars/Maintenance-Status. Mindestens für:

- TUI-Framework (Bubble Tea, tview)
- Editor-Widget (textarea, edit, tui-rs-extensions)
- Markdown-Rendering (glamour für TUI, goldmark für Backend)
- Highlighting (chroma, syntect, tree-sitter)
- Fuzzy-Search (fzf-Algorithmus)
- Tree-Widget (bubbles/tree, tview tree)
- File-Watching (fsnotify)

### Schritt 3 — Go vs Rust Entscheidung mit Begründung

Vergleiche **konkret für diesen Anwendungsfall** (nicht generisch). Liefer eine Tabelle mit: Editor-Qualität, Clipboard-Reliability (Windows-Terminal-Hölle!), Maus-Bugs, Build-Zeit, Binary-Größe, Iteration-Speed, Maturity der Editor-Komponente.

Dennis' Hinweis: **"Wenn Terminal-Einschränkungen bestimmte Tastenkombinationen verhindern, darf nicht stillschweigend auf Vim-Keybindings ausgewichen werden."** Das ist ein **massiver** Risiko-Punkt — untersuche ob Ctrl+C, Ctrl+V, Ctrl+S im Terminal-Raw-Mode Konflikte haben.

### Schritt 4 — Technische Risiken (PFLICHT)

- **Ctrl+C in Bubble Tea**: das ist SIGINT! Wie fängt man das ab ohne Prozess zu killen? (Lösung: `tea.WithMouseCellMotion`, key handling, etc.)
- **Ctrl+S / Ctrl+Q / Ctrl+Z**: viele Terminals interpretieren das als Flow-Control. XOFF/XON abklemmen? `stty`?
- **Clipboard von TUI aus**: macOS `pbcopy`, Linux `xclip`/`wl-copy`, Windows `clip.exe`. Wrapper mit Fallback auf internes Clipboard.
- **Mouse in SSH / nested terminals**: nur Cell-Motion vs All-Motion, was funktioniert remote?
- **Maus-Selektion**: Standard-Terminal-Selektion geht verloren wenn TUI Fullscreen ist. Wie kriegt man Text in System-Clipboard?
- **Große Files**: 10000+ Zeilen Markdown — Performance von Bubble Tea's Renderer + textarea scroll?
- **Windows-Terminal**: Unicode-Breite, ANSI-Codes, Clipboard-API.

### Schritt 5 — Architektur vorschlagen

Komponenten-Diagramm als ASCII:

```
+------------------+
| App Shell (main) |
+--------+---------+
         |
   +-----+-----+-----+
   |     |     |     |
 Files  Editor  TOC  Preview
   |     |     |     |
 Workspace  Buffer Model  Markdown AST  Rich Render
   |     |     |     |
   +-----+-----+-----+
         |
   +-----+-----+-----+
   |     |     |     |
  Cmd   Search Palette Config
 Palette        
```

Module mit Go-Files (`files.go`, `editor.go`, `toc.go`, `preview.go`, `palette.go`, `cmd.go`, `search.go`, `config.go`).

### Schritt 6 — MVP-Definition

Aus der Spec-Liste der MVP-Items, plus welche Reihenfolge. Phasen:

- **MVP-A**: File-Tree, Markdown-Open, Save, Standard-Shortcuts, Undo/Redo, System-Clipboard, Syntax-Highlighting, Maus-Sidebar-Resize, Status-Bar
- **MVP-B**: Live-TOC H1-H4, TOC-Navigation, Editor-Preview-Split, Preview-Live-Update
- **MVP-C**: Quick-Open Fuzzy, Command Palette, Workspace-Search, Tabs

### Schritt 7 — UI als ASCII-Wireframe

Mindestens 3 Screens:

- **Default Workspace**: Toolbar | Tree | Editor | Preview | Status
- **Quick Open (Ctrl+P)**: Modal über Workspace, Fuzzy-Eingabe, Ergebnisliste
- **Command Palette (Ctrl+Shift+P)**: Modal, Kommandos + Recent

ASCII-Boxen mit klarer Beschriftung welcher Bereich Maus-Klicks bekommt, welcher Cursor hat.

### Schritt 8 — Implementierung (STOP nach 7!)

**Phase 1 endet hier.** Du lieferst einen Report mit:

1. Stack-Entscheidung (Go oder Rust) **mit Begründung**
2. Library-Liste mit URLs und Status
3. Architektur-Skizze
4. Risiko-Liste mit Lösungs-Ansätzen
5. MVP-Scope in Phasen A/B/C
6. 3 ASCII-Wireframes

**STOP nach dem Report. Implementierung kommt in Phase 2, nach User-OK.**

## Akzeptanzkriterien Phase 1

1. **Keine Zeile Anwendungs-Code geschrieben.** Nur Spec-Analyse, evtl. ein `go mod init mdskim2` ist OK aber kein echter Code.
2. Stack-Entscheidung dokumentiert mit konkreten Pro-Contra-Punkten.
3. Mindestens 5 konkrete Library-Empfehlungen mit Repo-URL.
4. Architektur als ASCII-Diagramm.
5. Mindestens 3 ASCII-Wireframes (Default, Quick-Open, Command-Palette).
6. Risiko-Liste mit Lösungs-Ansätzen für Ctrl+C/S/Q, Clipboard, Mouse-SSH.
7. MVP in 3 Phasen A/B/C gegliedert.

## Operatives

- Working dir: `~/projects/mdskim2/`
- Initial-Setup: `git init`, leerer commit auf `main`
- Repo liegt **NICHT** in `~/projects/mdskim` (altes Projekt unangetastet)
- **Keine** Modifikation von `~/.claude/settings.json`, `~/.claude/CLAUDE.md`, `~/projects/mdskim/**`
- **`prime-agent` ist dein einziger Coding-Agent** — keine omp, kein Codex, kein Claude.

## Decision-Gates

- Trivial: Library-URL-Format, Markdown-Bibliothek, Variablenname, Pfad → entscheiden.
- Non-Trivial: **Stack-Wahl Go/Rust, Architektur-Wechsel, MVP-Reihenfolge, ASCII-Layout** → STOP, an User.
- HARD STOP: User fragt "warum Go" und du antwortest "warum nicht Rust" ohne Stack-Entscheidung.

## Communication

- Output: prägnant, mit Markdown-Sections, klarer Stack-Pick.
- 1-Frage-Sektion am Ende: "Was soll ich als nächstes tun?" — kurze 2-3-Optionen-Liste.
- KEINE Code-Snippets (außer in Wireframes).
- KEINE "Ich werde jetzt X implementieren" Versprechen — du **planst**, du implementierst noch nicht.

## Memory-Hinweise

Dennis' Memory:
- "Eine Antwort, kein Menü" — bei Architektur-Entscheidungen ist aber genau das ein Menü legitim.
- "Self-Report ≠ Verification" — am Ende: konkrete Files gelistet, nicht "ich habe es getan".
- Prime-Agent-Pilot-Tests sind nicht ausreichend für Layout-Verifikation — gilt für Phase-2 später.
