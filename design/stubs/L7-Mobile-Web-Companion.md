# L7 — Mobile/Web-Companion

## Zweck
Read-only Web-Preview für mobile Browser.

## MVP-Status
- Out-of-MVP (R10)
- Nicht im MVP-Scope

## Architektur-Stub
```go
// internal/obsidian/l7_webcompanion.go (R10) — PLACEHOLDER
package obsidian

func WebListen(addr string) error {
	return nil // R10: Stub, noop
}
```

## Nächste Schritte (post-MVP)
- Embedded HTTP-Server (chi/gin)
- WebSocket-Updates
- Read-only Mobile-UI
