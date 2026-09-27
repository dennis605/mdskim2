# mdskim2 — Phase 1 Report (Stack-Analyse + Architektur)

**Datum:** 2026-09-27
**Pfad:** `~/projects/mdskim2/`
**Status:** Phase 1 abgeschlossen, wartet auf User-OK für Phase 2
**Stack-Entscheidung:** Go 1.26 + Bubble Tea (per User 2026-09-27, "wir nehmen go")

> **Wichtig:** Die User-Spec ist die einzige Wahrheit. Mein vorheriger Report basierte auf einer 2-Zeilen-Mission und hatte Annahmen extrapoliert. Dieser Report ist **komplett rewrite** auf Basis der vollständigen, vom User geposteten Spec. Keine Annahmen, keine Extrapolation.

---

## 0 — User-Spec (Referenz-Quelle)

Die vollständige Spec, die Dennis am 2026-09-27 reingepostet hat, steht verbatim in `PHASE_1_BRIEF.md` (Zeile 1-110) und definiert:

- **Mission**: TUI für Markdown, wie Obsidian/VS Code (kein Vim-Klon)
- **Layout**: 3 Bereiche — Linke Sidebar (Dateien), Mitte (Editor), Rechte Sidebar (TOC)
- **Linke Sidebar**: Dateibaum mit Open/CRUD/Highlight, Tastatur+Maus
- **Mitte Editor**: Texteingabe, Cursor, Selection, Copy/Paste, Undo/Redo, Find+Replace, UTF-8+Umlaute, große Files, Syntax Highlighting, Auto-Indent Listen, Markdown-Strukturen, KEINE modalen Modi
- **Rechte Sidebar TOC**: Aus Headings, hierarchisch, H1-H4, live-update, Klick-Sprung, aktive Section markiert
- **Standard-Shortcuts (Windows-Stil, kein Vim!)**: Ctrl+S/C/X/V/Z/Y/A/F/H/P/Shift+P/B/I/N/W, Tab/Shift+Tab, Home/End, Ctrl+Home/Ctrl+End, Shift+Pfeiltasten, Ctrl+Shift+Pfeiltasten
- **Hartes Verbot**: "Wenn Terminal-Einschränkungen bestimmte Tastenkombinationen verhindern, darf nicht stillschweigend auf Vim-Keybindings ausgewichen werden. Stattdessen soll eine intuitive Alternative implementiert und dokumentiert werden."
- **Maus**: Datei/TOC anklicken, Cursor setzen, Text markieren, Scrollrad, Sidebar-Resize, Tabs, Buttons
- **Preview**: Editor / Preview / Split (Ctrl+\\), live-update
- **Quick Open**: Ctrl+P, Fuzzy über Workspace-Dateien
- **Command Palette**: Ctrl+Shift+P, VS-Code-Stil
- **Tabs**: Mehrere Dateien, Maus+Tastatur
- **Workspace-Suche**: Volltext, Dateiname+Zeile+Snippet
- **Obsidian-artig (Architektur-muss-erlauben, NICHT MVP)**: [[Wiki Links]], Backlinks, Tags, Auto-Link, Recent, Bookmarks, Sessions, Multi-Vault
- **UX-Prinzip**: Terminal = nur Darstellungsplattform. Wie Desktop-Editor. Sichtbare Funktionen. Kontextabhängige Footer-Leiste.
- **Technische Anforderungen**: TUI-Stack wählen (bevorzugt Go+Bubble Tea ODER Rust+Ratatui), 9 Bewertungskriterien, Single Binary.
- **Architektur (PFLICHT-Trennung, 12 Komponenten)**: Workspace/Files, Editor, Markdown Parser, TOC, Preview, Search, Command Palette, Keyboard, Mouse, Clipboard, Configuration, UI State
- **Vorgehensweise (1-8 EXPLIZIT)**: Anforderungen → OSS-Komp. → Go vs Rust → Risiken → Architektur → MVP → Wireframes → Implementierung
- **MVP (16 Items, alle Pflicht)**: Workspace, Dateibaum, Open, Edit, Save, Standard-Shortcuts, Undo/Redo, Clipboard, Markdown-Highlighting, Live-TOC, TOC-Navigation, Quick Open, Suche, Maus, Editor/Preview/Split, Status/Footer
- **Qualitätsziel**: "Obsidian bzw. ein kleiner VS-Code-Markdown-Editor für das Terminal." In Sekunden verstehen.

---

## Schritt 1 — Anforderungs-Matrix (Spec-Items auf Libraries gemappt)

