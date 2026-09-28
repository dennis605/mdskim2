// Package app implementiert die mdskim2-Anwendung als Bubble-Tea-Model.
package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/dennis605/mdskim2/internal/backlinks"
	"github.com/dennis605/mdskim2/internal/editor"
	"github.com/dennis605/mdskim2/internal/grep"
	"github.com/dennis605/mdskim2/internal/markdown"
	"github.com/dennis605/mdskim2/internal/palette"
	"github.com/dennis605/mdskim2/internal/preview"
	"github.com/dennis605/mdskim2/internal/recent"
	"github.com/dennis605/mdskim2/internal/search"
	"github.com/dennis605/mdskim2/internal/tabs"
	"github.com/dennis605/mdskim2/internal/ui"
	"github.com/dennis605/mdskim2/internal/workspace"
	"regexp"
	"sort"
)

type Model struct {
	workspacePath string
	width         int
	height        int
	quitting      bool

	workspace  *workspace.Workspace
	treeRender *workspace.TreeRenderer
	cursorIdx  int
	flatList   []*workspace.FileNode

	buffer      *editor.Buffer
	clipboard   string
	selection   string
	mode        string
	currentFile string

	// R6: Search-Modal
	searchQuery   string
	searchResults []search.Match
	searchIdx     int
	searchActive  bool

	// R7: Preview-Mode
	previewMode  bool   // wenn true, zeige Preview statt Editor-Buffer
	previewSplit bool   // Split-View (Editor + Preview)
	previewCache string // cached preview render

	// R8: Tabs + Quick-Open
	tabs           *tabs.Manager
	quickOpenMode  bool
	quickOpenQuery string
	quickOpenList  []string

	theme   ui.Theme
	layout  ui.Layout
	version string

	// R9: Command Palette
	palette       *palette.Registry
	paletteActive bool
	paletteQuery  string

	// R10: Workspace-Grep
	workspaceHits []grep.Hit

	saveError string

	// U2 Focus-Modell: "tree" oder "editor"
	focus string

	// U3: Editor-Features
	lineWrap           bool
	goToLineMode       bool
	editorScrollOffset int
	goToLineQuery      string

	// U4: Editor-Features (Read-Only, Whitespace, Bracket-Match)
	readOnly       bool
	whitespaceMark bool // Ctrl+Shift+W — Zeige Tabs und trailing Spaces

	// U4: Persistence
	recent       *recent.List
	autoSaveTick int
	autoSavedAt  string // Zeitstempel für Anzeige

	// U6: Right-Pane-Tab: 0=Preview | 1=TOC | 2=Backlinks
	rightTab int

	// U6: Backlinks cache
	backlinksCache []backlinks.Entry
	backlinksFor   string

	// U8: Tree-based file operations + tree filter
	promptMode   string // "" | "new-file" | "new-folder" | "rename" | "confirm-delete"
	promptQuery  string
	treeFilter   string
	treeFiltered bool

	// U8.4: Wiki-Link-Autocomplete beim Tippen von [[
	wikiLinkPopup   bool
	wikiLinkQuery   string
	wikiLinkMatches []string
	wikiLinkIdx     int

	// U9: Discoverability (Obsidian-style)
	findInFilesMode    bool
	findInFilesQuery   string
	findInFilesResults []grep.Hit
	findInFilesIdx     int
	recentMenuMode     bool
	recentMenuIdx      int
}

func New(workspacePath string) Model {
	rec := recent.New()
	ws, _ := workspace.Load(workspacePath)
	r := workspace.NewTreeRenderer()
	flat := ws.FlatList(r.CollapsedDirs)

	m := Model{
		recent:        rec,
		workspacePath: workspacePath,
		width:         120,
		height:        40,
		mode:          "EDIT",
		focus:         "tree",
		version:       "R10: Polish + Stubs",
		theme:         ui.Light(),
		layout:        ui.DefaultLayout(),
		workspace:     ws,
		treeRender:    r,
		flatList:      flat,
		tabs:          tabs.NewManager(),
		palette:       palette.NewRegistry(),
	}
	m.layout.Compute(m.width, m.height)

	// Register default commands
	m.palette.Register(palette.Command{Name: "save", Description: "Save current file", Keywords: []string{"write", "store", "ctrl+s"}})
	m.palette.Register(palette.Command{Name: "open", Description: "Open file", Keywords: []string{"load", "read"}})
	m.palette.Register(palette.Command{Name: "preview", Description: "Toggle preview", Keywords: []string{"render", "glamour"}})
	m.palette.Register(palette.Command{Name: "find", Description: "Open search modal", Keywords: []string{"search", "fuzzy"}})
	m.palette.Register(palette.Command{Name: "replace", Description: "Replace in file", Keywords: []string{"substitute"}})
	m.palette.Register(palette.Command{Name: "bold", Description: "Insert bold markup", Keywords: []string{"strong"}})
	m.palette.Register(palette.Command{Name: "italic", Description: "Insert italic markup", Keywords: []string{"em"}})
	m.palette.Register(palette.Command{Name: "quit", Description: "Quit application", Keywords: []string{"exit", "close"}})
	m.palette.Register(palette.Command{Name: "recent", Description: "Open Recent Files list", Keywords: []string{"recent", "history"}, OnRun: func() {
		m.paletteActive = false
		// Build quick-open for recent
		m.quickOpenMode = true
		m.quickOpenQuery = ""
		m.quickOpenList = nil
		for _, rf := range m.recent.All() {
			m.quickOpenList = append(m.quickOpenList, rf.Display+": "+rf.Path)
		}
	}})
	m.palette.Register(palette.Command{Name: "readonly", Description: "Toggle Read-Only mode for current file", Keywords: []string{"ro", "lock", "readonly"}, OnRun: func() {
		m.readOnly = !m.readOnly
	}})
	m.palette.Register(palette.Command{Name: "lineending", Description: "Show line ending indicator for current file", Keywords: []string{"lf", "crlf", "le"}, OnRun: func() {
		// Cycle LF -> CRLF -> CR -> LF
		if m.buffer != nil {
			switch m.buffer.LineEnding {
			case "\n":
				m.buffer.LineEnding = "\r\n"
			case "\r\n":
				m.buffer.LineEnding = "\r"
			default:
				m.buffer.LineEnding = "\n"
			}
		}
	}})
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.Tick(30*time.Second, func(t time.Time) tea.Msg { return autoSaveTickMsg(t) }),
	)
}

type autoSaveTickMsg time.Time

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case autoSaveTickMsg:
		if m.buffer != nil && m.buffer.Modified && m.buffer.Path != "" {
			err := m.buffer.Save()
			if err == nil {
				m.saveError = ""
				m.autoSavedAt = time.Now().Format("15:04:05")
			} else {
				m.saveError = err.Error()
			}
		}
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layout.Compute(m.width, m.height)
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.MouseMsg:
		return m.handleMouse(msg)
	case DailyNoteTriggerMsg:
		return m.openOrCreateDailyNote()
	case FindInFilesTriggerMsg:
		m.findInFilesMode = true
		m.findInFilesQuery = ""
		m.findInFilesResults = nil
		m.findInFilesIdx = 0
		return m, nil
	case RecentTriggerMsg:
		m.recentMenuMode = true
		m.recentMenuIdx = 0
		return m, nil
	case TaskToggleTriggerMsg:
		m.toggleCurrentTaskLine()
		return m, nil
	}
	return m, nil
}

// DailyNoteTriggerMsg ist ein Test-Hook, der Ctrl+Shift+D auslöst ohne
// den Umweg über tea.KeyMsg. Wird nur in Tests verwendet.
type DailyNoteTriggerMsg struct{}

// FindInFilesTriggerMsg ist ein Test-Hook für Ctrl+Shift+F.
type FindInFilesTriggerMsg struct{}

// RecentTriggerMsg ist ein Test-Hook für Ctrl+Shift+O.
type RecentTriggerMsg struct{}

