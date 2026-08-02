package registry

import (
	"fmt"

	"github.com/matpdev/cpp-gen/internal/generator/customtemplate"
)

// ResolveSource turns a --template value that isn't a built-in name ("blank",
// "vulkan") into a concrete template source, for use as
// config.ProjectConfig.TemplateSource.
//
// If raw already looks like an explicit source (local path,
// "owner/repo[/subdir][#ref]", or a Git URL — validated via
// customtemplate.ParseSource), it's returned as-is, with no registry lookup
// and no network access. Otherwise raw is treated as a registry alias (e.g.
// "raylib-app") and resolved via Resolve.
func ResolveSource(raw string, verbose bool) (string, error) {
	if _, err := customtemplate.ParseSource(raw); err == nil {
		return raw, nil
	}

	entry, ok := Resolve(raw, verbose)
	if !ok {
		return "", fmt.Errorf(
			"template %q não encontrado (não é um caminho local, owner/repo, URL git, "+
				"nem um alias cadastrado); rode `cpp-gen templates` para ver os disponíveis",
			raw,
		)
	}

	return entry.Source, nil
}
