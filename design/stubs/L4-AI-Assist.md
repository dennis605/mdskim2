# L4 — AI-Assist

## Zweck
Inline-AI-Vorschläge während des Editierens.

## MVP-Status
- Out-of-MVP (R10)
- Nicht im MVP-Scope

## Architektur-Stub
```go
// internal/obsidian/l4_aiassist.go (R10) — PLACEHOLDER
package obsidian

type AIProvider interface {
	Complete(prefix string) (suggestion string, err error)
}

var Provider AIProvider // R10: Stub, noop
```

## Nächste Schritte (post-MVP)
- OpenAI/Anthropic-Provider
- Local-Ollama-Provider
- Suggestion-Modal (Ctrl+Space)
- Inline-Suggestion (Auto-Complete)