// TaskToggleTriggerMsg ist ein Test-Hook für Ctrl+Enter (Task-List-Toggle).
type TaskToggleTriggerMsg struct{}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// U8: Prompt-Mode für Tree-File-Operations hat Vorrang
	if m.promptMode != "" {
		switch msg.Type {
		case tea.KeyEsc:
			m.promptMode = ""
			m.promptQuery = ""
			return m, nil
		case tea.KeyEnter:
			return m.executePrompt()
		case tea.KeyBackspace:
			if len(m.promptQuery) > 0 {
				m.promptQuery = m.promptQuery[:len(m.promptQuery)-1]
			}
			return m, nil
		}
		if msg.Type == tea.KeyRunes {
			for _, r := range msg.Runes {
				m.promptQuery += string(r)
			}
			return m, nil
		}
		return m, nil
	}

	// U8.3: Tree-Filter-Mode — Esc leert, Backspace kürzt, Runes ergänzen, Enter akzeptiert
	if m.treeFiltered {
		switch msg.Type {
		case tea.KeyEsc:
			m.treeFiltered = false
			m.treeFilter = ""
			m.applyTreeFilter()
			return m, nil
		case tea.KeyBackspace:
			if len(m.treeFilter) > 0 {
				m.treeFilter = m.treeFilter[:len(m.treeFilter)-1]
				m.applyTreeFilter()
			}
			return m, nil
		case tea.KeyEnter:
			m.treeFiltered = false
			return m, nil
		case tea.KeyRunes:
			m.treeFilter += string(msg.Runes)
			m.applyTreeFilter()
			return m, nil
		}
	}

	// U8.4: Wiki-Link-Popup-Interceptor (Esc/Arrow/Enter/Runes)
	if m.wikiLinkPopup {
		switch msg.Type {
		case tea.KeyEsc:
			m.wikiLinkPopup = false
			m.wikiLinkQuery = ""
			return m, nil
		case tea.KeyEnter, tea.KeyTab:
			m.acceptWikiLink()
			return m, nil
		case tea.KeyUp:
			if m.wikiLinkIdx > 0 {
				m.wikiLinkIdx--
			}
			return m, nil
		case tea.KeyDown:
			if m.wikiLinkIdx < len(m.wikiLinkMatches)-1 {
				m.wikiLinkIdx++
			}
			return m, nil
		case tea.KeyBackspace:
			if len(m.wikiLinkQuery) > 0 {
				m.wikiLinkQuery = m.wikiLinkQuery[:len(m.wikiLinkQuery)-1]
				m.refreshWikiLinkMatches()
			}
			return m, nil
		case tea.KeyRunes:
			m.wikiLinkQuery += string(msg.Runes)
			m.refreshWikiLinkMatches()
			return m, nil
		}
	}

	// U9.3: Recent-Files-Modal-Interceptor
	if m.recentMenuMode {
		files := m.recentFilesList()
		switch msg.Type {
		case tea.KeyEsc:
			m.recentMenuMode = false
			return m, nil
		case tea.KeyEnter:
			if m.recentMenuIdx < len(files) {
				m.recentMenuMode = false
				m.openRecentFile(files[m.recentMenuIdx].Path)
			}
			return m, nil
		case tea.KeyUp:
			if m.recentMenuIdx > 0 {
				m.recentMenuIdx--
			}
			return m, nil
		case tea.KeyDown:
			if m.recentMenuIdx < len(files)-1 {
				m.recentMenuIdx++
			}
			return m, nil
		case tea.KeyRunes:
			// Optionaler Filter-Modus: aktuell nicht implementiert (Esc zum Schließen)
			return m, nil
		}
	}

	// U9.2: Find-in-Files-Modal-Interceptor
	if m.findInFilesMode {
		switch msg.Type {
		case tea.KeyEsc:
			m.findInFilesMode = false
			m.findInFilesQuery = ""
			m.findInFilesResults = nil
			return m, nil
		case tea.KeyEnter:
			if m.findInFilesIdx < len(m.findInFilesResults) {
				hit := m.findInFilesResults[m.findInFilesIdx]
				m.findInFilesMode = false
				m.findInFilesQuery = ""
				m.openGrepHit(hit)
			}
			return m, nil
		case tea.KeyUp:
			if m.findInFilesIdx > 0 {
				m.findInFilesIdx--
			}
			return m, nil
		case tea.KeyDown:
			if m.findInFilesIdx < len(m.findInFilesResults)-1 {
				m.findInFilesIdx++
			}
			return m, nil
		case tea.KeyBackspace:
			if len(m.findInFilesQuery) > 0 {
				m.findInFilesQuery = m.findInFilesQuery[:len(m.findInFilesQuery)-1]
				m.runFindInFiles()
			}
			return m, nil
		case tea.KeyRunes:
			m.findInFilesQuery += string(msg.Runes)
			m.runFindInFiles()
			return m, nil
		}
	}

	switch msg.Type {
	case tea.KeyCtrlQ:
		m.quitting = true
		return m, tea.Quit
	case tea.KeyCtrlS:
		return m.saveBuffer()
	case tea.KeyCtrlZ:
		if m.buffer != nil {
			m.buffer.Undo()
		}
		return m, nil
	case tea.KeyCtrlY:
		if m.buffer != nil {
			m.buffer.Redo()
		}
		return m, nil
	case tea.KeyCtrlC:
		if m.buffer != nil {
			line := m.buffer.Lines[m.buffer.CursorRow]
			clipboard.WriteAll(line)
			m.clipboard = line
		}
		return m, nil
	case tea.KeyCtrlX:
		if m.buffer != nil {
			line := m.buffer.Lines[m.buffer.CursorRow]
			clipboard.WriteAll(line)
			m.clipboard = line
			m.buffer.DeleteLine()
		}
		return m, nil
	case tea.KeyCtrlP:
		// Toggle Preview-Mode
		if m.buffer != nil {
			m.previewMode = !m.previewMode
			if m.previewMode {
				rendered, err := preview.Render(m.buffer.ToString())
				if err == nil {
					m.previewCache = rendered
				}
			}
		}
		return m, nil

	case tea.KeyCtrlT:
		// Toggle Quick-Open
		m.quickOpenMode = !m.quickOpenMode
		if m.quickOpenMode {
			m.quickOpenList = nil
			for _, node := range m.flatList {
				m.quickOpenList = append(m.quickOpenList, node.Path)
			}
		} else {
			m.quickOpenQuery = ""
		}
		return m, nil

	case tea.KeyCtrlK:
		// Toggle Command Palette
		m.paletteActive = !m.paletteActive
		if !m.paletteActive {
			m.paletteQuery = ""
		}
		return m, nil

	case tea.KeyCtrlD:
		// Zeile duplizieren
		if m.buffer != nil && m.focus == "editor" {
			m.buffer.DuplicateLine()
		}
		return m, nil

	case tea.KeyCtrlW:
		if msg.String() == "ctrl+shift+w" {
			// Whitespace-Indicator Toggle
			m.whitespaceMark = !m.whitespaceMark
			return m, nil
		}
		// Soft-Wrap Toggle
		m.lineWrap = !m.lineWrap
		return m, nil

	case tea.KeyF6:
		// Cycle pane focus forward
		next := m.cycleFocus(1)
		m = m.setFocus(next)
		return m, nil
	case tea.KeyF5:
		// Refresh Tree from disk
		m.refreshWorkspace()
		return m, nil
	case tea.KeyCtrlR:
		if msg.String() == "ctrl+shift+r" {
			// Read-Only Toggle (Ctrl+Shift+R)
			m.readOnly = !m.readOnly
			return m, nil
		}
		if msg.String() == "ctrl+l" {
			// Reload from disk (Ctrl+L)
			if m.buffer != nil && m.buffer.Path != "" {
				buf, err := editor.LoadFromFile(m.buffer.Path)
				if err == nil {
					m.buffer = buf
					return m, nil
				}
			}
			return m, nil
		}

	case tea.KeyCtrlG:
		// Go-to-line Modal
		m.goToLineMode = !m.goToLineMode
		if !m.goToLineMode {
			m.goToLineQuery = ""
		}
		return m, nil

	case tea.KeyCtrlV:
		if m.buffer != nil {
			text, err := clipboard.ReadAll()
			if err == nil && text != "" {
				m.buffer.InsertString(text)
				m.clipboard = text
			}
		}
		return m, nil
	case tea.KeyCtrlB:
		if m.buffer != nil {
			m.buffer.InsertString("****")
			m.buffer.MoveCursor(0, -2)
		}
		return m, nil
	case tea.KeyCtrlF:
		// Toggle Search-Modal
		m.searchActive = !m.searchActive
		if m.searchActive && m.buffer == nil {
			m.searchActive = false
		}
		if !m.searchActive {
			m.searchQuery = ""
			m.searchResults = nil
			m.searchIdx = 0
		}
		return m, nil

	case tea.KeyCtrlH:
		// Replace: einfach Ersetzen aller Vorkommen (ein-Schritt-Modal)
		if m.buffer == nil || m.searchQuery == "" {
			return m, nil
		}
		opt := search.Options{CaseSensitive: false}
		newLines, count := search.ReplaceAll(m.buffer.Lines, m.searchQuery, "REPLACED", opt)
		if count > 0 {
			m.buffer.Lines = newLines
			m.buffer.Modified = true
			m.buffer.SnapshotHistory()
		}
		m.mode = fmt.Sprintf("REPLACE: %d", count)
		return m, nil

	case tea.KeyCtrlI:
		if m.buffer != nil {
			m.buffer.InsertString("**")
			m.buffer.MoveCursor(0, -1)
		}
		return m, nil
	case tea.KeyEsc:
		// Esc priorisiert: Modals > Tree > Quit
		if m.searchActive {
			m.searchActive = false
			return m, nil
		}
		if m.quickOpenMode {
			m.quickOpenMode = false
			return m, nil
		}
		if m.paletteActive {
			m.paletteActive = false
			return m, nil
		}
		if m.goToLineMode {
			m.goToLineMode = false
			m.goToLineQuery = ""
			return m, nil
		}
		// Esc: cycle focus back one pane (replaces "always go to tree")
		prev := m.cycleFocus(-1)
		m = m.setFocus(prev)
		return m, nil
	case tea.KeyUp:
		if m.focus == "editor" && m.buffer != nil {
			m.buffer.MoveCursor(-1, 0)
		} else {
			// Tree-Navigation: Pfeiltasten bleiben im Tree
			m.focus = "tree"
			if m.cursorIdx > 0 {
				m.cursorIdx--
				m.autoOpenAtCursor()
			}
		}
		return m, nil
	case tea.KeyDown:
		if m.focus == "editor" && m.buffer != nil {
			m.buffer.MoveCursor(1, 0)
		} else {
			// Tree-Navigation
			m.focus = "tree"
			if m.cursorIdx < len(m.flatList)-1 {
				m.cursorIdx++
				m.autoOpenAtCursor()
			}
		}
		return m, nil
	case tea.KeyLeft:
		if m.focus == "editor" && m.buffer != nil {
			m.buffer.MoveCursor(0, -1)
		} else if m.cursorIdx >= 0 && m.cursorIdx < len(m.flatList) {
			node := m.flatList[m.cursorIdx]
			if node.IsDir {
				// Im Tree: Links auf collapsed dir → expand, expanded → focus tree
				m.treeRender.ToggleDir(node.Path)
				m.flatList = m.workspace.FlatList(m.treeRender.CollapsedDirs)
			}
		}
		return m, nil
	case tea.KeyRight:
		if m.focus == "editor" && m.buffer != nil {
			m.buffer.MoveCursor(0, 1)
		} else if m.cursorIdx >= 0 && m.cursorIdx < len(m.flatList) {
			node := m.flatList[m.cursorIdx]
			if node.IsDir {
				// Im Tree: Rechts auf collapsed dir → expand
				m.treeRender.ToggleDir(node.Path)
				m.flatList = m.workspace.FlatList(m.treeRender.CollapsedDirs)
			}
		}
		return m, nil
	case tea.KeyHome:
		if m.focus == "editor" && m.buffer != nil {
			m.buffer.Home()
		} else {
			m.focus = "tree"
			m.cursorIdx = 0
			m.autoOpenAtCursor()
		}
		return m, nil
	case tea.KeyEnd:
		if m.focus == "editor" && m.buffer != nil {
			m.buffer.End()
		} else {
			m.focus = "tree"
			if len(m.flatList) > 0 {
				m.cursorIdx = len(m.flatList) - 1
				m.autoOpenAtCursor()
			}
		}
		return m, nil
	case tea.KeyCtrlJ:
		// U9.5: Ctrl+Enter (LF) — Task-List-Toggle falls Cursor auf Task-Zeile
		if m.focus == "editor" && m.buffer != nil {
			m.toggleCurrentTaskLine()
		}
		return m, nil
	case tea.KeyEnter:
		// Go-to-line Modal: Enter führt Sprung aus
		if m.goToLineMode {
			if m.buffer != nil {
				var n int
				fmt.Sscanf(m.goToLineQuery, "%d", &n)
				m.buffer.GoToLine(n)
				m.focus = "editor"
			}
			m.goToLineMode = false
			m.goToLineQuery = ""
			return m, nil
		}
		// Search-Modal: Enter zyklisch durch Matches
		if m.searchActive && len(m.searchResults) > 0 {
			m.searchIdx++
			if m.searchIdx >= len(m.searchResults) {
				m.searchIdx = 0
			}
			match := m.searchResults[m.searchIdx]
			if m.buffer != nil {
				m.buffer.CursorRow = match.Line
				m.buffer.CursorCol = match.Col
				m.focus = "editor"
			}
			return m, nil
		}
		// Im Tree: Enter toggelt directory ODER wechselt in Editor
		if m.focus == "tree" && m.buffer != nil {
			node := m.flatList[m.cursorIdx]
			if node.IsDir {
				m.treeRender.ToggleDir(node.Path)
				m.flatList = m.workspace.FlatList(m.treeRender.CollapsedDirs)
				return m, nil
			}
			// File: Wechsel in Editor-Focus
			m.focus = "editor"
			return m, nil
		}
		// Im Editor ohne Modals: Enter = neue Zeile
		if m.buffer != nil && m.focus == "editor" {
			m.buffer.InsertNewLine()
			return m, nil
		}
		// Sonst: selectCurrent (legacy fallback)
		return m.selectCurrent()
	case tea.KeyBackspace:
		if m.readOnly && m.buffer != nil {
			return m, nil
		}
		m.autoSavedAt = ""
		if m.goToLineMode && m.goToLineQuery != "" {
			m.goToLineQuery = m.goToLineQuery[:len(m.goToLineQuery)-1]
			return m, nil
		}
		if m.searchActive && m.searchQuery != "" {
			if len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
			}
			if m.buffer != nil {
				m.searchResults = search.Find(m.buffer.Lines, m.searchQuery, search.Options{})
			}
			return m, nil
		}
		if m.buffer != nil && m.focus == "editor" {
			m.buffer.DeleteChar()
			return m, nil
		}
		return m, nil
	case tea.KeyDelete:
		if m.buffer != nil && m.focus == "editor" {
			m.buffer.DeleteCharForward()
		}
		return m, nil
	default:
		// U8: Tree-based file operations (a/A/d/r) im Tree-Focus
		if m.focus == "tree" && msg.Type == tea.KeyRunes {
			switch string(msg.Runes) {
			case "a":
				m.promptMode = "new-file"
				m.promptQuery = ""
				return m, nil
			case "A":
				m.promptMode = "new-folder"
				m.promptQuery = ""
				return m, nil
			case "d":
				m.promptMode = "confirm-delete"
				m.promptQuery = ""
				return m, nil
			case "r":
				m.promptMode = "rename"
				m.promptQuery = ""
				return m, nil
			case "/":
				m.treeFiltered = true
				m.treeFilter = ""
				m.applyTreeFilter()
				return m, nil
			}
		}

		// Shift+F6: cycle focus backward
		if msg.String() == "shift+f6" {
			prev := m.cycleFocus(-1)
			m = m.setFocus(prev)
			return m, nil
		}
		// U6.1: Ctrl+Shift+Arrow keys for pane navigation
		// U9: Ctrl+Shift+D (Daily Note), Ctrl+Shift+F (Find in Files), Ctrl+Shift+O (Recent)
		arrowKey := msg.String()
		if strings.HasPrefix(arrowKey, "ctrl+shift+") {
			switch strings.TrimPrefix(arrowKey, "ctrl+shift+") {
			case "up":
				panes := m.visiblePanes()
				if len(panes) > 0 {
					m = m.setFocus(panes[0])
				}
				return m, nil
			case "down":
				panes := m.visiblePanes()
				if len(panes) > 0 {
					m = m.setFocus(panes[len(panes)-1])
				}
				return m, nil
			case "left":
				prev := m.cycleFocus(-1)
				m = m.setFocus(prev)
				return m, nil
			case "right":
				next := m.cycleFocus(+1)
				m = m.setFocus(next)
				return m, nil
			case "d":
				// U9.1: Daily Note — öffnet oder erstellt <workspace>/Daily Notes/YYYY-MM-DD.md
				mm, _ := m.openOrCreateDailyNote()
				return mm, nil
			case "f":
				// U9.2: Find in Files — öffnet Workspace-Grep-Modal
				m.findInFilesMode = true
				m.findInFilesQuery = ""
				m.findInFilesResults = nil
				return m, nil
			case "o":
				// U9.3: Recent Files — öffnet Recent-Menü
				m.recentMenuMode = true
				m.recentMenuIdx = 0
				return m, nil
			}
		}
		// Alt+1..4: direct pane jump
		key := msg.String()
		if strings.HasPrefix(key, "alt+") && len(key) == 5 {
			digit := key[len(key)-1]
			if digit >= '1' && digit <= '4' {
				panes := m.visiblePanes()
				idx := int(digit - '1')
				if idx < len(panes) {
					m = m.setFocus(panes[idx])
				}
				return m, nil
			}
			// U6: Alt+5/6/7 = right-pane tab switch (Preview | TOC | Backlinks)
			// U9.4: Alt+8 = Tags tab
			if digit >= '5' && digit <= '8' {
				m.rightTab = int(digit - '5')
				m.focus = "toc"
				return m, nil
			}
		}
	}
	if msg.Type == tea.KeyRunes {
		if m.readOnly && m.buffer != nil {
			return m, nil
		}
		m.autoSavedAt = ""
		// Go-to-line Mode: Query aufbauen
		if m.goToLineMode {
			for _, r := range msg.Runes {
				if r >= '0' && r <= '9' {
					m.goToLineQuery += string(r)
				}
			}
			return m, nil
		}
		// Wenn Search aktiv: Query aufbauen
		if m.searchActive && m.buffer != nil {
			for _, r := range msg.Runes {
				m.searchQuery += string(r)
			}
			// Recompute matches
			m.searchResults = search.Find(m.buffer.Lines, m.searchQuery, search.Options{})
			m.searchIdx = 0
			return m, nil
		}
		if m.buffer != nil {
			for _, r := range msg.Runes {
				m.applyAutoPair(r)
				m.maybeTriggerWikiLink(r)
				_ = m // keep m in scope (pointer receiver changes persist via &m)
			}
		}
		return m, nil
	}
	return m, nil
}