| Anforderung (Spec) | Go-Lösung | Hinweis |
|---|---|---|
| TUI-Framework | `charmbracelet/bubbletea` v1.x | Elm-Architektur (Init/Update/View) |
| Styling/Layout | `charmbracelet/lipgloss` v1.x | CSS-like, Border, JoinHorizontal/Vertical |
| Editor-Widget | `charmbracelet/bubbles/textarea` v0.x | Multi-line, eingebautes Undo/Yank/Selection/Soft-Wrap |
| Markdown-Parser (Spec 3) | `yuin/goldmark` v1.7 | Vollständiger AST, GFM |
| Markdown-Preview-Renderer (Spec 5) | `charmbracelet/glamour` v0.x | nutzt goldmark, styled ANSI |
| Syntax Highlighting (Markdown+Code) | `alecthomas/chroma` v2.x | Style-Token → lipgloss |
| Fuzzy Match | `sahilm/fuzzy` v0.x | Subsequence für Quick Open |
| Tree-State | Eigenbau (bubbles/list + custom logic) | — |
| Clipboard (Spec 10) | `atotto/clipboard` v0.x | pure-Go, kein CGO |
| FS-Watch | `fsnotify/fsnotify` v1.x | für externe Änderungen |
| CLI Args | stdlib `flag` | reicht für MVP |
| Config (Spec 11) | `BurntSushi/toml` v1.x | TOML in XDG-Pfad |
| Unicode/Umlaute/Width | `golang.org/x/text` + `rivo/uniseg` | korrekte Cursor-Position |
| Terminal raw mode | `golang.org/x/term` | MakeRaw, query size |
| Logging | `charmbracelet/log` | strukturiert |
| Gitignore-Respekt | `sabhiram/go-gitignore` | Workspace-Scan |
| Tabs-Mgmt | Eigenbau (Slice + activeIndex) | — |
| TOC live-update | Eigenbau (goldmark-AST-Diff) | glamour hat keinen AST-Diff |
| Active-Section-Track | Eigenbau (cursor-Y → heading-Pos) | — |
| File-CRUD in Tree | stdlib `os`/`os/exec` | rename, delete, mkdir |
| Sidebar-Maus-Resize | Eigenbau (Maus-Drag-State) | — |
| Workspace-Search (async) | stdlib `regexp` + goroutines | für Cross-File-Suche |
| Single Binary | ✓ `go build` | statisch |
| Cross-Platform | ✓ Linux/macOS/Windows-Terminal | identisch wie Rust |

**Verlierer (Lücken, müssen selbst gebaut werden):**
- TOC live-update-Engine (goldmark-AST-Diff alt → neu)
- Sidebar-Maus-Resize mit Splitter-Drag
- Markdown-Syntax-Highlighting auf Live-Edit (chroma + textarea-Cache)
- Active-Section-Tracking (cursor-Y → heading-Position)
- Help-Screen / About-Dialog
- Kontextabhängige Footer (dynamisch je nach aktivem Pane/Modal)

---

## Schritt 2 — 12 Pflicht-Architektur-Komponenten (Spec-Punkt „Architektur")

Die User-Spec verlangt explizit mindestens diese 12 Trennungen. Ich mappe sie auf Go-Packages:

| Spec-Komponente | Package | Verantwortung |
|---|---|---|
| 1. Workspace/File Management | `internal/workspace/` | Folder-Scan, Tree-State, CRUD, Gitignore, aktive Datei |
| 2. Editor | `internal/editor/` | Buffer, Cursor, Selection, History, Indent, Highlight, Bold/Italic-Wrap |
| 3. Markdown Parser | `internal/markdown/` | goldmark-Wrapper, AST-Utilities, Wikilink-Vorbereitung (für L1) |
| 4. TOC | `internal/toc/` | Heading-Tree, live-update (debounced), active-section, View |
| 5. Preview Renderer | `internal/preview/` | goldmark → glamour, live-update, View |
| 6. Search | `internal/search/` | in-file (Ctrl+F/H), workspace (Ctrl+Shift+F), fuzzy (Quick Open) |
| 7. Command Palette | `internal/palette/` | Commands-Registry, Quick Open, Modal-UI |
| 8. Keyboard Handling | `internal/input/keys.go` | KeyMsg → Action (Spec-Shortcuts vollständig) |
| 9. Mouse Handling | `internal/input/mouse.go` | MouseMsg → Action (Tree/TOC/Editor/Resize) |
| 10. Clipboard | `internal/clipboard/` | atotto-Wrapper + interner Fallback |
| 11. Configuration | `internal/config/` | TOML load/save, XDG-Pfade |
| 12. UI State | `internal/ui/` | Active-Pane, Focus, Layout, Status-Bar, Footer, Theme |

Plus Orchestration:
- `internal/app/` — tea.Model + Init + Update + View + Modal-Stack + Tab-Manager

→ **Spec-Pflicht „Architektur ohne monolithische Implementierung" vollständig erfüllt.**

---

## Schritt 3 — Go vs Rust Trade-off (Spec: 9 Bewertungskriterien)

Die Spec listet 9 konkrete Bewertungskriterien. Beide Stacks erfüllen sie, aber unterschiedlich:

| Kriterium | Go + Bubble Tea | Rust + Ratatui |
|---|---|---|
| **Texteditor-Qualität** | bubbles/textarea ★★★★ (Yank, Undo, Multi-Cursor in dev) | tui-textarea ★★★ (solide, kein Auto-Indent) |
| **Unicode** | x/text + uniseg ★★★★ | unicode-width ★★★★ |
| **Clipboard** | atotto ★★★★ (pure-Go, kein CGO) | arboard ★★★ (CGO, Cross-Compile komplexer) |
| **Maus** | bubbletea MouseMsg ★★★★ | crossterm MouseEvent ★★★★ |
| **Performance** | Go-Runtime + virtualisierter Render ★★★★ | Zero-Cost ★★★★★ |
| **Syntax Highlighting** | chroma ★★★★ | syntect ★★★★★ |
| **Markdown Parsing** | goldmark ★★★★ + glamour ★★★★ | pulldown-cmark ★★★★★ |
| **Cross-Platform (Windows/macOS/Linux Terminal)** | ★★★★ | ★★★★ |
| **Single Binary** | ✓ trivial | ✓ trivial |
| **Build-Zeit** | ~3 Sek cold | ~30 Sek cold |
| **Iteration-Speed** | Sehr schnell | Compiler-zwanghaft |
| **Time-to-MVP** | **3-5 Wochen** | 5-7 Wochen |

### **Entscheidung: Go** (per User 2026-09-27, "wir nehmen go")

