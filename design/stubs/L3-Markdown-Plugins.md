# L3 — Markdown-Plugins

## Zweck
Plug-in-System für Markdown-Extensions (Math, Mermaid, Wiki-Links).

## MVP-Status
- Out-of-MVP (R10)
- Highlighter ist hartkodiert (Headings, Bold, Italic, Code, Links)

## Architektur-Stub
```go
// internal/obsidian/l3_plugins.go (R10) — PLACEHOLDER
package obsidian

type Plugin interface {
	Name() string
	Process(line string) string
}

var Registry []Plugin // R10: Stub, leer
```

## Nächste Schritte (post-MVP)
- Go-Plugin-System (hashicorp/go-plugin)
- Wasm-Plugin-System (wazero)
- Mermaid-Renderer (Inline-SVG via Graphviz)
- Math-Jax-Renderer (KaTeX/LaTeX)
- Wiki-Link-Parser ([[link]])