// applyAutoPair verarbeitet ein eingetipptes Zeichen und paart es automatisch
// mit dem passenden Schluss-Zeichen (Cursor landet MITTEN im Paar, kein zweiter
// Tastendruck nötig). Bei Schluss-Zeichen wird das doppelte Schließen übersprungen,
// wenn das nächste Zeichen identisch ist.
func (m Model) applyAutoPair(r rune) {
	if m.buffer == nil {
		return
	}
	openToClose := map[rune]rune{
		'(': ')',
		'[': ']',
		'{': '}',
		'"': '"',
	}
	if close, ok := openToClose[r]; ok {
		m.buffer.InsertChar(r)
		m.buffer.InsertChar(close)
		// Cursor zwischen die beiden Zeichen setzen
		if m.buffer.CursorCol > 0 {
			m.buffer.CursorCol--
		}
		return
	}
	// Schluss-Zeichen: nächstes Zeichen prüfen, ggf. überspringen
	closing := []rune{')', ']', '}', '"'}
	for _, c := range closing {
		if r == c {
			line := m.buffer.Lines[m.buffer.CursorRow]
			runes := []rune(line)
			if m.buffer.CursorCol < len(runes) && runes[m.buffer.CursorCol] == c {
				m.buffer.CursorCol++
				return
			}
			break
		}
	}
	m.buffer.InsertChar(r)
}

func (m Model) saveBuffer() (tea.Model, tea.Cmd) {
	if m.buffer == nil {
		return m, nil
	}
	if err := m.buffer.Save(); err != nil {
		m.saveError = err.Error()
		m.mode = "SAVE-ERROR"
	} else {
		m.saveError = ""
		m.mode = "EDIT"
	}
	return m, nil
}

func (m Model) toggleCurrentDir() (tea.Model, tea.Cmd) {
	if m.cursorIdx < 0 || m.cursorIdx >= len(m.flatList) {
		return m, nil
	}
	node := m.flatList[m.cursorIdx]
	if node.IsDir {
		m.treeRender.ToggleDir(node.Path)
		m.flatList = m.workspace.FlatList(m.treeRender.CollapsedDirs)
		if m.cursorIdx >= len(m.flatList) {
			m.cursorIdx = len(m.flatList) - 1
		}
	}
	return m, nil
}

func (m Model) selectCurrent() (tea.Model, tea.Cmd) {
	if m.cursorIdx < 0 || m.cursorIdx >= len(m.flatList) {
		return m, nil
	}
	node := m.flatList[m.cursorIdx]
	if node.IsDir {
		m.treeRender.ToggleDir(node.Path)
		m.flatList = m.workspace.FlatList(m.treeRender.CollapsedDirs)
		if m.cursorIdx >= len(m.flatList) {
			m.cursorIdx = len(m.flatList) - 1
		}
		return m, nil
	}
	buf, err := editor.LoadFromFile(node.Path)
	if err == nil {
		m.buffer = buf
		m.recent.Add(node.Path)
		m.currentFile = node.Path
		m.mode = "EDIT"
		if m.tabs != nil {
			m.tabs.Open(node.Path, buf)
		}
	}
	return m, nil
}

// autoOpenAtCursor öffnet automatisch die Datei unter dem Cursor im Tree
// wenn es ein File ist (kein Directory). Skip bei Buffern, die gerade editiert werden.
func (m Model) autoOpenAtCursor() {
	if m.cursorIdx < 0 || m.cursorIdx >= len(m.flatList) {
		return
	}
	node := m.flatList[m.cursorIdx]
	if node.IsDir {
		return
	}
	// Vermeide Reload wenn schon offen (User tippt noch)
	if m.buffer != nil && m.currentFile == node.Path {
		return
	}
	buf, err := editor.LoadFromFile(node.Path)
	if err == nil {
		m.buffer = buf
		m.recent.Add(node.Path)
		m.currentFile = node.Path
		m.mode = "EDIT"
		if m.tabs != nil {
			m.tabs.Open(node.Path, buf)
		}
	}
}