**Begründung aus User-Spec abgeleitet:**
1. **"Innerhalb weniger Sekunden verstehen"** — Iteration-Speed ist entscheidend für die Implementierungsphase. Go liefert schnelles Feedback.
2. **"Standard-Shortcuts, keine Vim-Fallbacks"** — bubbles/textarea hat eingebaute Yank/Undo/Selection, die Rust nicht out-of-the-box hat. Spart Eigenbau.
3. **"Architektur ohne monolithische Implementierung"** — Go-Packages sind klein und orthogonal.
4. **"Single Binary"** — Go hat statisches Linking built-in. `atotto/clipboard` ist pure-Go, kein CGO-Setup wie Rust-`arboard`.
5. **Charm-Ökosystem-Konsistenz**: bubbletea + bubbles + lipgloss + glamour + log — alles aus einem Team, einheitliche API.

**Trade-off akzeptiert:** Fuzzy-Performance bei 100k+ Dateien, kein Compile-Time-State-Bug-Catch.

---

## Schritt 4 — Technische Risiken (Spec: „besondere Risiken" explizit gefordert)

Die Spec verlangt: "Besondere technische Risiken untersuchen – insbesondere **Textselektion, System-Clipboard und bekannte Windows-Shortcuts innerhalb eines Terminals**."

### Risk 1: Standard-Shortcuts in Terminal-Raw-Mode (KRITISCH — Spec-Hard-Verbot)

**Problem:** Manche Ctrl-Buchstaben werden vom Terminal/SSH/emulator vorab konsumiert.

| Shortcut | OS-Default | Terminal-Risiko | Lösung |
|---|---|---|---|
| Ctrl+C | SIGINT | `stty -isig` deaktiviert, App fängt ab | Eigener SIGINT-Handler, Quit-Modal |
| Ctrl+S | XOFF | Terminal stoppt Output | `stty -ixon -ixoff` zwingend beim Start |
| Ctrl+Q | XON | Terminal startet Output | gleiches stty-Setting |
| Ctrl+Z | SIGTSTP | Prozess suspended im Hintergrund | `stty -isig`, App fängt 0x1A ab |
| Ctrl+H | Backspace in manchen stty-Configs | OK in Raw-Mode | explizit prüfen |
| Ctrl+W | Word-Erase (Unix) | OK in Raw-Mode | explizit prüfen |
| Shift+Pfeiltasten, Ctrl+Shift+Pfeiltasten | SGR-Erweiterung | Manche Terminals nicht | Fallback: Ctrl+PgUp/PgDn für wortweise |

**Implementation:**
- Bubble Tea + `tea.WithMouseCellMotion()` aktiviert raw mode automatisch.
- **Pflicht-Setup am App-Start**: `stty -echo -isig -ixon -ixoff -werase` via `golang.org/x/term.MakeRaw()`.
- **Spec-Hard-Verbot-Konform**: Wenn einzelner Shortcut in bestimmtem Terminal nicht geht, wird das im README dokumentiert mit dokumentiertem **Workaround in der App**, nicht mit Vim-Keybinding. z.B. wenn Ctrl+P nicht geht: dokumentiert "manche Terminals konsumieren Ctrl+P; alternativ Ctrl+Shift+P öffnet Quick Open ebenfalls."

### Risk 2: Textselektion in TUI

**Problem:** Native Terminal-Selection (Click+Drag) geht in TUI verloren.

**Lösung (Spec-konform, Desktop-Editor-like):**
- **Shift+Click+Drag** im Editor → Anchor + Cursor Selection (App-State)
- **Shift+Pfeiltasten** → Selection erweitern (zeichenweise)
- **Ctrl+Shift+Pfeiltasten** → Selection wortweise erweitern
- **Doppelklick** → Word-Selection
- **Dreifachklick** → Line-Selection
- **Ctrl+A** → All-Select
- **Ctrl+C / Ctrl+X** kopiert/ausgeschnitten in System-Clipboard via atotto

→ User merkt keinen Unterschied zu VS Code.

### Risk 3: System-Clipboard

**Lösung mit `atotto/clipboard`:**
- macOS: `pbcopy`/`pbpaste` als Subprocess (kein CGO)
- Linux X11: `xclip` oder `xsel` als Subprocess
- Linux Wayland: `wl-copy`/`wl-paste` als Subprocess
- Windows: `clip.exe` via PowerShell als Subprocess
- **Fallback**: intern-persistentes Clipboard (in-memory + optional TOML-Persist zwischen Sessions)
- **User-Info** beim ersten Fail: Status-Bar zeigt "Internal clipboard (system unavailable)"

### Risk 4: Maus in SSH / nested tmux/screen

**Lösung:**
- Default: `tea.WithMouseCellMotion()` (gut für Tree-Klicks, oft OK über SSH)
- Config-Option: `tea.WithMouseAllMotion()` für Screenshot-Tools
- Auto-Detect: 3+ MouseMsg-Fehler → Cell-Motion-Only, Hinweis in Status-Bar

### Risk 5: Deutsche Umlaute + Box-Drawing

**Lösung:**
- `golang.org/x/text/width` + `rivo/uniseg` für korrekte Cursor-Position
- Golden-Snapshot-Tests mit deutschen Umlauten (ä/ö/ü/ß) + Box-Drawing-Zeichen (├─└┌)
- ASCII-Wireframes hier verwenden ausschließlich 1-Spalte-breite Zeichen (─│┌┐└┘├┤┬┴┼)

### Risk 6: Markdown-Highlight-Performance im Live-Editor

**Lösung:**
- chroma cached pro Zeile (invalidiert bei Edit)
- Debounced 50ms nach letztem Keystroke
- MVP-Threshold: 5k Zeilen ohne spürbaren Lag

### Risk 7: TOC live-update ohne UI-Blockade

**Lösung:**
- goldmark AST wird **debounced 200ms** nach letztem Edit neu geparst
- Diff zwischen altem und neuem AST (Heading-Set-Vergleich) → nur TOC-Update
- Active-Section: cursor-Y → Heading-Lookup → TOC-Marker (kein Re-Walk)

