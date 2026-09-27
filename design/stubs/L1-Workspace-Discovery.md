# L1 — Workspace Discovery

## Zweck
Erkennt Workspace-Typen und konfiguriert default-renderer.

## MVP-Status
- Out-of-MVP (R10)
- Vorbereitung: `workspace/Load()` mit Type-Detection

## Architektur-Stub
```go
// internal/obsidian/l1_discovery.go (R10) — PLACEHOLDER
package obsidian

func DetectWorkspaceType(path string) WorkspaceType {
	return TypeWorkspace // MVP: nur Workspace-Variante
}
```

## Nächste Schritte (post-MVP)
- Git-Repo-Detection (.git/ vorhanden?)
- Obsidian-Vault-Detection (.obsidian/ vorhanden?)
- Mixed-Workspace (mehrere Typen in einem Tree)
- Workspace-Type-Indicator in Status-Bar