func (m Model) View() string {
	m.clampScrollOffset()

	if m.quitting {
		return "mdskim2 — beendet.\n"
	}
	if m.width == 0 || m.height == 0 {
		return fmt.Sprintf("mdskim2 — Workspace: %s — R5 MD-Highlight\n", m.workspacePath)
	}

	header := m.layout.Header(m.workspacePath, m.theme)
	sidebar := m.layout.Sidebar(m.renderSidebar(), m.theme, m.focus == "tree")
	editorContent := m.renderEditor()
	if m.paletteActive && m.palette != nil {
		matches := m.palette.Search(m.paletteQuery)
		var matchLines []string
		matchLines = append(matchLines, fmt.Sprintf("CMD: %s", m.paletteQuery))
		for i, name := range matches {
			if i >= 5 {
				break
			}
			desc, _ := m.palette.Run(name)
			matchLines = append(matchLines, fmt.Sprintf("  > %s — %s", name, desc))
		}
		editorContent = strings.Join(matchLines, "\n") + "\n" + editorContent
	}
	if m.searchActive {
		count := len(m.searchResults)
		barText := fmt.Sprintf("SUCHE: %s [Enter: Jump, Esc: Close, Ctrl+H: Replace, %d Treffer]", m.searchQuery, count)
		editorContent = barText + "\n" + editorContent
	}
	editor := m.layout.Editor(editorContent, m.theme, m.focus == "editor" || m.focus == "preview")

	// Right pane tabs: Preview | TOC | Backlinks (U6 — Python-mdskim look)
	tabLabels := []string{"Preview", "TOC", "Backlinks", "Tags"}
	tabContents := []string{
		m.renderPreviewTab(),
		m.renderTOC(),
		m.renderBacklinksTab(),
	}
	right := m.layout.RightPane(tabLabels, m.rightTab, tabContents, m.theme,
		m.focus == "toc" || m.focus == "preview")

	body := m.layout.Compose(sidebar, editor, right)

	mode := m.mode
	if m.saveError != "" {
		mode = "ERR: " + m.saveError
	}

	statusInfo := ui.StatusInfo{
		Workspace: m.workspacePath,
		File:      m.currentFile,
		Encoding:  "UTF-8",
		Mode:      mode,
		Version:   m.version,
	}
	if m.buffer != nil {
		statusInfo.Lines = m.buffer.TotalLines()
		statusInfo.Words = m.buffer.TotalWords()
		statusInfo.Modified = m.buffer.Modified
	}
	status := m.layout.RenderStatus(statusInfo, m.theme)
	footer := m.layout.RenderFooter(ui.DefaultShortcuts(), m.theme)

	// U6: Toolbar (1-zeilige Shortcut-Leiste unter dem Header)
	toolbar := m.layout.Toolbar("mdskim", m.toolbarShortcuts(), m.theme)

	parts := []string{header, toolbar, body, status}
	// U8: Prompt-Bar zwischen Status und Footer
	if promptBar := m.renderPrompt(m.width); promptBar != "" {
		parts = append(parts, promptBar)
	}
	parts = append(parts, footer)
	result := strings.Join(parts, "\n")

	// U9.2: Find-in-Files Modal-Overlay
	if m.findInFilesMode {
		result = m.renderFindInFilesOverlay(result)
	}

	// U9.3: Recent-Files Modal-Overlay
	if m.recentMenuMode {
		result = m.renderRecentMenuOverlay(result)
	}

	return result
}

// renderRecentMenuOverlay rendert das Recent-Files-Menü über dem Hauptview.
func (m Model) renderRecentMenuOverlay(base string) string {
	files := m.recentFilesList()
	var lines []string
	lines = append(lines, "\x1b[1;37m\x1b[48;5;63m  Recent Files  \x1b[0m")
	if len(files) == 0 {
		lines = append(lines, "\x1b[2m  No recent files yet\x1b[0m")
	} else {
		max := 10
		if len(files) < max {
			max = len(files)
		}
		start := 0
		if m.recentMenuIdx >= max {
			start = m.recentMenuIdx - max + 1
		}
		for i := start; i < start+max; i++ {
			f := files[i]
			marker := "  "
			if i == m.recentMenuIdx {
				marker = "\x1b[48;5;220m▶ \x1b[0m"
			}
			when := time.Unix(f.When, 0).Format("15:04:05")
			name := f.Display
			if name == "" {
				name = filepath.Base(f.Path)
			}
			if len(name) > 40 {
				name = name[:40] + "…"
			}
			lines = append(lines, fmt.Sprintf("%s\x1b[36m%s\x1b[0m  \x1b[33m%s\x1b[0m", marker, name, when))
		}
	}
	lines = append(lines, "\x1b[2m  ↑↓ Navigate · Enter Open · Esc Close\x1b[0m")
	modal := strings.Join(lines, "\n")

	baseLines := strings.Split(base, "\n")
	modalLines := strings.Split(modal, "\n")
	start := (len(baseLines) - len(modalLines)) / 2
	if start < 0 {
		start = 0
	}
	for i, ml := range modalLines {
		if start+i < len(baseLines) {
			padded := ml + strings.Repeat(" ", 80-len(stripANSI(ml)))
			baseLines[start+i] = "\x1b[K" + padded
		}
	}
	return strings.Join(baseLines, "\n")
}

// renderFindInFilesOverlay rendert ein Modal über dem Hauptview mit der aktuellen
// Query und den Resultaten (Treffer-Liste).
func (m Model) renderFindInFilesOverlay(base string) string {
	var lines []string
	lines = append(lines, "\x1b[1;37m\x1b[48;5;63m  Find in Files: "+m.findInFilesQuery+"█  \x1b[0m")
	if m.findInFilesQuery == "" {
		lines = append(lines, "\x1b[2m  Type to search...  (Esc to cancel)\x1b[0m")
	} else if len(m.findInFilesResults) == 0 {
		lines = append(lines, "\x1b[2m  No matches\x1b[0m")
	} else {
		lines = append(lines, "\x1b[2m  "+itoa(len(m.findInFilesResults))+" matches:\x1b[0m")
		max := 8
		if len(m.findInFilesResults) < max {
			max = len(m.findInFilesResults)
		}
		start := 0
		if m.findInFilesIdx >= max {
			start = m.findInFilesIdx - max + 1
		}
		for i := start; i < start+max; i++ {
			hit := m.findInFilesResults[i]
			marker := "  "
			if i == m.findInFilesIdx {
				marker = "\x1b[48;5;220m▶ \x1b[0m"
			}
			rel, _ := filepath.Rel(m.workspace.RootPath, hit.Path)
			if rel == "" {
				rel = hit.Path
			}
			match := strings.TrimSpace(hit.Match)
			if len(match) > 50 {
				match = match[:50] + "…"
			}
			lines = append(lines, fmt.Sprintf("%s\x1b[36m%s\x1b[0m:\x1b[33m%d\x1b[0m: %s", marker, rel, hit.Line+1, match))
		}
	}
	modal := strings.Join(lines, "\n")
	// Replace base's middle section with modal
	baseLines := strings.Split(base, "\n")
	modalLines := strings.Split(modal, "\n")
	start := (len(baseLines) - len(modalLines)) / 2
	if start < 0 {
		start = 0
	}
	for i, ml := range modalLines {
		if start+i < len(baseLines) {
			// Pad to 80 chars
			padded := ml + strings.Repeat(" ", maxLen([]string{baseLines[start+i]}, 80)-len(stripANSI(ml)))
			baseLines[start+i] = "\x1b[K" + padded
		}
	}
	return strings.Join(baseLines, "\n")
}

// itoa ist ein Mini-Wrapper (fmt.Sprintf("%d", ...) ist zu teuer).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}

// maxLen returns the max of len(stripANSI(s)) over all strings in arr.
func maxLen(arr []string, fallback int) int {
	m := fallback
	for _, s := range arr {
		l := len(stripANSI(s))
		if l > m {
			m = l
		}
	}
	return m
}

// stripANSI entfernt ANSI-Escape-Sequenzen aus s (für Längen-Berechnung).
func stripANSI(s string) string {
	var out []byte
	inEsc := false
	for i := 0; i < len(s); i++ {
		if inEsc {
			if s[i] >= 0x40 && s[i] <= 0x7e {
				inEsc = false
			}
			continue
		}
		if s[i] == 0x1b {
			inEsc = true
			continue
		}
		out = append(out, s[i])
	}
	return string(out)
}

func (m Model) renderSidebar() string {
	if m.workspace == nil {
		return "FILES\n(kein Workspace)"
	}
	headerMark := "\x1b[2m FILES \x1b[0m"
	if m.focus == "tree" {
		headerMark = "\x1b[48;5;63m\x1b[1;37m FILES \x1b[0m"
	}
	// Truncate filenames to avoid lipgloss line-wrapping within the sidebar pane.
	maxName := m.layout.SidebarWidth - 10
	if maxName < 8 {
		maxName = 8
	}
	truncate := func(name string) string {
		runes := []rune(name)
		if len(runes) > maxName {
			return string(runes[:maxName-1]) + "…"
		}
		return name
	}
	var lines []string
	lines = append(lines, headerMark)
	// U8.3: Tree-Filter-Input-Zeile (wenn aktiv)
	if m.treeFiltered {
		filterLine := "\x1b[48;5;220m\x1b[30m /" + m.treeFilter + "_ \x1b[0m"
		if len(m.treeFilter) == 0 {
			filterLine = "\x1b[48;5;220m\x1b[30m /_ \x1b[0m"
		}
		lines = append(lines, filterLine)
	}
	for i, node := range m.flatList {
		cn := *node
		cn.Name = truncate(node.Name)
		rendered := m.treeRender.FormatNode(&cn)
		var line string
		if i == m.cursorIdx {
			line = "\x1b[1;33m\x1b[0m " + strings.TrimPrefix(rendered, " ")
			if m.focus == "tree" {
				// aktive Zeile zusätzlich hervorheben
				line = "\x1b[7m" + strings.TrimPrefix(rendered, " ") + "\x1b[0m"
				line = "\x1b[1;33m\x1b[0m " + line
			}
		} else {
			line = "  " + rendered
		}
		lines = append(lines, line)
	}
	if m.workspace.TotalFiles() == 0 {
		lines = append(lines, "(leer)")
	}
	footerHint := "\n\x1b[2m ↑↓ Navig · Enter Tree→Edit \x1b[0m"
	return strings.Join(lines, "\n") + footerHint
}