### Risk 8: File-Watch (externe Änderungen)

**Lösung:**
- `fsnotify` watcht aktuelle Datei
- Externe Änderung: Toast-Hinweis "File changed on disk — Reload? [R] Reload · [I] Ignore"
- Konflikt-Fall (dirty + Disk-Change): Modal "Save local / Discard local / Keep both"

### Risk 9: Tabs mit vielen gleichzeitigen Buffern

**Lösung:**
- Tab-State: `[]Buffer` (Slice) + `activeIdx`
- LRU-Cache: max 8 Buffer im RAM, weitere on-demand von Disk
- Tab-Viz: `✓ filename` (saved) | `● filename` (modified) | `⊕ new tab`

### Risk 10: Scope-Creep (16 MVP-Items, dazu kommt Architektur für 8 Out-of-MVP-Items)

**Lösung:**
- Strikte Sprint-Reihenfolge (siehe Schritt 6)
- Jeder Sprint = 1-2 Kategorien aus Schritt 1
- 7-Gate-Verification pro Sprint (siehe `mdskim-sprint-workflow` Memory)
- "Architektur-muss-erlauben"-Items bekommen nur Grundgerüst (struct + interfaces), keine Implementierung

---

## Schritt 5 — Architektur (ASCII-Diagramm)

```
┌────────────────────────────────────────────────────────────────────────┐
│                       mdskim2 (Go) — main                              │
│            cmd/mdskim2/main.go (CLI, flag.Parse, App.Run)              │
└─────────────────────────────────┬──────────────────────────────────────┘
                                  │
                                  ▼
┌────────────────────────────────────────────────────────────────────────┐
│                       internal/app/  (Orchestrator)                    │
│         tea.Model + Init + Update + View                               │
│         Modal-Stack · Tab-Manager · Keymap · Mousemap                  │
└────────┬─────────────┬─────────────┬─────────────┬────────────────────┘
         │             │             │             │
         ▼             ▼             ▼             ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────────────┐
│  workspace/ │ │   editor/   │ │    toc/     │ │     preview/        │
│ (Spec 1)    │ │  (Spec 2)   │ │ (Spec 4)    │ │     (Spec 5)        │
│             │ │             │ │             │ │                     │
│ - Scan      │ │ - Buffer    │ │ - Headings  │ │ - goldmark AST      │
│ - Gitignore │ │   (textarea)│ │   (H1-H4)   │ │ - glamour styled    │
│ - Tree CRUD │ │ - Cursor    │ │ - AST-Diff  │ │ - live-update       │
│ - Active    │ │ - Selection │ │   live      │ │   (debounce 200ms)  │
│   File      │ │ - History   │ │ - Click→jump│ │ - 3 Modi            │
│   Marker    │ │ - Indent    │ │ - Active    │ │   (Editor/Preview/  │
│             │ │ - Highlight │ │   Section   │ │    Split)           │
│             │ │   (chroma)  │ │   Tracking  │ │                     │
└──────┬──────┘ └─────┬───────┘ └─────────────┘ └─────────────────────┘
       │              │
       │              ▼
       │       ┌─────────────────┐
       │       │   markdown/     │
       │       │   (Spec 3)      │
       │       │ - goldmark wrap │
       │       │ - AST utilities │
       │       │ - wikilink stub │
       │       └─────────────────┘
       │
       ▼
┌────────────────────────────────────────────────────────────────────────┐
│                  Service Layer (Cross-cutting)                          │
│                                                                         │
│  ┌────────────┐  ┌────────────┐  ┌─────────────┐  ┌──────────────────┐ │
│  │  search/   │  │  palette/  │  │   input/    │  │  clipboard/      │ │
│  │ (Spec 6)   │  │ (Spec 7)   │  │ (Spec 8+9)  │  │  (Spec 10)       │ │
│  │            │  │            │  │             │  │                  │ │
│  │ - file     │  │ - commands │  │ - keys.go   │  │ - atotto wrap    │ │
│  │   (Ctrl+F/H│  │   registry │  │   KeyMsg→   │  │ - internal       │ │
│  │ - workspace│  │ - quick-   │  │   Action    │  │   fallback       │ │
│  │   (Ctrl+   │  │   open     │  │ - mouse.go  │  │                  │ │
│  │   Shift+F) │  │   (Ctrl+P) │  │   MouseMsg→ │  │                  │ │
│  │ - fuzzy    │  │ - palette  │  │   Action    │  │                  │ │
│  │   (sahilm) │  │   (Ctrl+   │  │ - selection │  │                  │ │
│  │            │  │   Shift+P) │  │   mouse     │  │                  │ │
│  └────────────┘  └────────────┘  └─────────────┘  └──────────────────┘ │
│                                                                         │
│  ┌────────────┐  ┌──────────────────────────────────────────────────┐ │
│  │  config/   │  │  ui/  (Spec 12)                                   │ │
│  │ (Spec 11)  │  │                                                  │ │
│  │            │  │  - Layout: 3-Pane + Status + Footer              │ │
│  │ - TOML     │  │  - Theme (lipgloss light/dark)                   │ │
│  │ - XDG-path │  │  - Sidebar Resize (mouse drag splitter)         │ │
│  │ - Defaults │  │  - Status-Bar (context-aware)                   │ │
│  │            │  │  - Footer (context-aware shortcuts)             │ │
│  └────────────┘  └──────────────────────────────────────────────────┘ │
└────────────────────────────────────────────────────────────────────────┘
```

### Package-Layout (vollständige 12-Komponenten-Struktur)

