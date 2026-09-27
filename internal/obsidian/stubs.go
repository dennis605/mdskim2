// Package obsidian enthält experimentelle, out-of-MVP Architektur-Stubs.
// Siehe design/stubs/L*.md für den Scope und die nächsten Schritte.
package obsidian

// L1 — Workspace-Discovery
type WorkspaceType int

const (
	TypeWorkspace WorkspaceType = iota
	TypeGitRepo
	TypeObsidianVault
	TypeMixed
)

func DetectWorkspaceType(path string) WorkspaceType {
	// R10: Placeholder, gibt MVP-Variante zurück
	return TypeWorkspace
}

// L2 — Multi-Workspace-Tabs
func OpenWorkspace(path string) {
	// R10: Placeholder
}

// L3 — Plugin-Interface
type Plugin interface {
	Name() string
	Process(line string) string
}

var PluginRegistry []Plugin

// L4 — AI-Assist
type AIProvider interface {
	Complete(prefix string) (suggestion string, err error)
}

var CurrentAIProvider AIProvider

// L5 — Collab-CRDT
type CRDTDoc struct{}

func (d *CRDTDoc) Apply(_ interface{}) {}

// L6 — Cloud-Sync-Backend
type SyncBackend interface {
	Pull(path string) error
	Push(path string) error
}

var CurrentBackend SyncBackend

// L7 — Web-Companion
func WebListen(_ string) error { return nil }

// L8 — Telemetry
type Event struct {
	Name       string
	Properties map[string]interface{}
}

var EventSink []Event