func (m Model) renderEditor() string {
	if m.buffer == nil {
		return "\x1b[1;36m[NO-BUFFER]\x1b[0m \x1b[1;37m# Willkommen bei mdskim2\x1b[0m\n\n" +
			"\x1b[33m↑↓\x1b[0m navigiert Tree und lädt Datei automatisch\n" +
			"\x1b[33mEnter\x1b[0m wechselt in den Editor\n" +
			"\x1b[33mEsc\x1b[0m zurück zum Tree\n" +
			"\x1b[33mCtrl+S\x1b[0m Speichern, \x1b[33mCtrl+Q\x1b[0m Beenden\n\n" +
			"Workspace: \x1b[1m" + m.workspacePath + "\x1b[0m"
	}

	// Header-Bar mit Datei + Modifiziert-Flag + Focus
	dirtyMark := "  "
	if m.buffer.Modified {
		dirtyMark = "\x1b[1;31m●\x1b[0m"
	}
	focusMark := "\x1b[48;5;240m\x1b[37m TREE \x1b[0m"
	if m.focus == "editor" {
		focusMark = "\x1b[48;5;63m\x1b[1;37m EDIT \x1b[0m"
	}
	wrapMark := ""
	if m.lineWrap {
		wrapMark = " \x1b[2;37m[WRAP]\x1b[0m"
	}
	readOnlyMark := "  "
	if m.readOnly {
		readOnlyMark = " \x1b[48;5;130m\x1b[1;37m[RO]\x1b[0m"
	}
	autoSavedMark := ""
	if m.autoSavedAt != "" {
		autoSavedMark = fmt.Sprintf(" \x1b[48;5;236m\x1b[1;32m[AUTO-SAVED %s]\x1b[0m", m.autoSavedAt)
	}
	headerLine := focusMark + " \x1b[1;37m" + trimPath(m.currentFile, 50) + "\x1b[0m " + dirtyMark + readOnlyMark + wrapMark + autoSavedMark

	// Selection-Range
	srSel, scSel, erSel, ecSel, hasSel := m.buffer.SelectionRange()

	// Line numbers + body mit Block-Cursor + Selection-Highlight
	totalLines := len(m.buffer.Lines)
	lineNumWidth := len(fmt.Sprintf("%d", totalLines))
	if lineNumWidth < 2 {
		lineNumWidth = 2
	}

	// Viewport-Scrolling: zeige Buffer ab scrollOffset
	viewportHeight := m.editorViewportHeight()
	visibleStart := m.editorScrollOffset
	visibleEnd := visibleStart + viewportHeight
	if visibleEnd > totalLines {
		visibleEnd = totalLines
	}
	if visibleStart >= visibleEnd {
		visibleStart = 0
		visibleEnd = totalLines
		if totalLines > viewportHeight {
			visibleEnd = viewportHeight
		}
	}

	// Bracket-Match: find matching char if cursor on bracket
	var brRow, brCol int
	var hasBracket bool
	if m.focus == "editor" && m.buffer != nil {
		brRow, brCol, hasBracket = findMatchingBracket(m.buffer.Lines, m.buffer.CursorRow, m.buffer.CursorCol)
		_ = brRow
		_ = brCol
	}

	var bodyLines []string
	for i := visibleStart; i < visibleEnd; i++ {
		line := m.buffer.Lines[i]
		// Gutter: line number
		var numFmt string
		if i == m.buffer.CursorRow {
			numFmt = fmt.Sprintf("\x1b[1;33m%*d\x1b[0m", lineNumWidth, i+1)
		} else {
			numFmt = fmt.Sprintf("\x1b[2;37m%*d\x1b[0m", lineNumWidth, i+1)
		}

		// Build content with highlighting
		displayLine := line
		if m.whitespaceMark {
			displayLine = renderWhitespace(line)
		}
		p := markdown.HighlightParams{CurrentLine: m.buffer.CursorRow}
		rendered := markdown.HighlightLineWith(displayLine, p, i)

		// Selection-Markierung: reverse-video Background auf markiertem Bereich
		if hasSel && i >= srSel && i <= erSel {
			rendered = applySelectionBG(rendered, line, i, srSel, scSel, erSel, ecSel)
		}

		// Block-Cursor an CursorCol
		if i == m.buffer.CursorRow && m.focus == "editor" {
			rendered = injectBlockCursor(rendered, line, m.buffer.CursorCol)
		}

		// Bracket-Match-Highlight auf der matchenden Zeile
		if hasBracket && i == brRow {
			runes := []rune(line)
			if brCol >= 0 && brCol < len(runes) {
				rendered = injectCharHighlight(rendered, line, brCol, "\x1b[48;5;220m\x1b[30m", "\x1b[0m")
			}
		}

		bodyLines = append(bodyLines, numFmt+" │ "+rendered)
	}
	body := strings.Join(bodyLines, "\n")

	// Scroll-Indicator rechts (1 Zeichen pro Block)
	var scrollArrow string
	if visibleStart > 0 {
		scrollArrow += "\x1b[33m↑\x1b[0m"
	}
	if visibleEnd < totalLines {
		scrollArrow += "\x1b[33m↓\x1b[0m"
	}
	scrollInfo := ""
	if scrollArrow != "" {
		scrollInfo = " " + scrollArrow + fmt.Sprintf(" (%d-%d/%d)", visibleStart+1, visibleEnd, totalLines)
	}

	// Status-Bar mit allen Editor-Infos
	var currentLineLen int
	if m.buffer.CursorRow < len(m.buffer.Lines) {
		currentLineLen = utf8.RuneCountInString(m.buffer.Lines[m.buffer.CursorRow])
	}
	wc := m.buffer.WordCount()
	cc := m.buffer.CharCount()
	cursorPos := fmt.Sprintf("\x1b[48;5;236m \x1b[1;33m Ln %d/%d \x1b[0m\x1b[48;5;236m \x1b[33m Col %d/%d \x1b[0m",
		m.buffer.CursorRow+1, totalLines, m.buffer.CursorCol+1, currentLineLen)
	counters := fmt.Sprintf("\x1b[2;37m Words %d · Chars %d \x1b[0m", wc, cc)
	// Line-Ending-Indicator (U4)
	lineEndMark := "LF"
	if m.buffer.LineEnding == "\r\n" {
		lineEndMark = "CRLF"
	} else if m.buffer.LineEnding == "\r" {
		lineEndMark = "CR"
	}
	lineEndingStat := fmt.Sprintf(" \x1b[2;37mEOL %s \x1b[0m", lineEndMark)
	hints := "\x1b[2;37m [Shift+Arrows] Sel · [Ctrl+D] Dup · [Ctrl+G] Go-to · [Ctrl+W] Wrap \x1b[0m"

	// Optional: Go-to-line Modal-Bar
	gotoBar := ""
	if m.goToLineMode {
		gotoBar = "\n\x1b[48;5;240m\x1b[1;37m Go to line: \x1b[0m\x1b[1;33m" + m.goToLineQuery + "_\x1b[0m"
	}

	return headerLine + "\n" + body + "\n" + cursorPos + " " + counters + lineEndingStat + " " + scrollInfo + "\n" + hints + gotoBar
}

// editorViewportHeight returns the available height for the editor content.
func (m Model) editorViewportHeight() int {
	return m.height - 8
}

// editorScrollOffset tracks the first visible line.
func (m *Model) clampScrollOffset() {
	if m.buffer == nil {
		return
	}
	vp := m.editorViewportHeight()
	if m.editorScrollOffset < 0 {
		m.editorScrollOffset = 0
	}
	maxOff := len(m.buffer.Lines) - vp
	if maxOff < 0 {
		maxOff = 0
	}
	if m.editorScrollOffset > maxOff {
		m.editorScrollOffset = maxOff
	}
	if m.buffer.CursorRow < m.editorScrollOffset {
		m.editorScrollOffset = m.buffer.CursorRow
	}
	if m.buffer.CursorRow >= m.editorScrollOffset+vp {
		m.editorScrollOffset = m.buffer.CursorRow - vp + 1
	}
}

// applySelectionBG wandelt den gerenderten String so um, dass der markierte Bereich in Selection-Farbe erscheint.
func applySelectionBG(rendered string, originalLine string, lineIdx, sr, sc, er, ec int) string {
	startCol := 0
	endCol := len(originalLine)
	if lineIdx == sr {
		startCol = sc
	}
	if lineIdx == er {
		endCol = ec
	}
	if startCol >= endCol {
		return rendered
	}
	// Wrap the substring in selection highlight
	before := string([]rune(originalLine)[:startCol])
	selText := string([]rune(originalLine)[startCol:endCol])
	after := string([]rune(originalLine)[endCol:])
	// Note: für selected text rendern wir plain — highlighting wird durch selection BG ersetzt
	return before + "\x1b[48;5;57m" + selText + "\x1b[0m" + after
}

// injectBlockCursor fügt einen Block-Cursor (reverse-video) an col in line ein.
func injectBlockCursor(rendered string, originalLine string, col int) string {
	runes := []rune(originalLine)
	if col > len(runes) {
		col = len(runes)
	}
	before := string(runes[:col])
	if col < len(runes) {
		char := runes[col]
		after := string(runes[col+1:])
		return before + "\x1b[7m" + string(char) + "\x1b[0m" + after
	}
	return rendered + "\x1b[7m \x1b[0m"
}

func (m Model) tocHeadings() []markdown.Heading {
	if m.buffer == nil {
		return nil
	}
	return markdown.Headings(m.buffer.ToString())
}

func (m Model) renderPreviewTab() string {
	if m.buffer == nil {
		return "(keine Datei geöffnet)"
	}
	// Render markdown with glamour
	rendered, err := preview.Render(m.buffer.ToString())
	if err != nil {
		return fmt.Sprintf("(Render-Fehler: %v)", err)
	}
	return rendered
}

func (m Model) renderBacklinksTab() string {
	if m.buffer == nil || m.buffer.Path == "" {
		return "(keine Datei geöffnet)"
	}
	m.refreshBacklinks()
	if len(m.backlinksCache) == 0 {
		return "Keine Backlinks für diese Datei.\nTipp: erstelle [[Wiki-Links]] in anderen Dateien."
	}
	var lines []string
	lines = append(lines, fmt.Sprintf("%d Backlink(s):", len(m.backlinksCache)))
	for _, e := range m.backlinksCache {
		label := e.LinkTarget
		if e.LinkAlias != "" {
			label = e.LinkAlias + " → " + label
		}
		base := filepath.Base(e.SourceFile)
		lines = append(lines, fmt.Sprintf("  • %s:%d  %s", base, e.SourceLine, label))
	}
	return strings.Join(lines, "\n")
}

func (m Model) refreshBacklinks() {
	if m.buffer == nil || m.buffer.Path == "" {
		return
	}
	if m.backlinksFor == m.buffer.Path {
		return
	}
	ws := m.workspacePath
	if m.workspace != nil && m.workspace.RootPath != "" {
		ws = m.workspace.RootPath
	}
	entries, err := backlinks.Collect(ws, m.buffer.Path)
	if err == nil {
		m.backlinksCache = entries
		m.backlinksFor = m.buffer.Path
	}
}

func (m Model) toolbarShortcuts() []ui.Shortcut {
	return []ui.Shortcut{
		{Key: "Ctrl+O", Description: "Open"},
		{Key: "Ctrl+S", Description: "Save"},
		{Key: "Ctrl+P", Description: "Preview"},
		{Key: "Ctrl+K", Description: "Palette"},
		{Key: "Ctrl+Shift+←→", Description: "Panes"},
		{Key: "Ctrl+Q", Description: "Quit"},
	}
}

func (m Model) renderTOC() string {
	hs := m.tocHeadings()
	if len(hs) == 0 {
		return "INHALT\n──────\n▸ (Datei öffnen)"
	}
	tocLines := []string{"INHALT", "──────"}
	for _, h := range hs {
		indent := strings.Repeat("  ", h.Level-1)
		bullet := "▸"
		if h.Level == 1 {
			bullet = "▾"
		}
		tocLines = append(tocLines, fmt.Sprintf("%s%s %s", indent, bullet, h.Text))
	}
	return strings.Join(tocLines, "\n")
}

func trimPath(path string, max int) string {
	if len(path) <= max {
		return path
	}
	return "…" + path[len(path)-max+1:]
}

// Quitting returns true wenn Ctrl+Q gedrückt wurde.
func (m Model) Quitting() bool { return m.quitting }