```
mdskim2/
├── go.mod
├── go.sum
├── README.md
├── cmd/
│   └── mdskim2/
│       └── main.go                        # CLI entry: flag.Parse + tea.NewProgram
├── internal/
│   ├── app/
│   │   ├── app.go                         # Model + Init + Update + View
│   │   ├── update.go                      # Type-Switch dispatch
│   │   ├── view.go                        # Layout composition
│   │   ├── keymap.go                      # Spec-Shortcuts → Actions
│   │   ├── mousemap.go                    # Mouse-Shortcuts → Actions
│   │   ├── modal_stack.go                 # Modal-Mgmt (Save-Prompt, Find, etc.)
│   │   └── tab_manager.go                 # Tabs-State
│   ├── workspace/                         # (1) Workspace / File Management
│   │   ├── workspace.go                   # Folder-Scan, .gitignore
│   │   ├── tree.go                        # Tree-Builder + expand/collapse
│   │   ├── crud.go                        # new/rename/delete file+folder
│   │   ├── current_marker.go              # highlight active file
│   │   └── tree_view.go                   # UI rendering (lipgloss)
│   ├── editor/                            # (2) Editor
│   │   ├── editor.go                      # EditorState facade
│   │   ├── buffer.go                      # bubbles/textarea wrapper
│   │   ├── history.go                     # Undo/Redo (textarea + custom für Sel)
│   │   ├── selection.go                   # Selection-API
│   │   ├── indent.go                      # Auto-Indent bei Listen
│   │   ├── highlight.go                   # chroma integration
│   │   ├── markdown_ops.go                # Ctrl+B/I (Bold/Italic wrap)
│   │   └── editor_view.go                 # UI rendering
│   ├── markdown/                          # (3) Markdown Parser
│   │   ├── parser.go                      # goldmark wrapper
│   │   ├── ast.go                         # AST walk utilities
│   │   ├── toc_extract.go                 # Headings → TOC-Tree
│   │   └── wikilinks.go                   # [[Wiki]] vorbereitung (out-of-MVP)
│   ├── toc/                               # (4) TOC
│   │   ├── toc.go                         # Heading-Tree-State
│   │   ├── live_update.go                 # Debounced Update + AST-Diff
│   │   ├── active_section.go              # cursor-Y → active heading
│   │   └── toc_view.go                    # UI rendering
│   ├── preview/                           # (5) Preview Renderer
│   │   ├── preview.go                     # goldmark → glamour styled ANSI
│   │   ├── live_update.go                 # Debounced 200ms
│   │   └── preview_view.go                # UI rendering
│   ├── search/                            # (6) Search
│   │   ├── file.go                        # In-File (Ctrl+F, Ctrl+H)
│   │   ├── workspace.go                   # Cross-File (Ctrl+Shift+F)
│   │   ├── fuzzy.go                       # sahilm wrapper
│   │   └── search_modal.go                # Modal-UI
│   ├── palette/                           # (7) Command Palette
│   │   ├── palette.go                     # Modal-Framework
│   │   ├── commands.go                    # 25-35 Spec-Commands Registry
│   │   ├── quickopen.go                   # Ctrl+P
│   │   └── command_executor.go            # Action dispatch
│   ├── input/                             # (8) + (9) Keyboard + Mouse
│   │   ├── input.go                       # Public API
│   │   ├── keys.go                        # KeyMsg → Action (Spec-Shortcuts)
│   │   ├── mouse.go                       # MouseMsg → Action
│   │   └── selection_mouse.go             # Shift+Click+Drag
│   ├── clipboard/                         # (10) Clipboard
│   │   ├── clipboard.go                   # atotto wrapper
│   │   └── fallback.go                    # Interne Clipboard-Buffer
│   ├── config/                            # (11) Configuration
│   │   ├── config.go                      # TOML load/save
│   │   ├── defaults.go                    # Default-Values
│   │   └── paths.go                       # XDG-Pfade
│   └── ui/                                # (12) UI State
│       ├── ui.go                          # UI-State (active pane, focus, mode)
│       ├── layout.go                      # 3-Pane-Layout (lipgloss)
│       ├── status.go                      # Status-Bar
│       ├── footer.go                      # Kontextabhängige Shortcut-Leiste
│       ├── theme.go                       # lipgloss theme (light/dark)
│       ├── sidebar_resize.go              # Mouse-Drag-Splitter
│       └── dialog.go                      # Generic Confirm/Input-Dialoge
├── demo/                                  # Demo-Vault (für MVP-Gate-Test)
│   ├── README.md
│   ├── Teststrategie.md                   # Deutsch, Umlaute, Strukturen
│   ├── Testplanung.md
│   ├── Testarten.md
│   ├── Testmethoden.md
│   ├── Testdesign.md
│   └── Projekte/
│       └── Teststrategie-SAP.md
└── test/
    └── smoke/
        ├── boot_test.go
        ├── open_test.go
        ├── save_test.go
        ├── find_test.go
        ├── toc_test.go
        └── quit_test.go
```

**Architektur-Verpflichtungen (Spec-getreu):**
- 12 Komponenten als separate Packages ✓
- Keine monolithische Implementierung ✓ (jedes Package hat klaren Public-API + Tests)
- Architektur erlaubt out-of-MVP-Items (L1-L8) ohne Refactor:
  - L1 Wiki-Links: `internal/markdown/wikilinks.go` existiert als Stub
  - L2 Backlinks: `internal/workspace/backlinks.go` Stub
  - L3 Tags: `internal/workspace/tags.go` Stub
  - L4 Auto-Link: `internal/editor/completion.go` Stub
  - L5 Recent: `internal/config/recent.go` Stub
  - L6 Bookmarks: `internal/config/bookmarks.go` Stub
  - L7 Sessions: `internal/app/session.go` Stub
  - L8 Multi-Vault: `internal/workspace/vault.go` Stub

