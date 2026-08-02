package registry

import _ "embed"

//go:embed catalog.json
var embeddedCatalogJSON []byte

// embeddedEntries returns the default catalog bundled into the cpp-gen
// binary. It starts empty and is meant to be populated by maintainers over
// releases; a malformed catalog.json is a packaging bug, so it's reported
// but never crashes the CLI.
func embeddedEntries(verbose bool) []Entry {
	entries, err := parseCatalogJSON(embeddedCatalogJSON)
	if err != nil {
		warnf(verbose, "catálogo embutido inválido: %v", err)
		return nil
	}
	return entries
}