// FlatList returns the FlatList for testing.
func (m Model) FlatList() []*workspace.FileNode { return m.flatList }

// WithCursor returns a copy with cursorIdx set.
func (m Model) WithCursor(idx int) Model {
	if idx >= 0 {
		m.cursorIdx = idx
	}
	return m
}

// CurrentFile returns the current file path.
func (m Model) CurrentFile() string { return m.currentFile }

// Cursor returns current cursor position.
func (m Model) Cursor() int { return m.cursorIdx }

func (m Model) Buffer() *editor.Buffer { return m.buffer }

// IsBufferOpen returns true if the buffer is loaded.
func (m Model) IsBufferOpen() bool { return m.buffer != nil }

// HighlightedLines returns the buffer content with ANSI-highlight applied per line.
// Public for testing — bypasses lipgloss width constraint.
func (m Model) HighlightedLines() []string {
	if m.buffer == nil {
		return nil
	}
	var hl []string
	p := markdown.HighlightParams{CurrentLine: m.buffer.CursorRow}
	for _, line := range m.buffer.Lines {
		hl = append(hl, markdown.HighlightLine(line, p))
	}
	return hl
}

// TocText returns the TOC pane content (hierarchical headings).
func (m Model) TocText() string {
	hs := m.tocHeadings()
	if len(hs) == 0 {
		return "INHALT\n──────\n▸ (Datei öffnen)"
	}
	tocLines := []string{"INHALT", "──────"}
	for _, h := range hs {
		indent := strings.Repeat("  ", h.Level-1)
		bullet := "▸"
		if h.Level == 1 {
			bullet = "▾"
		}
		tocLines = append(tocLines, fmt.Sprintf("%s%s %s", indent, bullet, h.Text))
	}
	return strings.Join(tocLines, "\n")
}

// HeadingsCount returns number of headings in current buffer.
func (m Model) HeadingsCount() int {
	return len(m.tocHeadings())
}

// findMatchingBracket returns the (row, col) of the matching bracket for the char at (row, col).
func findMatchingBracket(lines []string, row, col int) (mr int, mc int, ok bool) {
	if row < 0 || row >= len(lines) {
		return 0, 0, false
	}
	line := []rune(lines[row])
	if col < 0 || col >= len(line) {
		return 0, 0, false
	}
	ch := line[col]
	openBrackets := map[rune]rune{'(': ')', '[': ']', '{': '}', '<': '>'}
	closeBrackets := map[rune]rune{')': '(', ']': '[', '}': '{', '>': '<'}
	if open, isOpen := openBrackets[ch]; isOpen {
		// Search forward
		depth := 1
		r := row
		c := col + 1
		for r < len(lines) {
			lr := []rune(lines[r])
			if c >= len(lr) {
				r++
				c = 0
				continue
			}
			if lr[c] == open {
				depth++
			} else if lr[c] == ch {
				// Different opening-style delimiter; skip
			} else if lr[c] == ')' || lr[c] == ']' || lr[c] == '}' || lr[c] == '>' {
				if closeBrackets[lr[c]] != ch {
					// different bracket class, skip
				} else {
					depth--
					if depth == 0 {
						return r, c, true
					}
				}
			}
			c++
		}
		return 0, 0, false
	}
	if _, isClose := closeBrackets[ch]; isClose {
		// Search backward
		depth := 1
		r := row
		c := col - 1
		for r >= 0 {
			lr := []rune(lines[r])
			for c >= 0 {
				switch lr[c] {
				case ')', ']', '}', '>':
					depth++
				case '(', '[', '{', '<':
					depth--
					if depth == 0 {
						return r, c, true
					}
				}
				c--
			}
			r--
			if r >= 0 {
				c = len([]rune(lines[r])) - 1
			}
		}
		return 0, 0, false
	}
	return 0, 0, false
}

// injectCharHighlight wraps a single character at col in line with prefix/suffix.
func injectCharHighlight(rendered string, originalLine string, col int, prefix string, suffix string) string {
	runes := []rune(originalLine)
	if col >= len(runes) {
		return rendered + prefix + " " + suffix
	}
	before := string(runes[:col])
	after := ""
	if col+1 < len(runes) {
		after = string(runes[col+1:])
	}
	return before + prefix + string(runes[col]) + suffix + after
}

// renderWhitespace rewrites trailing spaces as · and tabs as → in a string.
// keeps internal whitespace clean.
func renderWhitespace(s string) string {
	runes := []rune(s)
	// Find last non-whitespace index
	lastNonWS := len(runes) - 1
	for lastNonWS >= 0 && (runes[lastNonWS] == ' ' || runes[lastNonWS] == '\t') {
		lastNonWS--
	}
	var out []byte
	for i, ch := range runes {
		switch {
		case ch == '\t':
			out = append(out, []byte("\x1b[2;37m→\x1b[0m")...)
		case ch == ' ' && i > lastNonWS:
			out = append(out, []byte("\x1b[2;37m·\x1b[0m")...)
		default:
			out = append(out, []byte(string(ch))...)
		}
	}
	return string(out)
}

// visiblePanes returns the ordered list of currently visible pane names.
func (m Model) visiblePanes() []string {
	out := []string{"tree"}
	if m.buffer != nil {
		out = append(out, "editor")
	}
	if m.buffer != nil {
		out = append(out, "toc")
	}
	return out
}

// cycleFocus advances focus by direction (1 forward, -1 backward) through visible panes.
func (m Model) cycleFocus(direction int) string {
	panes := m.visiblePanes()
	if len(panes) == 0 {
		return "tree"
	}
	cur := 0
	for i, p := range panes {
		if p == m.focus {
			cur = i
			break
		}
	}
	next := cur + direction
	if next < 0 {
		next = len(panes) - 1
	}
	if next >= len(panes) {
		next = 0
	}
	return panes[next]
}

// setFocus updates m.focus and returns updated Model. Out-of-focus values map to default.
func (m Model) setFocus(f string) Model {
	switch f {
	case "tree", "editor", "preview", "toc":
		m.focus = f
	}
	if m.focus == "preview" && m.buffer == nil {
		m.focus = "tree"
	}
	return m
}

// RightTab returns the current right-pane tab index.
func (m Model) RightTab() int {
	return m.rightTab
}

// SetRightTab sets the right-pane tab index (0=Preview, 1=TOC, 2=Backlinks, 3=Tags).
func (m *Model) SetRightTab(idx int) {
	if idx < 0 {
		idx = 0
	}
	if idx > 3 {
		idx = 3
	}
	m.rightTab = idx
	m.focus = "toc"
}

// LoadFileForTest loads a file into the buffer (used by tests).
func (m *Model) LoadFileForTest(path string) {
	buf, err := editor.LoadFromFile(path)
	if err != nil {
		return
	}
	m.buffer = buf
	m.currentFile = path
	m.recent.Add(path)
	m.backlinksFor = ""
	m.refreshBacklinks()
}

func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Type != tea.MouseLeft {
		return m, nil
	}
	if m.width <= 0 || m.height <= 0 {
		return m, nil
	}

	// Body starts at Y = HeaderHeight + ToolbarHeight (= 2 default)
	bodyY := m.layout.HeaderHeight + m.layout.ToolbarHeight
	if msg.Y < bodyY {
		return m, nil
	}
	bodyBottom := m.height - m.layout.StatusHeight - m.layout.FooterHeight
	if msg.Y >= bodyBottom {
		return m, nil
	}

	// Pane boundaries (X is in absolute screen coords)
	sidebarEnd := m.layout.SidebarWidth
	editorEnd := sidebarEnd + m.layout.EditorWidth
	rightEnd := editorEnd + m.layout.RightWidth

	// Inner row offset = relative to body
	innerY := msg.Y - bodyY

	switch {
	case msg.X < sidebarEnd:
		return m.handleSidebarClick(msg.X, innerY)
	case msg.X < editorEnd:
		return m.handleEditorClick(msg.X-sidebarEnd, innerY)
	case msg.X < rightEnd:
		return m.handleRightPaneClick(msg.X-editorEnd, innerY)
	}
	return m, nil
}

func (m Model) handleSidebarClick(innerX, innerY int) (tea.Model, tea.Cmd) {
	// Sidebar has no top border; row 0 = "FILES" label, rows 1..N = tree entries.
	treeIdx := innerY - 1
	if treeIdx < 0 || treeIdx >= len(m.flatList) {
		return m, nil
	}
	m.cursorIdx = treeIdx
	m.focus = "tree"
	return m.selectCurrent()
}

func (m Model) handleEditorClick(innerX, innerY int) (tea.Model, tea.Cmd) {
	if m.buffer == nil {
		m.focus = "editor"
		return m, nil
	}
	// Editor: row 0 = pane title (e.g. "TREE filename.md"), rows 1+ = lines
	lineRow := innerY - 1 + m.editorScrollOffset
	if lineRow < 0 {
		lineRow = 0
	}
	totalLines := m.buffer.TotalLines()
	if totalLines == 0 {
		return m, nil
	}
	if lineRow >= totalLines {
		lineRow = totalLines - 1
	}
	const lineNumberGutter = 6
	editorCol := innerX - lineNumberGutter
	if editorCol < 0 {
		editorCol = 0
	}
	line := ""
	if lineRow < len(m.buffer.Lines) {
		line = m.buffer.Lines[lineRow]
	}
	if editorCol > len(line) {
		editorCol = len(line)
	}
	m.buffer.CursorRow = lineRow
	m.buffer.CursorCol = editorCol
	m.buffer.SelAnchorRow = -1
	m.buffer.SelAnchorCol = 0
	m.focus = "editor"
	return m, nil
}