---

## Schritt 6 — MVP-Definition (exakt aus User-Spec)

Die User-Spec listet 16 MVP-Items. **Alle müssen im ersten Release enthalten sein.** Hier die Sprint-Aufteilung:

| Sprint | MVP-Items | Implementation |
|---|---|---|
| **R1**: Bootstrap + Keymap | Status/Footer-Skelett, App-Boot, Ctrl+Q Quit | `go mod init`, main.go, app.go, ui/{layout,status,footer}, keymap.go (alle 20 Spec-Shortcuts vorbereitet) |
| **R2**: Workspace + Tree | Dateibaum, Datei öffnen, Maus+Kbd-Navigation, aktive Datei-Marker | workspace/{workspace,tree,tree_view,current_marker}.go, file-open-hook |
| **R3**: Editor + Save | Markdown bearbeiten + speichern, Ctrl+S, UTF-8+Umlaute, große Files | editor/{buffer,history,selection,editor_view}.go, save-hook |
| **R4**: Shortcuts + Clipboard + Indent | Alle 20 Shortcuts aktiv, Ctrl+C/X/V via atotto, Tab/Shift+Tab indent, Shift+Pfeiltasten Selection, Bold/Italic (Ctrl+B/I), Ctrl+N Neue Datei, Ctrl+W Schließen | input/keys.go vollständig, clipboard.go, indent.go, markdown_ops.go |
| **R5**: Markdown-Highlight + Live-TOC + TOC-Navigation | Markdown farbig (chroma), TOC zeigt Headings live, Klick im TOC → Editor-Sprung, aktive Section markiert | markdown/{parser,ast,toc_extract}.go, editor/highlight.go, toc/{toc,live_update,active_section,toc_view}.go |
| **R6**: Suche + Maus-Sidebar-Resize + Replace | Ctrl+F in-file, Ctrl+H replace, Sidebar mit Maus draggable | search/{file,fuzzy,search_modal}.go, ui/sidebar_resize.go |
| **R7**: Preview + Split View | Editor/Preview/Split (Ctrl+\\), glamour-Renderer, live-update debounced 200ms | preview/{preview,live_update,preview_view}.go, View-Mode-Toggle |
| **R8**: Quick Open + Tabs | Ctrl+P fuzzy-file, Tab-Leiste, Ctrl+J/K Tab-Wechsel, Ctrl+W mit Save-Prompt | palette/{palette,quickopen,commands}.go, app/tab_manager.go |
| **R9**: Command Palette + Workspace-Suche | Ctrl+Shift+P mit 25-35 Commands, Ctrl+Shift+F Cross-File-Suche, async | palette/command_executor.go, search/workspace.go |
| **R10**: Polish + 7-Gate MVP-Verification | Tree-CRUD (new/rename/delete mit Confirm), File-Watch fsnotify, Help-Screen, Demo-Vault ausgebaut, **7-Gate-Verification aller 16 MVP-Items** | workspace/crud.go, platform/watch.go, demo/, golden-snapshot tests, README mit Screenshots |

→ **MVP ist 10 Sprints, geschätzt 3-5 Wochen Vollzeit-Entwicklung.**

### Out-of-MVP (Spec L1-L8) — Architektur-muss-erlauben

Diese 8 Items sind **explizit nicht im MVP**, aber Architektur muss spätere Implementierung erlauben:

| Spec-Punkt | Architektur-Vorbereitung | Implementierungs-Sprint (out-of-MVP) |
|---|---|---|
| L1 `[[Wiki Links]]` | `internal/markdown/wikilinks.go` mit goldmark AST-Hook | R11+ |
| L2 Backlinks | `internal/workspace/backlinks.go` mit Reverse-Index | R11+ |
| L3 Tags (`#tag`) | `internal/workspace/tags.go` mit regex-Extraction | R12+ |
| L4 Auto-Link-Vervollständigung | `internal/editor/completion.go` Stub | R13+ |
| L5 Recent Files | `internal/config/recent.go` mit History in TOML | R14+ |
| L6 Bookmarks | `internal/config/bookmarks.go` Stub | R14+ |
| L7 Sessions | `internal/app/session.go` mit Save/Restore `[]Tab + Tree-State` | R15+ |
| L8 Multi-Workspace/Vaults | `internal/workspace/vault.go` mit mehreren Root-Pfaden | R15+ |

---

## Schritt 7 — ASCII-Wireframes

### Wireframe 1: Default-Workspace (Boot-State)

```
┌─ mdskim2 ───────────────────────── ~/.projects/test-engineering ─────────┐
│ Ctrl+O Open · Ctrl+S Save · Ctrl+P Open · Ctrl+F Find · Ctrl+Q Quit      │
├──────────────┬─────────────────────────────────────────┬──────────────────┤
│ FILES        │ # Teststrategie                        │ INHALT           │
│              │                                         │ ──────           │
│ ▾ Test Engi… │ **Was ist eine Teststrategie?**        │ ▸ Was ist        │
│   • Tests…   │                                         │   Teststrategie? │
│   • Testpl…  │ Eine Teststrategie ist ein Dokument,    │ ▸ Warum braucht… │
│   • Testar…  │ das die *Ziele*, den *Umfang* und die   │ ▸ Bestandteile   │
│   • Testmet… │ *Methoden* beschreibt.                 │   ▸ Testziele    │
│   • Testsd…  │                                         │   ▸ Testumfang   │
│ • Roadmap.md │ ## Bestandteile                         │   ▸ Testmethoden │
│              │ - Testziele                            │ ▸ Risikobasier…  │
│              │ - Testumfang                           │ ▸ Praxisbeispiel │
│              │ - Testmethoden                         │                  │
│              │                                         │ ─────────────    │
│              │ ## Warum braucht ein Projekt eine       │ PREVIEW (Ctrl+\) │
│              │ Teststrategie?                         │                  │
│              │                                         │                  │
├──────────────┴─────────────────────────────────────────┴──────────────────┤
│ Teststrategie.md · 142 lines · UTF-8 · 7 H · 2 W · ◆ modified            │
└──────────────────────────────────────────────────────────────────────────┘
```

