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
)

// KeyMapping ordnet spezifische Taste-Notationen einer Action zu.
// Wird in späteren Sprints genutzt; R1 hat die Logik in handleKey().
type KeyMapping struct {
	Key       string  // z.B. "ctrl+s"
	HelpText  string  // z.B. "Save"
	Action    Action
	Implemented bool // Sprint-Flag
}

// AllShortcuts listet alle Spec-Shortcuts mit Implementierungs-Status.
func AllShortcuts() []KeyMapping {
	return []KeyMapping{
		// R1 (Bootstrap)
		{Key: "ctrl+q", HelpText: "Quit", Action: QuitAction, Implemented: true},

		// R4 (Standard-Shortcuts + Clipboard)
		{Key: "ctrl+s", HelpText: "Save", Action: SaveAction, Implemented: false},
		{Key: "ctrl+o", HelpText: "Open", Action: OpenAction, Implemented: false},
		{Key: "ctrl+n", HelpText: "New File", Action: NewFileAction, Implemented: false},
		{Key: "ctrl+c", HelpText: "Copy", Action: NoOp, Implemented: false},
		{Key: "ctrl+x", HelpText: "Cut", Action: NoOp, Implemented: false},
		{Key: "ctrl+v", HelpText: "Paste", Action: NoOp, Implemented: false},
		{Key: "ctrl+z", HelpText: "Undo", Action: NoOp, Implemented: false},
		{Key: "ctrl+y", HelpText: "Redo", Action: NoOp, Implemented: false},
		{Key: "ctrl+a", HelpText: "Select All", Action: NoOp, Implemented: false},
		{Key: "ctrl+b", HelpText: "Bold", Action: BoldAction, Implemented: false},
		{Key: "ctrl+i", HelpText: "Italic", Action: ItalicAction, Implemented: false},

		// R6 (Suche)
		{Key: "ctrl+f", HelpText: "Find", Action: FindAction, Implemented: false},
		{Key: "ctrl+h", HelpText: "Replace", Action: ReplaceAction, Implemented: false},

		// R8 (Quick Open, Tabs)
		{Key: "ctrl+p", HelpText: "Quick Open", Action: QuickOpenAction, Implemented: false},
		{Key: "ctrl+w", HelpText: "Close Tab", Action: CloseTabAction, Implemented: false},

		// R9 (Command Palette, Workspace-Suche)
		{Key: "ctrl+shift+p", HelpText: "Commands", Action: CommandPaletteAction, Implemented: false},
	}
}
