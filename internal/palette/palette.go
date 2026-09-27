// Package palette implementiert eine Mini-Command-Line-Sprache (mdskim2-cmds).
package palette

import (
	"sort"
	"strings"
)

// Command repräsentiert ein ausführbares Kommando.
type Command struct {
	Name        string
	Description string
	Keywords    []string
}

// Registry hält alle verfügbaren Commands.
type Registry struct {
	commands map[string]*Command
	order    []string
}

// NewRegistry erzeugt eine leere Registry.
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]*Command),
	}
}

// Register fügt einen Command hinzu.
func (r *Registry) Register(cmd Command) {
	r.commands[cmd.Name] = &cmd
	r.order = append(r.order, cmd.Name)
}

// Run führt einen Command aus und gibt die Description zurück.
// Echte Side-Effects werden über einen Callback gehandhabt.
func (r *Registry) Run(name string) (string, bool) {
	cmd, ok := r.commands[name]
	if !ok {
		return "", false
	}
	return cmd.Description, true
}

// Search sucht Commands nach Query (Substring in Name oder Keywords).
func (r *Registry) Search(query string) []string {
	var results []string
	for _, name := range r.order {
		cmd := r.commands[name]
		if query == "" {
			results = append(results, name)
			continue
		}
		q := strings.ToLower(query)
		if strings.Contains(strings.ToLower(name), q) {
			results = append(results, name)
			continue
		}
		for _, kw := range cmd.Keywords {
			if strings.Contains(strings.ToLower(kw), q) {
				results = append(results, name)
				break
			}
		}
	}
	sort.Strings(results)
	return results
}