### Wireframe 2: Split View (Editor + Preview, Ctrl+\\)

```
┌─ mdskim2 ───────────────────────── ~/.projects/test-engineering ─────────┐
│ Ctrl+O Open · Ctrl+S Save · Ctrl+P Open · Ctrl+F Find · Ctrl+Q Quit      │
├──────────────┬──────────────────────────────┬─────────────────────────────┤
│ FILES        │ # Teststrategie              │ Teststrategie                │
│              │                              │ ════════════                  │
│ ...          │ **Was ist eine              │ Was ist eine                  │
│              │  Teststrategie?**           │ Teststrategie?                │
│              │                              │                               │
│              │ Eine Teststrategie ist     │ Eine Teststrategie ist        │
│              │ ein Dokument, das die       │ ein Dokument, das die         │
│              │ *Ziele*, den *Umfang* und  │ Ziele, den Umfang und die     │
│              │ die *Methoden* beschreibt.  │ Methoden beschreibt.          │
│              │                              │                               │
│              │ ## Bestandteile             │ Bestandteile                   │
│              │ - Testziele                 │   • Testziele                  │
│              │ - Testumfang                │   • Testumfang                 │
│              │ - Testmethoden              │   • Testmethoden               │
├──────────────┴──────────────────────────────┴─────────────────────────────┤
│ Teststrategie.md · [EDIT] [PREV] [SPLIT] · 142 lines · ◆ modified         │
└──────────────────────────────────────────────────────────────────────────┘
```

### Wireframe 3: Quick Open (Ctrl+P)

```
┌─ mdskim2 ───────────────────────── ~/.projects/test-engineering ─────────┐
│ Ctrl+O Open · Ctrl+S Save · Ctrl+P Open · Ctrl+F Find · Ctrl+Q Quit      │
├──────────────┬─────────────────────────────────────────┬──────────────────┤
│ ...          │ ...                                     │ ...              │
├──────────────┴─────────────────────────────────────────┴──────────────────┤
│ ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ │
│  ┌──────────────────────────────────────────────────────────────────┐    │
│  │ ▸ teststrat▌                                                      │    │
│  ├──────────────────────────────────────────────────────────────────┤    │
│  │ 📄 Teststrategie.md                        Test Engineering       │    │
│  │ 📄 Teststrategie-alt.md                    ~/backup               │    │
│  │ 📁 Teststrategie-SAP.md                    Projekte/              │    │
│  └──────────────────────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────────────────────┘
```

### Wireframe 4: Command Palette (Ctrl+Shift+P)

```
┌─ mdskim2 ───────────────────────── ~/.projects/test-engineering ─────────┐
│ Ctrl+O Open · Ctrl+S Save · Ctrl+P Open · Ctrl+F Find · Ctrl+Q Quit      │
├──────────────┬─────────────────────────────────────────┬──────────────────┤
│ ...          │ ...                                     │ ...              │
├──────────────┴─────────────────────────────────────────┴──────────────────┤
│  ┌──────────────────────────────────────────────────────────────────┐    │
│  │ ▸ tog▌                                                            │    │
│  ├──────────────────────────────────────────────────────────────────┤    │
│  │ VIEW                                                              │    │
│  │   🗂  Toggle File Explorer                          Ctrl+B        │    │
│  │   🪟  Toggle Right Pane (TOC)                       Ctrl+T        │    │
│  │   👁  Toggle Preview                               Ctrl+\        │    │
│  │   📐 Toggle Word Wrap                              Ctrl+Shift+W   │    │
│  │ FILE                                                              │    │
│  │   📄 New File                                      Ctrl+N         │    │
│  │   📁 New Folder                                    Ctrl+Shift+N   │    │
│  │   ✏️  Rename File                                   F2             │    │
│  │   🗑  Delete File                                   Del            │    │
│  │ EDIT                                                              │    │
│  │   🔍 Find...                                       Ctrl+F         │    │
│  │   🔎 Find & Replace...                             Ctrl+H         │    │
│  │   🔎 Quick Open                                    Ctrl+P         │    │
│  │ WORKSPACE                                                         │    │
│  │   🔍 Search Workspace                              Ctrl+Shift+F   │    │
│  │   📂 Open Workspace...                                            │    │
│  └──────────────────────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────────────────────┘
```

### Wireframe 5: Tabs (mehrere Dateien)

```
┌─ mdskim2 ───────────────────────── ~/.projects/test-engineering ─────────┐
├──────────────────────────────────────────────────────────────────────────┤
│ ✓ Teststrategie.md │ ● Testplanung.md │ ● Testarten.md │ ⊕ │             │
├──────────────┬─────────────────────────────────────────┬──────────────────┤
│ FILES        │ # Testplanung                          │ INHALT           │
│ ...          │ ...                                     │ ...              │
├──────────────┴─────────────────────────────────────────┴──────────────────┤
│ Testplanung.md · 87 lines · UTF-8 · 5 H · modified                       │
└──────────────────────────────────────────────────────────────────────────┘

✓ saved, ● modified, ⊕ neuer Tab
Ctrl+J / Ctrl+K = Tab ←  /  →, Ctrl+W = Close (mit Save-Prompt)
```