func (m Model) handleRightPaneClick(innerX, innerY int) (tea.Model, tea.Cmd) {
	// Right pane: row 0 = tab-bar, rows 1..N = content (no top border)
	if innerY == 0 {
		// Tab bar click — calculate which tab by X position
		totalInnerWidth := m.layout.RightWidth
		if totalInnerWidth < 1 {
			totalInnerWidth = 1
		}
		tabWidth := totalInnerWidth / 3
		if tabWidth < 1 {
			tabWidth = 1
		}
		idx := innerX / tabWidth
		if idx > 2 {
			idx = 2
		}
		m.rightTab = idx
		m.focus = "toc"
		return m, nil
	}

	contentRow := innerY - 1
	if contentRow < 0 {
		return m, nil
	}
	m.focus = "toc"
	switch m.rightTab {
	case 0:
		// Preview tab — read-only content. Click sets focus but no action.
		return m, nil
	case 1:
		// TOC tab — click on heading jumps editor
		// renderTOC: row 0 = "INHALT", row 1 = "──────", row 2..N = headings
		hs := m.tocHeadings()
		treeIdx := contentRow - 2
		if treeIdx < 0 || treeIdx >= len(hs) {
			return m, nil
		}
		if m.buffer != nil {
			target := hs[treeIdx].Line - 1
			if target < 0 {
				target = 0
			}
			if target >= len(m.buffer.Lines) {
				target = len(m.buffer.Lines) - 1
			}
			m.buffer.CursorRow = target
			m.buffer.CursorCol = 0
			h := m.editorViewportHeight()
			m.editorScrollOffset = clampOffset(target-3, m.editorScrollOffset, h)
			m.focus = "editor"
		}
		return m, nil
	case 2:
		// Backlinks tab — click opens source file
		// renderBacklinksTab: row 0 = title, rows 1..N = entries (or "no links" message)
		// Simple approach: first entry is at contentRow, but renderBacklinksTab might have
		// different layout. We look up by entry and set cursorIdx.
		if contentRow >= len(m.backlinksCache) {
			return m, nil
		}
		entry := m.backlinksCache[contentRow]
		idx := m.findTreeIdxByPath(entry.SourceFile)
		if idx >= 0 {
			m.cursorIdx = idx
			return m.selectCurrent()
		}
		return m, nil
	case 3:
		// Tags tab — click on tag jumps editor to first occurrence
		// renderTagsTab: row 0 = title, row 1 = separator, rows 2..N = tag entries
		if m.buffer == nil {
			return m, nil
		}
		tags := markdown.Tags(m.buffer.ToString())
		counts := map[string]int{}
		for _, t := range tags {
			counts[t.Name]++
		}
		type kv struct {
			name string
			line int
		}
		var kvs []kv
		seen := map[string]bool{}
		for _, t := range tags {
			if seen[t.Name] {
				continue
			}
			seen[t.Name] = true
			kvs = append(kvs, kv{t.Name, t.Line})
		}
		tagIdx := contentRow - 2
		if tagIdx < 0 || tagIdx >= len(kvs) {
			return m, nil
		}
		if kvs[tagIdx].line < len(m.buffer.Lines) {
			m.buffer.CursorRow = kvs[tagIdx].line
			m.buffer.CursorCol = 0
			h := m.editorViewportHeight()
			m.editorScrollOffset = clampOffset(kvs[tagIdx].line-3, m.editorScrollOffset, h)
			m.focus = "editor"
		}
		return m, nil
	}
	return m, nil
}

func (m Model) findTreeIdxByPath(path string) int {
	for i, node := range m.flatList {
		if node.Path == path {
			return i
		}
	}
	return -1
}

func clampOffset(target, current, viewportH int) int {
	if target < current {
		return target
	}
	if target > current+viewportH-3 {
		return target - viewportH + 3
	}
	return current
}

// applyTreeFilter filtert die flatList anhand von m.treeFilter.
// Wenn der Filter leer ist, wird die Original-flatList wiederhergestellt.
func (m *Model) applyTreeFilter() {
	if !m.treeFiltered || m.workspace == nil {
		return
	}
	if m.treeFilter == "" {
		m.flatList = m.workspace.FlatList(m.treeRender.CollapsedDirs)
		if m.cursorIdx >= len(m.flatList) {
			m.cursorIdx = len(m.flatList) - 1
		}
		if m.cursorIdx < 0 {
			m.cursorIdx = 0
		}
		return
	}
	orig := m.workspace.FlatList(m.treeRender.CollapsedDirs)
	filtered := orig[:0]
	for _, n := range orig {
		if n.IsDir {
			// Verzeichnisse immer anzeigen wenn darin ein Match wäre
			// Vereinfachung: bei leerer Filter-Eingabe alle anzeigen, sonst nur Name-Match
			if strings.Contains(strings.ToLower(n.Name), strings.ToLower(m.treeFilter)) {
				filtered = append(filtered, n)
			}
		} else {
			if strings.Contains(strings.ToLower(n.Name), strings.ToLower(m.treeFilter)) {
				filtered = append(filtered, n)
			}
		}
	}
	m.flatList = filtered
	if m.cursorIdx >= len(m.flatList) {
		m.cursorIdx = len(m.flatList) - 1
	}
	if m.cursorIdx < 0 {
		m.cursorIdx = 0
	}
}

// refreshWorkspace lädt den Tree neu aus dem Disk (für F5/Ctrl+R).
func (m *Model) refreshWorkspace() {
	if m.workspace == nil {
		return
	}
	ws, err := workspace.Load(m.workspace.RootPath)
	if err != nil {
		return
	}
	m.workspace = ws
	m.treeRender = workspace.NewTreeRenderer()
	m.flatList = m.workspace.FlatList(m.treeRender.CollapsedDirs)
	if m.cursorIdx >= len(m.flatList) {
		m.cursorIdx = len(m.flatList) - 1
	}
	if m.cursorIdx < 0 {
		m.cursorIdx = 0
	}
	m.refreshBacklinks()
}

// executePrompt führt die aktuelle Prompt-Aktion aus (Enter im Prompt).
func (m *Model) executePrompt() (tea.Model, tea.Cmd) {

	if m.cursorIdx < 0 || m.cursorIdx >= len(m.flatList) {
		m.promptMode = ""
		m.promptQuery = ""
		return *m, nil
	}
	node := m.flatList[m.cursorIdx]
	parent := node.Path
	if !node.IsDir {
		parent = filepath.Dir(node.Path)
	}
	switch m.promptMode {
	case "new-file":
		name := strings.TrimSpace(m.promptQuery)
		if name == "" {
			m.promptMode = ""
			m.promptQuery = ""
			return *m, nil
		}
		newPath := filepath.Join(parent, name)
		if err := os.WriteFile(newPath, []byte(""), 0644); err == nil {
			m.refreshWorkspace()
			m.cursorIdx = m.findTreeIdxByPath(newPath)
			if m.cursorIdx >= 0 {
				// Datei automatisch öffnen (frisch angelegte Datei direkt aktiv)
				updated, _ := m.selectCurrent()
				if um, ok := updated.(Model); ok {
					m.currentFile = um.currentFile
					m.buffer = um.buffer
					m.mode = um.mode
				}
			}
		}
		m.promptMode = ""
		m.promptQuery = ""
		return *m, nil
	case "new-folder":
		name := strings.TrimSpace(m.promptQuery)
		if name == "" {
			m.promptMode = ""
			m.promptQuery = ""
			return *m, nil
		}
		newPath := filepath.Join(parent, name)
		if err := os.MkdirAll(newPath, 0755); err == nil {
			m.refreshWorkspace()
			m.cursorIdx = m.findTreeIdxByPath(newPath)
		}
		m.promptMode = ""
		m.promptQuery = ""
		return *m, nil
	case "rename":
		name := strings.TrimSpace(m.promptQuery)
		if name == "" {
			m.promptMode = ""
			m.promptQuery = ""
			return *m, nil
		}
		newPath := filepath.Join(filepath.Dir(node.Path), name)
		if err := os.Rename(node.Path, newPath); err == nil {
			m.refreshWorkspace()
			m.cursorIdx = m.findTreeIdxByPath(newPath)
		}
		m.promptMode = ""
		m.promptQuery = ""
		return *m, nil
	case "confirm-delete":
		if strings.ToLower(strings.TrimSpace(m.promptQuery)) != "y" {
			m.promptMode = ""
			m.promptQuery = ""
			return *m, nil
		}
		if node.IsDir {
			if err := os.RemoveAll(node.Path); err == nil {
				m.refreshWorkspace()
				if m.cursorIdx >= len(m.flatList) {
					m.cursorIdx = len(m.flatList) - 1
				}
			}
		} else {
			if err := os.Remove(node.Path); err == nil {
				m.refreshWorkspace()
				if m.cursorIdx >= len(m.flatList) {
					m.cursorIdx = len(m.flatList) - 1
				}
			}
		}
		m.promptMode = ""
		m.promptQuery = ""
		return *m, nil
	}
	m.promptMode = ""
	m.promptQuery = ""
	return *m, nil
}

// renderPrompt rendert die Prompt-Bar am unteren Bildschirmrand.
func (m Model) renderPrompt(width int) string {
	if m.promptMode == "" {
		return ""
	}
	var label string
	switch m.promptMode {
	case "new-file":
		label = "Neue Datei:"
	case "new-folder":
		label = "Neuer Ordner:"
	case "rename":
		label = "Neuer Name:"
	case "confirm-delete":
		label = "Wirklich löschen? (y/n):"
	}
	bar := "\x1b[48;5;63m\x1b[1;37m " + label + " \x1b[0m"
	bar += " \x1b[7m" + m.promptQuery + " \x1b[0m"
	bar += " \x1b[2m(Esc=Abbruch)\x1b[0m"
	return bar
}

// FocusForTest returns the current focus (used by tests).
func (m Model) FocusForTest() string {
	return m.focus
}

// CurrentFileForTest returns the path of the currently loaded file (used by tests).
func (m Model) CurrentFileForTest() string {
	return m.currentFile
}

// CycleFocusForTest returns the next/prev focus name without applying it.
func (m Model) CycleFocusForTest(direction int) string {
	return m.cycleFocus(direction)
}

// SetFocusForTest applies a focus name (used by tests).
func (m *Model) SetFocusForTest(name string) {
	m2 := m.setFocus(name)
	m.focus = m2.focus
}

// FlatListForTest returns the current flatList (used by tests).
func (m Model) FlatListForTest() []string {
	out := make([]string, len(m.flatList))
	for i, n := range m.flatList {
		if n.IsDir {
			out[i] = "[DIR] " + n.Name
		} else {
			out[i] = "[FILE] " + n.Name
		}
	}
	return out
}

// BufferForTest returns the current buffer (read-only, used by tests).
func (m Model) BufferForTest() *editor.Buffer {
	return m.buffer
}

// WidthForTest returns width (used by tests).
func (m Model) WidthForTest() int { return m.width }

// HeightForTest returns height (used by tests).
func (m Model) HeightForTest() int { return m.height }

// PromptModeForTest returns the current prompt mode (used by tests).
func (m Model) PromptModeForTest() string { return m.promptMode }

// PromptQueryForTest returns the current prompt query (used by tests).
func (m Model) PromptQueryForTest() string { return m.promptQuery }

// WorkspaceRootForTest returns the workspace root path (used by tests).
func (m Model) WorkspaceRootForTest() string {
	if m.workspace != nil {
		return m.workspace.RootPath
	}
	return ""
}

// TreeFilteredForTest returns whether tree filter mode is active (used by tests).
func (m Model) TreeFilteredForTest() bool { return m.treeFiltered }

// TreeFilterForTest returns the current tree filter query (used by tests).
func (m Model) TreeFilterForTest() string { return m.treeFilter }

