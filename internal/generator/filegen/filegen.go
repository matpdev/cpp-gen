// Package filegen is the orchestrator for standalone C++ file generation.
// It is invoked by the `cpp-gen generate` command to produce .hpp/.cpp pairs
// (or single-header variants) for classes, structs, free-function modules and
// interfaces, together with an optional companion test file.
package filegen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/matpdev/cpp-gen/internal/localconfig"
)

// ─────────────────────────────────────────────────────────────────────────────
// Public types
// ─────────────────────────────────────────────────────────────────────────────

// FileRequest describes what cpp-gen generate should produce.
type FileRequest struct {
	// Name is the identifier in any case convention; it will be normalised
	// internally.  Examples: "FooBar", "foo-bar", "foo_bar".
	Name string

	// Type selects the kind of C++ construct to generate (class | struct | free | interface).
	Type FileType

	// Brief is an optional short description written into the @brief header field.
	Brief string

	// OutputDir overrides both the include and src directories when set.
	// When empty, paths from LocalConfig.Paths are used.
	OutputDir string

	// NoTest skips test file generation regardless of the config setting.
	NoTest bool
}

// GeneratedFile records a single file that was handled by Generate.
type GeneratedFile struct {
	// Path is the absolute or project-relative path of the file.
	Path string

	// Skipped is true when the file already existed on disk and was not
	// overwritten.
	Skipped bool
}

// ─────────────────────────────────────────────────────────────────────────────
// Entry point
// ─────────────────────────────────────────────────────────────────────────────