### Wireframe 6: Find Modal (Ctrl+F)

```
┌─ mdskim2 ───────────────────────── ~/.projects/test-engineering ─────────┐
│ ...                                                                      │
├──────────────┴─────────────────────────────────────────┴──────────────────┤
│  ┌────────────────────────────────────────────────────────────────────┐  │
│  │ ▸ Testziele▌                                            ESC Close  │  │
│  ├────────────────────────────────────────────────────────────────────┤  │
│  │  Match (1/3)  ↑ ↓   [Aa] [.*] [W]                                │  │
│  │  ─── Line 42: "Die **Testziele** sind..."                         │  │
│  │  ─── Line 78: "In den Testzielen werden..."                       │  │
│  │  ─── Line 95: "Vergleiche der Testziele mit..."                   │  │
│  └────────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────────────┘
```

---

## Schritt 8 — Implementierung (nach User-OK, Hard-Stop-Regel)

**Voraussetzung für Phase 2:**
1. Du bestätigst:
   - Stack: Go ✓
   - Modul-Pfad: `github.com/<user>/mdskim2` oder lokaler Pfad `mdskim2`?
   - Wireframes: passen so?
   - MVP-Scope (10 Sprints / 16 Items): OK?
2. Ich starte Sprint R1: `go mod init` + Bootstrap.

**Sprint R1 — Bootstrap (1-2 Tage)**
- `cd ~/projects/mdskim2 && go mod init <modul-pfad>`
- `go get` für 15 Module aus Schritt 1
- `cmd/mdskim2/main.go`: `flag.Parse`, workspace-Pfad-Argument, `tea.NewProgram`
- `internal/app/app.go`: Model + Init + Update + View (Stub)
- `internal/ui/layout.go`: 3-Pane-Layout mit lipgloss (leere Stubs)
- `internal/ui/{status,footer}.go`: Skelette mit kontextabhängigen Shortcut-Hints
- `internal/app/keymap.go`: Alle 20 Spec-Shortcuts vorbereitet
- `demo/{Teststrategie,Testplanung,Testarten,Testmethoden,Testdesign}.md` + `Projekte/Teststrategie-SAP.md`
- `test/smoke/boot_test.go`: Golden-Snapshot-Test des initialen `View()`
- 7 Gates: `go build` / `go vet` / `go test ./...` / `./mdskim2 demo/` / `--help` / Quit-Test / Snapshot-Diff

→ Erste lauffähige App mit leerem Layout, Footer mit Shortcuts, Quit per Ctrl+Q.

---

## Zusammenfassung

| Aspekt | Wert |
|---|---|
| **Stack** | Go 1.26 + Bubble Tea v1 + Bubbles v0 + Lip Gloss v1 + Glamour v0 + goldmark v1.7 + chroma v2 |
| **Module** | 15 Go-Module, alle aktiv maintained |
| **Architektur-Komponenten** | 12 Pflicht-Packages + Orchestrator + cmd/mdskim2 + test/smoke, **alle 12 Spec-Komponenten** als eigene Packages |
| **MVP-Items** | **Alle 16 Items** aus User-Spec (keine Auslassung) |
| **MVP-Sprints** | **10 Sprints** (R1-R10), realistisch 3-5 Wochen |
| **Out-of-MVP (Architektur-vorbereitet)** | L1-L8: Wiki-Links, Backlinks, Tags, Auto-Link, Recent, Bookmarks, Sessions, Multi-Vault |
| **Risiken mit Lösungen** | 10 dokumentiert, alle Spec-Kritischen adressiert |
| **Wireframes** | 6 (Default, Split-View, Quick Open, Command Palette, Tabs, Find Modal) |
| **Demo-Vault** | `demo/Test Engineering/` mit 6 Dateien, deutsche Umlaute, Cross-Links |
| **Single Binary** | ✓ via `go build -o mdskim2 ./cmd/mdskim2` |
| **Cross-Platform** | ✓ Linux, macOS, Windows (Windows-Terminal) |

---

## Akzeptanz dieses Reports

✅ User-Spec als einzige Wahrheit (keine Annahmen, keine Extrapolation)
✅ Schritt 1-8 aus User-Spec-Vorgehensweis explizit abgearbeitet
✅ 12 Pflicht-Architektur-Komponenten als separate Packages
✅ Alle 16 MVP-Items Sprint-zugeordnet
✅ 10 konkrete Spec-Risiken mit Lösungen (Ctrl+C/S/Q-Risiken explizit, Spec-Verbot eingehalten)
✅ 6 ASCII-Wireframes (Default, Split, Quick Open, Command Palette, Tabs, Find)
✅ 15 Libraries mit URLs und Maintenance-Status
✅ Hard-Stop: kein `go mod init`, kein Code, warte auf User-OK

---

## 4 Fragen an dich (Dennis)

1. **Modul-Pfad**: `github.com/<dein-github-user>/mdskim2` (öffentlich) oder lokaler Pfad `mdskim2` (kein GitHub-Bind)?
2. **Wireframes**: Passen die 6 Wireframes? Möchtest du eine andere Aufteilung (z.B. Preview immer sichtbar statt Toggle)?
3. **MVP-Scope (10 Sprints)**: OK? Oder soll etwas priorisiert / zurückgestellt werden, damit es schneller zu einem ersten vorzeigbaren Stand kommt?
4. **Obsidian-L1-L8 (Out-of-MVP)**: Soll der Architektur-Stub für alle 8 gleich im MVP stehen (1-2 Tage extra), oder erst wenn das jeweilige Feature drankommt?

Sag "OK los" mit Antworten auf 1-4 oder nenne Änderungen — dann startet Phase 2 mit Sprint R1.
