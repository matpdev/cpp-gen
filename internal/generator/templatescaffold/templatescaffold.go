// Package templatescaffold generates a starter custom template — a folder
// with the file/placeholder conventions internal/generator/customtemplate
// expects, ready for the user to edit into their own template.
//
// Unlike customtemplate.Generate, files here are copied byte-for-byte: the
// whole point is to leave "{{.Name}}"-style placeholders intact on disk for
// the user to read and build on, not to substitute them immediately.
package templatescaffold

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed files
var filesFS embed.FS

// Generate copies the embedded starter template into destDir, creating it
// (and any parent directories) as needed. It refuses to run against a
// destDir that already exists and isn't empty, to avoid clobbering existing
// work.
func Generate(destDir string) error {
	if entries, err := os.ReadDir(destDir); err == nil && len(entries) > 0 {
		return fmt.Errorf("%s já existe e não está vazio", destDir)
	}

	return fs.WalkDir(filesFS, "files", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, relErr := filepath.Rel("files", path)
		if relErr != nil {
			return relErr
		}
		if relPath == "." {
			return nil
		}

		destPath := filepath.Join(destDir, filepath.FromSlash(relPath))

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		if mkErr := os.MkdirAll(filepath.Dir(destPath), 0755); mkErr != nil {
			return fmt.Errorf("mkdir para %s: %w", destPath, mkErr)
		}

		content, readErr := filesFS.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("ler %s: %w", path, readErr)
		}

		return os.WriteFile(destPath, content, 0644)
	})
}