// maybeTriggerWikiLink öffnet das Wiki-Link-Popup, wenn der User [[ tippt,
// und filtert es live weiter, sobald er den Namen weiter schreibt.
func (m *Model) maybeTriggerWikiLink(r rune) {
	if m.buffer == nil {
		return
	}
	if r == '[' {
		line := m.buffer.Lines[m.buffer.CursorRow]
		runes := []rune(line)
		// Der gerade eingefügte '[' steht an CursorCol-1.
		// Wir prüfen das Zeichen davor — bei "[[" steht dort die vorherige '['.
		newBrackCol := m.buffer.CursorCol - 1
		prevCol := newBrackCol - 1
		if prevCol < 0 || prevCol >= len(runes) {
			return
		}
		if runes[prevCol] != '[' {
			return
		}
		// Doppelte [ erkannt → Popup öffnen
		m.wikiLinkPopup = true
		m.wikiLinkQuery = ""
		m.refreshWikiLinkMatches()
		return
	}
	// Andere Zeichen: Popup ggf. weiter nach rechts filtern oder schließen
	if !m.wikiLinkPopup {
		return
	}
	if r == ']' {
		// Akzeptiert: User hat den Link selbst geschrieben → Popup schließen
		m.wikiLinkPopup = false
		m.wikiLinkQuery = ""
		return
	}
	// Beim Tippen Filter-Query verlängern
	m.wikiLinkQuery += string(r)
	m.refreshWikiLinkMatches()
}

// refreshWikiLinkMatches holt die zur Query passenden Markdown-Dateien aus
// dem Workspace und legt sie als Popup-Items ab.
func (m *Model) refreshWikiLinkMatches() {
	m.wikiLinkMatches = nil
	if m.workspace == nil {
		return
	}
	all := m.workspace.FlatList(m.treeRender.CollapsedDirs)
	for _, n := range all {
		if n.IsDir {
			continue
		}
		// Nur Markdown-Dateien
		if !strings.HasSuffix(n.Name, ".md") && !strings.HasSuffix(n.Name, ".markdown") {
			continue
		}
		if m.wikiLinkQuery == "" || strings.Contains(strings.ToLower(n.Name), strings.ToLower(m.wikiLinkQuery)) {
			m.wikiLinkMatches = append(m.wikiLinkMatches, n.Name)
		}
	}
	if m.wikiLinkIdx >= len(m.wikiLinkMatches) {
		m.wikiLinkIdx = 0
	}
}

// acceptWikiLink fügt den gewählten Match als Link-Text ein ([[Name]]) und schließt das Popup.
func (m *Model) acceptWikiLink() {
	if m.buffer == nil {
		m.wikiLinkPopup = false
		m.wikiLinkQuery = ""
		return
	}
	if m.wikiLinkIdx < 0 || m.wikiLinkIdx >= len(m.wikiLinkMatches) {
		m.wikiLinkPopup = false
		m.wikiLinkQuery = ""
		return
	}
	link := m.wikiLinkMatches[m.wikiLinkIdx]
	// Entferne das durch applyAutoPair eingefügte "[[]]" (4 Zeichen vor Cursor).
	// "[[" rückwärts (DeleteChar), "]]" vorwärts (DeleteCharForward).
	for i := 0; i < 2; i++ {
		m.buffer.DeleteChar()
	}
	for i := 0; i < 2; i++ {
		m.buffer.DeleteCharForward()
	}
	// Insert "[[Name]]" als Wiki-Link
	m.buffer.InsertString("[[" + link + "]]")
	m.wikiLinkPopup = false
	m.wikiLinkQuery = ""
	m.wikiLinkIdx = 0
}

// WikiLinkPopupForTest returns whether the wiki-link popup is active (used by tests).
func (m Model) WikiLinkPopupForTest() bool { return m.wikiLinkPopup }

// WikiLinkMatchesForTest returns the current wiki-link match list (used by tests).
func (m Model) WikiLinkMatchesForTest() []string { return m.wikiLinkMatches }

// openOrCreateDailyNote erstellt die Daily-Note für heute, falls sie nicht existiert,
// und lädt sie in den Editor. Format: <workspace>/Daily Notes/YYYY-MM-DD.md.
func (m *Model) openOrCreateDailyNote() (tea.Model, tea.Cmd) {
	if m.workspace == nil {
		return *m, nil
	}
	root := m.workspace.RootPath
	notesDir := filepath.Join(root, "Daily Notes")
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		return *m, nil
	}
	today := time.Now().Format("2006-01-02")
	notePath := filepath.Join(notesDir, today+".md")
	if _, err := os.Stat(notePath); os.IsNotExist(err) {
		body := "# " + today + "\n\n"
		if err := os.WriteFile(notePath, []byte(body), 0644); err != nil {
			return *m, nil
		}
		// Tree neu laden
		m.refreshWorkspace()
		// Zur Notiz im Tree navigieren
		idx := m.findTreeIdxByPath(notePath)
		if idx >= 0 {
			m.cursorIdx = idx
			updated, cmd := m.selectCurrent()
			return updated.(Model), cmd
		}
	}
	// Wenn schon existiert: Datei einfach laden
	if idx := m.findTreeIdxByPath(notePath); idx >= 0 {
		m.cursorIdx = idx
		updated, cmd := m.selectCurrent()
		return updated.(Model), cmd
	}
	return *m, nil
}

// runFindInFiles führt grep.Search() im Workspace aus und aktualisiert die Results.
func (m *Model) runFindInFiles() {
	m.findInFilesResults = nil
	if m.workspace == nil || m.findInFilesQuery == "" {
		return
	}
	m.findInFilesResults = grep.Search(m.workspace.RootPath, m.findInFilesQuery)
	if m.findInFilesIdx >= len(m.findInFilesResults) {
		m.findInFilesIdx = 0
	}
}

// openGrepHit öffnet die Datei eines grep-Treffers und positioniert den Cursor auf der Zeile.
func (m *Model) openGrepHit(hit grep.Hit) {
	if m.workspace == nil {
		return
	}
	idx := m.findTreeIdxByPath(hit.Path)
	if idx < 0 {
		return
	}
	m.cursorIdx = idx
	updated, cmd := m.selectCurrent()
	mm, ok := updated.(Model)
	if !ok {
		_ = cmd
		return
	}
	if mm.buffer != nil {
		if hit.Line < len(mm.buffer.Lines) {
			mm.buffer.CursorRow = hit.Line
			mm.buffer.CursorCol = 0
		}
	}
	*m = mm
}

// FindInFilesModeForTest exposes m.findInFilesMode.
func (m Model) FindInFilesModeForTest() bool { return m.findInFilesMode }

// FindInFilesQueryForTest exposes m.findInFilesQuery.
func (m Model) FindInFilesQueryForTest() string { return m.findInFilesQuery }

// FindInFilesResultsForTest exposes m.findInFilesResults.
func (m Model) FindInFilesResultsForTest() []grep.Hit { return m.findInFilesResults }

// recentFilesList gibt eine Kopie der Recent-Liste zurück (MRU-sorted).
func (m Model) recentFilesList() []recent.File {
	if m.recent == nil {
		return nil
	}
	return m.recent.All()
}

// openRecentFile öffnet eine Datei aus der Recent-Liste und fügt sie oben hinzu.
func (m *Model) openRecentFile(path string) {
	if m.workspace == nil {
		return
	}
	idx := m.findTreeIdxByPath(path)
	if idx >= 0 {
		m.cursorIdx = idx
		updated, cmd := m.selectCurrent()
		if mm, ok := updated.(Model); ok {
			*m = mm
			_ = cmd
		}
	}
}

// RecentMenuModeForTest exposes m.recentMenuMode.
func (m Model) RecentMenuModeForTest() bool { return m.recentMenuMode }

// RecentMenuIdxForTest exposes m.recentMenuIdx.
func (m Model) RecentMenuIdxForTest() int { return m.recentMenuIdx }

// renderTagsTab rendert alle Tags aus dem aktuellen Buffer mit Vorkommen-Anzahl.
func (m Model) renderTagsTab() string {
	if m.buffer == nil || m.buffer.Path == "" {
		return "(keine Datei geöffnet)"
	}
	tags := markdown.Tags(m.buffer.ToString())
	if len(tags) == 0 {
		return "TAGS\n────\nKeine Tags in dieser Datei.\nTipp: #word im Text schreiben."
	}
	// Zähle Vorkommen je Tag
	counts := map[string]int{}
	lines := map[string][]int{}
	for _, t := range tags {
		counts[t.Name]++
		lines[t.Name] = append(lines[t.Name], t.Line+1)
	}
	// Sortiere nach Häufigkeit (absteigend)
	type kv struct {
		name  string
		count int
		first int
	}
	var kvs []kv
	for n, c := range counts {
		kvs = append(kvs, kv{n, c, lines[n][0]})
	}
	sort.Slice(kvs, func(i, j int) bool {
		if kvs[i].count != kvs[j].count {
			return kvs[i].count > kvs[j].count
		}
		return kvs[i].name < kvs[j].name
	})
	out := []string{fmt.Sprintf("TAGS (%d unique, %d total)", len(kvs), len(tags))}
	out = append(out, "─────────────────────────")
	for _, k := range kvs {
		out = append(out, fmt.Sprintf("  \x1b[38;5;141m%s\x1b[0m  %dx  \x1b[2m(line %d)\x1b[0m", k.name, k.count, k.first))
	}
	return strings.Join(out, "\n")
}

// TagsTabTextForTest exposes the Tags tab render output for tests.
func (m Model) TagsTabTextForTest() string {
	return m.renderTagsTab()
}

// toggleCurrentTaskLine toggelt eine Markdown-Task-Liste-Zeile an der aktuellen
// Cursor-Position: "- [ ] foo" ↔ "- [x] foo". Nicht-Task-Zeilen bleiben unverändert.
func (m *Model) toggleCurrentTaskLine() {
	if m.buffer == nil {
		return
	}
	if m.buffer.CursorRow < 0 || m.buffer.CursorRow >= len(m.buffer.Lines) {
		return
	}
	line := m.buffer.Lines[m.buffer.CursorRow]
	// Patterns: "- [ ] foo" / "- [x] foo" / "* [ ] foo" / "+ [ ] foo"
	taskRe := regexp.MustCompile(`^(\s*)([-*+]) \[([ xX])\]`)
	matches := taskRe.FindStringSubmatchIndex(line)
	if matches == nil {
		return
	}
	// groups: 1=indent, 2=list-marker, 3=bracket-char (space/x/X)
	// matches[6]:7 = group3 (bracket char) start/end
	start, end := matches[6], matches[7]
	if start < 0 || end < 0 || end > len(line) {
		return
	}
	runes := []rune(line)
	if start >= len(runes) {
		return
	}
	currentChar := runes[start]
	var newChar rune
	if currentChar == ' ' {
		newChar = 'x'
	} else {
		newChar = ' '
	}
	runes[start] = newChar
	newLine := string(runes)
	// Use Buffer method (sets Modified + snapshots for undo/redo)
	// instead of direct Lines[row] mutation.
	m.buffer.SetLine(m.buffer.CursorRow, newLine)
}
