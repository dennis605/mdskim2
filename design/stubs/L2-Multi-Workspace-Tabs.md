# L2 — Multi-Workspace-Tabs

## Zweck
Mehrere unabhängige Workspaces parallel offen halten.

## MVP-Status
- Out-of-MVP (R10)
- Vorbereitung: `tabs.Manager` ist bereits Thread-Safe

## Architektur-Stub
```go
// internal/obsidian/l2_multiworkspace.go (R10) — PLACEHOLDER
package obsidian

func (m *Manager) OpenWorkspace(path string) {
	// R10: Stub
}
```

## Nächste Schritte (post-MVP)
- Workspace-Picker (Ctrl+Shift+O)
- Workspace-Switcher
- Per-Workspace Buffer-History
