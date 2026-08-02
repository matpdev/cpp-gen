// Package registry implements discovery of external cpp-gen templates
// (internal/generator/customtemplate) across three merged sources:
//
//  1. An empty default catalog embedded in the binary (catalog.json),
//     meant to be populated by maintainers over time.
//  2. A curated JSON registry fetched from GitHub (or a local file/URL
//     override via CPPGEN_REGISTRY_URL), shared across all users.
//  3. A user-local file at <GlobalConfigDir>/templates.json, which the
//     user edits directly to add or override entries.
//
// All three use the same JSON shape: {"templates": [{"name", "description",
// "source", "tags"}, ...]}. "source" is anything accepted by
// customtemplate.ParseSource (local path, "owner/repo[/subdir][#ref]", or a
// full Git URL).
package registry

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Entry describes a single discoverable template.
type Entry struct {
	// Name is the short identifier used for lookup (e.g. "raylib-app").
	// Matching is case-insensitive.
	Name string `json:"name"`

	// Description explains what the template does/is for, shown to the user
	// when listing or searching.
	Description string `json:"description"`

	// Source is anything customtemplate.ParseSource accepts: a local path,
	// a "owner/repo[/subdir][#ref]" shorthand, or a full Git URL.
	Source string `json:"source"`

	// Tags are free-form keywords matched during Search (e.g. "game", "gui").
	Tags []string `json:"tags,omitempty"`
}

// Catalog is a merged, deduplicated list of Entry, ready for display or search.
type Catalog struct {
	Entries []Entry
}

// catalogFile is the on-disk/wire JSON shape shared by all three sources.
type catalogFile struct {
	Templates []Entry `json:"templates"`
}

// parseCatalogJSON decodes a catalogFile from raw JSON bytes.
func parseCatalogJSON(data []byte) ([]Entry, error) {
	var f catalogFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("JSON de catálogo inválido: %w", err)
	}
	return f.Templates, nil
}

// Load merges all three sources into a single Catalog, sorted by name.
// It never fails outright — a source that can't be read (offline, missing
// file, malformed JSON) is skipped; a warning is printed when verbose is true.
//
// Merge priority (highest wins on name collision): local user file >
// GitHub registry > embedded defaults.
func Load(verbose bool) Catalog {
	merged := map[string]Entry{}

	for _, e := range embeddedEntries(verbose) {
		merged[strings.ToLower(e.Name)] = e
	}

	ghEntries, err := loadGitHub()
	if err != nil {
		warnf(verbose, "não foi possível carregar o registro do GitHub: %v", err)
	}
	for _, e := range ghEntries {
		merged[strings.ToLower(e.Name)] = e
	}

	localEntries, err := loadLocal()
	if err != nil {
		warnf(verbose, "não foi possível carregar templates locais: %v", err)
	}
	for _, e := range localEntries {
		merged[strings.ToLower(e.Name)] = e
	}

	entries := make([]Entry, 0, len(merged))
	for _, e := range merged {
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool {
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})

	return Catalog{Entries: entries}
}

// Resolve looks up a template by exact name (case-insensitive), checking
// embedded defaults and the local user file first (no I/O beyond a local
// file read), and only falling back to the network-backed GitHub registry
// when the name isn't found locally. Used by `cpp-gen new --template <name>`
// to resolve a bare alias into an actual source.
func Resolve(name string, verbose bool) (Entry, bool) {
	key := strings.ToLower(strings.TrimSpace(name))
	if key == "" {
		return Entry{}, false
	}

	for _, e := range embeddedEntries(verbose) {
		if strings.ToLower(e.Name) == key {
			return e, true
		}
	}

	localEntries, err := loadLocal()
	if err != nil {
		warnf(verbose, "não foi possível carregar templates locais: %v", err)
	}
	for _, e := range localEntries {
		if strings.ToLower(e.Name) == key {
			return e, true
		}
	}

	ghEntries, err := loadGitHub()
	if err != nil {
		warnf(verbose, "não foi possível carregar o registro do GitHub: %v", err)
	}
	for _, e := range ghEntries {
		if strings.ToLower(e.Name) == key {
			return e, true
		}
	}

	return Entry{}, false
}

// Search filters Entries by a case-insensitive substring match against name,
// description and tags. An empty query returns every entry.
func (c Catalog) Search(query string) []Entry {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return c.Entries
	}

	var out []Entry
	for _, e := range c.Entries {
		haystack := strings.ToLower(e.Name + " " + e.Description + " " + strings.Join(e.Tags, " "))
		if strings.Contains(haystack, query) {
			out = append(out, e)
		}
	}
	return out
}

// warnf prints a soft warning to stdout when verbose is true. Registry
// lookups degrade gracefully — a source being unreachable never blocks
// `cpp-gen new`, it just means fewer templates show up.
func warnf(verbose bool, format string, args ...any) {
	if !verbose {
		return
	}
	fmt.Printf("    ⚠ registro: %s\n", fmt.Sprintf(format, args...))
}
