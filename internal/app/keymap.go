// Package app: keymap.go bündelt alle Spec-Shortcuts an einer Stelle.
// R1 hat nur Ctrl+Q implementiert; alle anderen Keys sind in
// handleKey() bzw. werden in R3-R9 implementiert.
package app

// Action repräsentiert eine semantische Aktion, die durch einen Shortcut
// ausgelöst wird. R1 hat nur QuitAction realisiert.
type Action int

const (
	NoOp Action = iota
	QuitAction
	SaveAction
	OpenAction
	QuickOpenAction
	CommandPaletteAction
	FindAction
	ReplaceAction
	BoldAction
	ItalicAction
	NewFileAction
	CloseTabAction
	DailyNoteAction
	FindInFilesAction
	RecentFilesAction
	NewTabAction
	RefreshAction
	GoToLineAction
	CyclePaneAction
	CyclePaneReverseAction
	ReadOnlyAction
	CloseFileAction
	FocusTreeAction
	FocusEditorAction
	FocusRightAction
	RightTabPreviewAction
	RightTabTOCAction
	RightTabBacklinksAction
	RightTabTagsAction
	TaskToggleAction
	NewFileTreeAction
	NewFolderTreeAction
	DeleteTreeAction
	RenameTreeAction
	TreeFilterAction
	WikiLinkAction
)

// KeyMapping ordnet spezifische Taste-Notationen einer Action zu.
// Wird in späteren Sprints genutzt; R1 hat die Logik in handleKey().
type KeyMapping struct {
	Key         string // z.B. "ctrl+s"
	HelpText    string // z.B. "Save"
	Action      Action
	Implemented bool // Sprint-Flag
}

// AllShortcuts listet alle Spec-Shortcuts mit Implementierungs-Status.
func AllShortcuts() []KeyMapping {
	return []KeyMapping{
		// R1 (Bootstrap)
		{Key: "ctrl+q", HelpText: "Quit", Action: QuitAction, Implemented: true},

		// R4 (Standard-Shortcuts + Clipboard)
		{Key: "ctrl+s", HelpText: "Save", Action: SaveAction, Implemented: true},
		{Key: "ctrl+o", HelpText: "Open", Action: OpenAction, Implemented: false},
		{Key: "ctrl+n", HelpText: "New File", Action: NewFileAction, Implemented: false},
		{Key: "ctrl+c", HelpText: "Copy", Action: NoOp, Implemented: true},
		{Key: "ctrl+x", HelpText: "Cut", Action: NoOp, Implemented: true},
		{Key: "ctrl+v", HelpText: "Paste", Action: NoOp, Implemented: true},
		{Key: "ctrl+z", HelpText: "Undo", Action: NoOp, Implemented: true},
		{Key: "ctrl+y", HelpText: "Redo", Action: NoOp, Implemented: true},
		{Key: "ctrl+a", HelpText: "Select All", Action: NoOp, Implemented: false},
		{Key: "ctrl+b", HelpText: "Bold", Action: BoldAction, Implemented: true},
		{Key: "ctrl+i", HelpText: "Italic", Action: ItalicAction, Implemented: true},

		// R6 (Suche)
		{Key: "ctrl+f", HelpText: "Find", Action: FindAction, Implemented: true},
		{Key: "ctrl+h", HelpText: "Replace", Action: ReplaceAction, Implemented: true},

		// R8 (Quick Open, Tabs)
		{Key: "ctrl+p", HelpText: "Quick Open", Action: QuickOpenAction, Implemented: true},
		{Key: "ctrl+w", HelpText: "Close Tab", Action: CloseTabAction, Implemented: true},

		// R9 (Command Palette, Workspace-Suche)
		{Key: "ctrl+shift+p", HelpText: "Commands", Action: CommandPaletteAction, Implemented: true},

		// U9 (Discoverability — Obsidian-style)
		{Key: "ctrl+shift+d", HelpText: "Daily Note", Action: DailyNoteAction, Implemented: true},
		{Key: "ctrl+shift+f", HelpText: "Find in Files", Action: FindInFilesAction, Implemented: true},
		{Key: "ctrl+shift+o", HelpText: "Recent Files", Action: RecentFilesAction, Implemented: true},

		// U6 — Pane Focus
		{Key: "f6", HelpText: "Next Pane", Action: CyclePaneAction, Implemented: true},
		{Key: "shift+f6", HelpText: "Previous Pane", Action: CyclePaneReverseAction, Implemented: true},
		{Key: "alt+1", HelpText: "Focus Tree", Action: FocusTreeAction, Implemented: true},
		{Key: "alt+2", HelpText: "Focus Editor", Action: FocusEditorAction, Implemented: true},
		{Key: "alt+3", HelpText: "Focus Right Pane", Action: FocusRightAction, Implemented: true},
		{Key: "alt+4", HelpText: "Focus Right Pane", Action: FocusRightAction, Implemented: true},
		{Key: "alt+5", HelpText: "Right Tab: Preview", Action: RightTabPreviewAction, Implemented: true},
		{Key: "alt+6", HelpText: "Right Tab: TOC", Action: RightTabTOCAction, Implemented: true},
		{Key: "alt+7", HelpText: "Right Tab: Backlinks", Action: RightTabBacklinksAction, Implemented: true},
		{Key: "alt+8", HelpText: "Right Tab: Tags", Action: RightTabTagsAction, Implemented: true},

		// U6.1 — Ctrl+Shift+Arrow Pane Navigation
		// (handled via msg.String() prefix, no dedicated constant here)

		// U3 — Editor Advanced
		{Key: "ctrl+l", HelpText: "Go-to-Line", Action: GoToLineAction, Implemented: true},
		{Key: "ctrl+g", HelpText: "Go-to-Line (alias)", Action: GoToLineAction, Implemented: true},

		// U4 — Read-Only / Auto-Save
		{Key: "ctrl+shift+r", HelpText: "Read-Only Toggle", Action: ReadOnlyAction, Implemented: true},

		// U8.1 — Tree File Operations
		{Key: "a", HelpText: "New File (Tree)", Action: NewFileTreeAction, Implemented: true},
		{Key: "shift+a", HelpText: "New Folder (Tree)", Action: NewFolderTreeAction, Implemented: true},
		{Key: "d", HelpText: "Delete (Tree)", Action: DeleteTreeAction, Implemented: true},
		{Key: "r", HelpText: "Rename (Tree)", Action: RenameTreeAction, Implemented: true},
		{Key: "f5", HelpText: "Refresh Tree", Action: RefreshAction, Implemented: true},
		{Key: "ctrl+r", HelpText: "Refresh Workspace", Action: RefreshAction, Implemented: true},

		// U8.3 — Tree Filter
		{Key: "/", HelpText: "Filter Tree", Action: TreeFilterAction, Implemented: true},

		// U8.4 — Wiki-Link Autocomplete
		{Key: "[[", HelpText: "Wiki-Link Popup", Action: WikiLinkAction, Implemented: true},

		// U9.5 — Task Toggle
		{Key: "ctrl+enter", HelpText: "Toggle Task [ ]/  [x]", Action: TaskToggleAction, Implemented: true},
	}
}
