// Package vulkan provides the Vulkan project template generator for cpp-gen.
//
// It embeds all template files from the templates/ subdirectory and
// writes them to the target project root, applying Go text/template
// substitution to files that contain project-specific variables.
package vulkan

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed templates
var templateFS embed.FS

// Data contains all template variables used during Vulkan project generation.
type Data struct {
	// Name is the original project name (e.g. "my-vulkan-app").
	Name string

	// NameUpper is the name in UPPER_SNAKE_CASE (e.g. "MY_VULKAN_APP").
	NameUpper string

	// NameSnake is the name in snake_case (e.g. "my_vulkan_app").
	NameSnake string

	// NamePascal is the name in PascalCase (e.g. "MyVulkanApp").
	NamePascal string

	// Description is the project description.
	Description string

	// Version is the project version in SemVer format (e.g. "1.0.0").
	Version string

	// Standard is the C++ standard as a numeric string (e.g. "23").
	Standard string

	// UseVCPKG indicates whether to generate vcpkg.json and the VCPKG toolchain
	// integration in CMakeLists.txt. When false, the project relies on system
	// packages or manual dependency management.
	UseVCPKG bool
}

// filesNeedingSubstitution lists paths relative to templates/ that are
// processed through text/template for variable substitution.
// All other files are copied verbatim.
var filesNeedingSubstitution = map[string]bool{
	"CMakeLists.txt":       true,
	"CMakePresets.json":    true,
	"vcpkg.json":           true,
	"tests/CMakeLists.txt": true,
	"src/main.cpp":         true,
}

// vcpkgOnlyFiles lists files that should only be generated when UseVCPKG is true.
var vcpkgOnlyFiles = map[string]bool{
	"vcpkg.json":               true,
	"vcpkg-configuration.json": true,
}

// Generate writes all Vulkan template files to root, applying
// Go template substitution for designated files.
func Generate(root string, data *Data, verbose bool) error {
	return fs.WalkDir(templateFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Strip the "templates/" prefix to get the relative destination path.
		relPath, relErr := filepath.Rel("templates", path)
		if relErr != nil {
			return relErr
		}

		if d.IsDir() {
			if relPath == "." {
				return nil
			}
			return os.MkdirAll(filepath.Join(root, filepath.FromSlash(relPath)), 0755)
		}

		relPathFwd := filepath.ToSlash(relPath)

		// Skip VCPKG-specific files when VCPKG is not enabled.
		if !data.UseVCPKG && vcpkgOnlyFiles[relPathFwd] {
			if verbose {
				fmt.Printf("    ~ %s (skipped â VCPKG disabled)\n", relPath)
			}
			return nil
		}

		destPath := filepath.Join(root, filepath.FromSlash(relPath))

		if mkErr := os.MkdirAll(filepath.Dir(destPath), 0755); mkErr != nil {
			return fmt.Errorf("mkdir para %s: %w", destPath, mkErr)
		}

		if verbose {
			fmt.Printf("    + %s\n", relPath)
		}

		// Use forward slashes as map keys regardless of OS.
		if filesNeedingSubstitution[relPathFwd] {
			return applyTemplate(path, destPath, data)
		}
		return copyFile(path, destPath)
	})
}

// applyTemplate processes a file through Go text/template and writes the result.
func applyTemplate(srcPath, destPath string, data *Data) error {
	content, err := templateFS.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("ler template %s: %w", srcPath, err)
	}

	tmpl, err := template.New(filepath.Base(srcPath)).Parse(string(content))
	if err != nil {
		return fmt.Errorf("parsear template %s: %w", srcPath, err)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("criar %s: %w", destPath, err)
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

// copyFile writes a file from the embedded FS to destPath verbatim.
func copyFile(srcPath, destPath string) error {
	content, err := templateFS.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("ler %s: %w", srcPath, err)
	}
	return os.WriteFile(destPath, content, 0644)
}