// Generate creates C++ files based on req and the project's local config.
// It resolves output paths, renders file comment headers, fills templates and
// writes files to disk.  Returns the list of files that were generated (or
// skipped because they already existed).
func Generate(req FileRequest, cfg *localconfig.LocalConfig) ([]GeneratedFile, error) {
	// ── 1. Normalise name ────────────────────────────────────────────────────
	namePascal := toPascalCase(req.Name)
	nameSnake := toSnakeCase(req.Name)
	nameUpper := toUpperSnake(req.Name)

	// ── 2. Resolve output directories ───────────────────────────────────────
	includeDir := cfg.Paths.Include
	srcDir := cfg.Paths.Src
	if req.OutputDir != "" {
		includeDir = req.OutputDir
		srcDir = req.OutputDir
	}

	// ── 3. Build include path for .cpp ──────────────────────────────────────
	// When namespace is enabled the #include uses the project name as a prefix
	// directory (e.g. "myproject/foo_bar.hpp") so that consumers install the
	// header under include/myproject/.
	includePath := nameSnake + ".hpp"
	if cfg.Namespace.Enabled {
		prefix := cfg.Namespace.Name
		if prefix == "" {
			prefix = cfg.ProjectName
		}
		if prefix != "" {
			includePath = filepath.ToSlash(filepath.Join(prefix, nameSnake+".hpp"))
		}
	}

	// ── 4. Determine namespace ───────────────────────────────────────────────
	namespace := ""
	if cfg.Namespace.Enabled {
		namespace = cfg.Namespace.Name
		if namespace == "" {
			namespace = cfg.ProjectName
		}
	}

	// ── 5. Build HeaderMeta ──────────────────────────────────────────────────
	dateFormat := cfg.Header.DateFormat
	if dateFormat == "" {
		dateFormat = "2006-01-02"
	}

	baseMeta := HeaderMeta{
		Author:       cfg.Author,
		Organization: cfg.Organization,
		License:      cfg.License,
		ProjectName:  cfg.ProjectName,
	}

	// ── 6. Decide which files to generate ───────────────────────────────────
	type fileSpec struct {
		path     string
		tmplStr  string
		fileName string // used in the header @file field
		brief    string
	}

	var specs []fileSpec

	switch req.Type {
	case FileTypeClass:
		specs = []fileSpec{
			{
				path:     filepath.Join(includeDir, nameSnake+".hpp"),
				tmplStr:  tmplClassHpp,
				fileName: nameSnake + ".hpp",
				brief:    req.Brief,
			},
			{
				path:     filepath.Join(srcDir, nameSnake+".cpp"),
				tmplStr:  tmplClassCpp,
				fileName: nameSnake + ".cpp",
				brief:    req.Brief,
			},
		}

	case FileTypeStruct:
		specs = []fileSpec{
			{
				path:     filepath.Join(includeDir, nameSnake+".hpp"),
				tmplStr:  tmplStructHpp,
				fileName: nameSnake + ".hpp",
				brief:    req.Brief,
			},
		}

	case FileTypeFree:
		specs = []fileSpec{
			{
				path:     filepath.Join(includeDir, nameSnake+".hpp"),
				tmplStr:  tmplFreeHpp,
				fileName: nameSnake + ".hpp",
				brief:    req.Brief,
			},
			{
				path:     filepath.Join(srcDir, nameSnake+".cpp"),
				tmplStr:  tmplFreeCpp,
				fileName: nameSnake + ".cpp",
				brief:    req.Brief,
			},
		}

	case FileTypeInterface:
		specs = []fileSpec{
			{
				path:     filepath.Join(includeDir, nameSnake+".hpp"),
				tmplStr:  tmplInterfaceHpp,
				fileName: nameSnake + ".hpp",
				brief:    req.Brief,
			},
		}

	default:
		return nil, fmt.Errorf("unknown file type %q", req.Type)
	}

	// ── 7. Render and write each source file ─────────────────────────────────
	var results []GeneratedFile

	for _, s := range specs {
		meta := baseMeta
		meta.FileName = s.fileName
		meta.Brief = s.brief

		header := BuildHeader(cfg.Header.Style, cfg.Header.Fields, dateFormat, meta)

		data := FileTemplateData{
			Header:       header,
			Name:         namePascal,
			NameSnake:    nameSnake,
			NameUpper:    nameUpper,
			Namespace:    namespace,
			HasNamespace: cfg.Namespace.Enabled,
			IncludePath:  includePath,
			ProjectName:  cfg.ProjectName,
		}

		content, err := renderFileTemplate(s.tmplStr, data)
		if err != nil {
			return results, fmt.Errorf("rendering %s: %w", s.fileName, err)
		}

		gf, err := writeGeneratedFile(s.path, content)
		if err != nil {
			return results, fmt.Errorf("writing %s: %w", s.path, err)
		}
		results = append(results, gf)
	}

	// ── 8. Optional test file ────────────────────────────────────────────────
	if !req.NoTest && cfg.Generate.CreateTest {
		testDir := cfg.Paths.Tests
		if testDir == "" {
			testDir = "tests"
		}

		testFileName := nameSnake + "_test.cpp"
		testPath := filepath.Join(testDir, testFileName)
		testBrief := fmt.Sprintf("Tests for %s.", namePascal)

		testMeta := baseMeta
		testMeta.FileName = testFileName
		testMeta.Brief = testBrief

		header := BuildHeader(cfg.Header.Style, cfg.Header.Fields, dateFormat, testMeta)

		testData := FileTemplateData{
			Header:       header,
			Name:         namePascal,
			NameSnake:    nameSnake,
			NameUpper:    nameUpper,
			Namespace:    namespace,
			HasNamespace: cfg.Namespace.Enabled,
			IncludePath:  includePath,
			ProjectName:  cfg.ProjectName,
		}

		var testTmpl string
		switch cfg.Generate.TestFramework {
		case "gtest":
			testTmpl = tmplTestGTest
		case "catch2":
			testTmpl = tmplTestCatch2
		default:
			testTmpl = tmplTestNone
		}

		content, err := renderFileTemplate(testTmpl, testData)
		if err != nil {
			return results, fmt.Errorf("rendering test file: %w", err)
		}

		gf, err := writeGeneratedFile(testPath, content)
		if err != nil {
			return results, fmt.Errorf("writing test file: %w", err)
		}
		results = append(results, gf)
	}

	return results, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Private helpers
// ─────────────────────────────────────────────────────────────────────────────

// writeGeneratedFile writes content to path, creating all parent directories
// as needed.  If the file already exists it is not overwritten; instead the
// returned GeneratedFile has Skipped set to true.
func writeGeneratedFile(path, content string) (GeneratedFile, error) {
	if _, err := os.Stat(path); err == nil {
		return GeneratedFile{Path: path, Skipped: true}, nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return GeneratedFile{}, fmt.Errorf("creating directory %q: %w", filepath.Dir(path), err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return GeneratedFile{}, fmt.Errorf("writing file %q: %w", path, err)
	}

	return GeneratedFile{Path: path, Skipped: false}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Name-case helpers
// ─────────────────────────────────────────────────────────────────────────────

// toPascalCase converts a name in any of kebab-case, snake_case or PascalCase
// to PascalCase.
//
//	"foo-bar"  → "FooBar"
//	"foo_bar"  → "FooBar"
//	"fooBar"   → "FooBar"
//	"FooBar"   → "FooBar"
func toPascalCase(name string) string {
	if strings.ContainsAny(name, "-_ ") {
		parts := strings.FieldsFunc(name, func(r rune) bool {
			return r == '-' || r == '_' || r == '.' || unicode.IsSpace(r)
		})
		for i, p := range parts {
			if len(p) == 0 {
				continue
			}
			runes := []rune(p)
			runes[0] = unicode.ToUpper(runes[0])
			parts[i] = string(runes)
		}
		return strings.Join(parts, "")
	}

	if len(name) == 0 {
		return name
	}
	runes := []rune(name)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// toSnakeCase converts a name in any supported convention to snake_case.
//
//	"FooBar"  → "foo_bar"
//	"foo-bar" → "foo_bar"
//	"foo_bar" → "foo_bar"
func toSnakeCase(name string) string {
	replaced := strings.NewReplacer("-", "_", ".", "_", " ", "_").Replace(name)

	var sb strings.Builder
	runes := []rune(replaced)
	for i, r := range runes {
		if unicode.IsUpper(r) && i > 0 {
			prev := runes[i-1]
			if unicode.IsLower(prev) || unicode.IsDigit(prev) {
				sb.WriteRune('_')
			}
		}
		sb.WriteRune(unicode.ToLower(r))
	}

	return sb.String()
}

// toUpperSnake converts a name to UPPER_SNAKE_CASE.
//
//	"FooBar"  → "FOO_BAR"
//	"foo-bar" → "FOO_BAR"
func toUpperSnake(name string) string {
	return strings.ToUpper(toSnakeCase(name))
}

// ─────────────────────────────────────────────────────────────────────────────
// Test file templates
// ─────────────────────────────────────────────────────────────────────────────

const tmplTestCatch2 = `{{- if .Header}}{{.Header}}

{{end -}}
#include <catch2/catch_test_macros.hpp>
#include "{{.IncludePath}}"

TEST_CASE("{{.Name}} tests", "[{{.NameSnake}}]") {
    // TODO: write tests
}
`

const tmplTestGTest = `{{- if .Header}}{{.Header}}

{{end -}}
#include <gtest/gtest.h>
#include "{{.IncludePath}}"

TEST({{.Name}}Test, BasicTest) {
    // TODO: write tests
}
`

const tmplTestNone = `{{- if .Header}}{{.Header}}

{{end -}}
#include "{{.IncludePath}}"

// TODO: write tests for {{.Name}}
`
