# L5 — Collab-Editing

## Zweck
Geteilte Buffer über mehrere Sessions (CRDT-basiert).

## MVP-Status
- Out-of-MVP (R10)
- Nicht im MVP-Scope

## Architektur-Stub
```go
// internal/obsidian/l5_collab.go (R10) — PLACEHOLDER
package obsidian

type CRDTDocument struct{}

func (d *CRDTDocument) Apply(operation interface{}) {
	// R10: Stub, noop
}
```

## Nächste Schritte (post-MVP)
- Yjs/Automerge-Integration
- WebSocket-Transport
- Peer-Discovery
