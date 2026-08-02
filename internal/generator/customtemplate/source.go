// Package customtemplate implements external project templates fetched from
// a local folder or a Git repository, in the spirit of `npm create <template>`
// or `cargo generate`. Unlike the built-in "blank" and "vulkan" templates,
// these are not compiled into the cpp-gen binary.
package customtemplate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SourceKind identifies where a template comes from.
type SourceKind int

const (
	// SourceLocal reads the template directly from a folder on disk.
	SourceLocal SourceKind = iota

	// SourceGit clones the template from a Git repository.
	SourceGit
)

// Source describes a resolved template location, as produced by ParseSource.
type Source struct {
	Kind SourceKind

	// LocalPath is the absolute path to the template folder, set when Kind == SourceLocal.
	LocalPath string

	// GitURL is the full clone URL, set when Kind == SourceGit.
	GitURL string

	// Ref is an optional branch or tag name (from a trailing "#ref").
	Ref string

	// Subdir is an optional subdirectory within the cloned repository that
	// holds the actual template (from a GitHub shorthand like "owner/repo/subdir").
	Subdir string
}

// ParseSource classifies a raw --template value into a local path or a Git
// source. Accepted forms:
//
//	<local-path>                          folder on disk (".", "/abs", "~/x", or anything that exists as a dir)
//	owner/repo[/subdir...][#ref]          GitHub shorthand, expanded to https://github.com/owner/repo.git
//	https://.../repo.git[#ref]            full HTTPS URL
//	git@host:owner/repo.git[#ref]         full SSH URL
func ParseSource(raw string) (Source, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Source{}, fmt.Errorf("fonte de template vazia")
	}

	body, ref := splitRef(raw)
	if body == "" {
		return Source{}, fmt.Errorf("fonte de template inválida %q", raw)
	}

	if isLocalLike(body) {
		abs, err := filepath.Abs(expandHome(body))
		if err != nil {
			return Source{}, fmt.Errorf("resolver caminho local %q: %w", body, err)
		}
		return Source{Kind: SourceLocal, LocalPath: abs, Ref: ref}, nil
	}

	if strings.HasPrefix(body, "http://") || strings.HasPrefix(body, "https://") || strings.HasPrefix(body, "git@") {
		return Source{Kind: SourceGit, GitURL: body, Ref: ref}, nil
	}

	// GitHub shorthand: owner/repo[/subdir...]
	parts := strings.Split(body, "/")
	if len(parts) >= 2 && parts[0] != "" && parts[1] != "" {
		owner, repo := parts[0], parts[1]
		subdir := strings.Join(parts[2:], "/")
		url := fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
		return Source{Kind: SourceGit, GitURL: url, Ref: ref, Subdir: subdir}, nil
	}

	return Source{}, fmt.Errorf(
		"fonte de template inválida %q; use um caminho local, owner/repo, ou uma URL git", raw,
	)
}

// splitRef separates a trailing "#ref" (branch/tag) from the source body.
func splitRef(s string) (body, ref string) {
	if idx := strings.LastIndex(s, "#"); idx != -1 {
		return s[:idx], s[idx+1:]
	}
	return s, ""
}

// isLocalLike reports whether body looks like (or actually is) a local
// filesystem path rather than a remote Git source.
func isLocalLike(body string) bool {
	if strings.HasPrefix(body, ".") || strings.HasPrefix(body, "/") || strings.HasPrefix(body, "~") {
		return true
	}
	if info, err := os.Stat(expandHome(body)); err == nil && info.IsDir() {
		return true
	}
	return false
}

// expandHome resolves a leading "~" or "~/" to the user's home directory.
func expandHome(p string) string {
	if p == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
		return p
	}
	if strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, p[2:])
		}
	}
	return p
}
