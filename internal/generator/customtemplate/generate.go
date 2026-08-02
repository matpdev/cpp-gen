package customtemplate

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Generate walks templateRoot, applying Go text/template substitution to
// both file contents and destination paths, and writes the result under
// destRoot. data is typically a *generator.TemplateData, passed as any to
// avoid an import cycle between this package and generator.
//
// A file or directory whose name renders to an empty string is skipped
// entirely — this lets a template conditionally include a whole file based
// on the user's choices, e.g. naming a file
// "{{if .UseVCPKG}}vcpkg.json{{end}}" so it's only generated when the user
// picked VCPKG as the package manager.
//
// Binary files and files that aren't valid templates (e.g. C++ using
// brace-init syntax like "arr{{1,2}}" is not valid template syntax) are
// copied verbatim instead of aborting generation.
func Generate(templateRoot, destRoot string, data any, verbose bool) error {
	fsys := os.DirFS(templateRoot)

	return fs.WalkDir(fsys, ".", func(relPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}
		if d.IsDir() && d.Name() == ".git" {
			return fs.SkipDir
		}

		destRelPath := renderPath(relPath, data)
		if strings.TrimSpace(destRelPath) == "" {
			if verbose {
				fmt.Printf("    ~ %s (omitido — condição do nome não satisfeita)\n", relPath)
			}
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		destPath := filepath.Join(destRoot, filepath.FromSlash(destRelPath))

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		if mkErr := os.MkdirAll(filepath.Dir(destPath), 0755); mkErr != nil {
			return fmt.Errorf("mkdir para %s: %w", destPath, mkErr)
		}

		content, readErr := fs.ReadFile(fsys, relPath)
		if readErr != nil {
			return fmt.Errorf("ler %s: %w", relPath, readErr)
		}

		rendered, substituted := renderContent(relPath, content, data)
		if verbose {
			if substituted {
				fmt.Printf("    + %s\n", destRelPath)
			} else {
				fmt.Printf("    ~ %s (copiado sem substituição — binário ou não é um template válido)\n", destRelPath)
			}
		}

		return os.WriteFile(destPath, rendered, 0644)
	})
}

// renderPath attempts to run relPath through text/template so file/dir
// names can contain placeholders like "{{.NameSnake}}.hpp". Falls back to
// the literal path when it isn't a valid template.
func renderPath(relPath string, data any) string {
	if !strings.Contains(relPath, "{{") {
		return relPath
	}
	tmpl, err := template.New("path").Parse(relPath)
	if err != nil {
		return relPath
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return relPath
	}
	return buf.String()
}

// renderContent attempts to render content as a Go template. It returns the
// original bytes unchanged (substituted=false) when the content looks
// binary or isn't a valid template, so callers can copy it verbatim.
func renderContent(relPath string, content []byte, data any) (rendered []byte, substituted bool) {
	if isBinary(content) {
		return content, false
	}

	tmpl, err := template.New(filepath.Base(relPath)).Parse(string(content))
	if err != nil {
		return content, false
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return content, false
	}
	return buf.Bytes(), true
}

// isBinary reports whether content looks like a binary file, using the
// common heuristic of checking for a NUL byte in the first 8000 bytes.
func isBinary(content []byte) bool {
	n := min(len(content), 8000)
	return bytes.IndexByte(content[:n], 0) != -1
}
