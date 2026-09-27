# L8 — Telemetry

## Zweck
Anonyme Usage-Statistiken zur Produktverbesserung.

## MVP-Status
- Out-of-MVP (R10)
- Opt-in only, kein Default-Tracking

## Architektur-Stub
```go
// internal/obsidian/l8_telemetry.go (R10) — PLACEHOLDER
package obsidian

type Event struct {
	Name string
	Properties map[string]interface{}
}

var Sink []Event // R10: Stub, leer
```

## Nächste Schritte (post-MVP)
- Opt-in Dialog beim ersten Start
- Privacy-konforme Event-Klassifikation
- Local-First-Option
