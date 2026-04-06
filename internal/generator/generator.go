// Package generator contains all C++ project generation logic for cpp-gen.
//
// The package is organized into specialized sub-modules:
//   - cmake.go      — Generation of CMakeLists.txt, CMakePresets.json and cmake/ helpers
//   - structure.go  — Creation of folder structure and initial source files
//   - git.go        — Git repository initialization and .gitignore / README generation
//   - clang.go      — Generation of .clangd and .clang-format
//   - ide/          — IDE-specific configurations (VSCode, CLion, Neovim)
//   - packages/     — Integration with VCPKG and FetchContent
//   - layout/       — Folder structure specifications (Separate, Merged, Flat, Modular, Two-Root)
//
// The public entry point is Generator, created via New() and executed via Generate().
package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
	"unicode"

	"github.com/matpdev/cpp-gen/internal/config"
	"github.com/matpdev/cpp-gen/internal/generator/ide"
	"github.com/matpdev/cpp-gen/internal/generator/layout"
	"github.com/matpdev/cpp-gen/internal/generator/packages"
)

// ─────────────────────────────────────────────────────────────────────────────
// TemplateData — data passed to all generation templates
// ─────────────────────────────────────────────────────────────────────────────

// TemplateData centralizes all variables available in Go templates
// (text/template) used to generate project files.
//
// In addition to the direct ProjectConfig fields, it includes derived name forms
// (NameUpper, NameSnake, NamePascal) and boolean flags to simplify
// conditional logic within templates ({{if .IsVSCode}}, etc.).
type TemplateData struct {
	// ── Metadata ───────────────────────────────────────────────────────────────

	// Name is the original project name, exactly as typed (e.g. "my-project").
	Name string

	// NameUpper is the name in UPPER_SNAKE_CASE, used in CMake variables and include guards.
	// Example: "my-project" → "MY_PROJECT"
	NameUpper string

	// NameSnake is the name in snake_case, used in file names and C++ functions.
	// Example: "my-project" → "my_project"
	NameSnake string

	// NamePascal is the name in PascalCase, used in C++ class and namespace names.
	// Example: "my-project" → "MyProject"
	NamePascal string

	// Description is the project description provided by the user.
	Description string

	// Author is the name of the author or organization.
	Author string

	// Version is the initial version in SemVer format (e.g. "1.0.0").
	Version string

	// Year is the current year, used in copyright headers and README.
	Year string

	// ── Technical configuration ────────────────────────────────────────────────

	// Standard is the C++ standard as a numeric string (e.g. "20").
	Standard string

	// ── Boolean flags for project type ────────────────────────────────────────
	// Derived from config.ProjectType to simplify templates.

	IsExecutable bool // true if TypeExecutable
	IsStaticLib  bool // true if TypeStaticLib
	IsHeaderOnly bool // true if TypeHeaderOnly

	// ── Boolean flags for package manager ─────────────────────────────────────

	UseVCPKG        bool // true if PkgVCPKG
	UseFetchContent bool // true if PkgFetchContent

	// ── Boolean IDE flags ─────────────────────────────────────────────────────

	IsVSCode bool // true if IDEVSCode
	IsCLion  bool // true if IDECLion
	IsNvim   bool // true if IDENvim
	IsZed    bool // true if IDEZed

	// ── Optional tool flags ───────────────────────────────────────────────────

	UseGit           bool   // initialize Git repository
	UseClangd        bool   // generate .clangd
	UseClangFormat   bool   // generate .clang-format
	ClangFormatStyle string // base style for .clang-format (e.g. "LLVM", "Google")

	// ── Folder layout ─────────────────────────────────────────────────────────
	// Derived from the layout.Spec calculated in buildTemplateData().

	// Layout is the identifier of the chosen folder pattern (e.g. "separate").
	Layout string

	// LayoutNote is an explanatory comment inserted in the root CMakeLists.txt
	// describing the adopted layout convention.
	LayoutNote string

	// LayoutCMakeSubdir is the add_subdirectory() argument in the root CMakeLists.txt
	// for the main target (e.g. "src", "mylib", "libs/mylib").
	LayoutCMakeSubdir string

	// LayoutCMakeIncludeBlock is the pre-formatted target_include_directories() block
	// for the main target, ready for direct insertion into the src/CMakeLists.txt template.
	LayoutCMakeIncludeBlock string

	// LayoutCMakeTestIncludeBlock is the include dirs block for the test target.
	LayoutCMakeTestIncludeBlock string

	// LayoutIncludePrefix is the directory prefix used in #include of generated C++ files
	// (e.g. "mylib/" → #include "mylib/file.hpp" ; "" → #include "file.hpp").
	LayoutIncludePrefix string

	// LayoutIsModular indicates that the layout uses libs/<name>/ with an executable in apps/.
	// When true, cmake.go adds the executable target directly in the root CMakeLists.txt.
	LayoutIsModular bool

	// LayoutModularLibDir is the relative path of the library in the modular layout
	// (e.g. "libs/mylib"). Empty in all other layouts.
	LayoutModularLibDir string

	// LayoutSpec is a reference to the complete Spec, available to structure and CMake
	// generators that need the file paths.
	LayoutSpec *layout.Spec
}

// ─────────────────────────────────────────────────────────────────────────────
// Generator
// ─────────────────────────────────────────────────────────────────────────────

// Generator is the main struct that coordinates the generation of all
// artifacts for a C++ project. Must be created with New() and executed with Generate().
type Generator struct {
	// cfg contains the original configuration provided by the user.
	cfg *config.ProjectConfig

	// data is the data derived from cfg, ready for use in templates.
	data *TemplateData

	// spec is the resolved folder layout for this project.
	spec *layout.Spec

	// root is the absolute path of the root directory of the project to be created.
	root string

	// verbose enables the display of each generated file during the process.
	verbose bool

	// steps accumulates the log entries for each step for final display.
	steps []stepResult
}

// stepResult represents the result of a generation step.
type stepResult struct {
	label   string // short step description (e.g. "Folder structure")
	success bool   // true if completed without errors
	err     error  // error that occurred, nil if success == true
}

// ─────────────────────────────────────────────────────────────────────────────
// Constructor
// ─────────────────────────────────────────────────────────────────────────────

// New creates a new Generator from the ProjectConfig and the verbose flag.
//
// Automatically derives the TemplateData (name forms, boolean flags, etc.),
// resolves the folder layout and calculates the root path of the project to be generated.
func New(cfg *config.ProjectConfig, verbose bool) *Generator {
	nameSnake := toSnakeCase(cfg.Name)
	spec := layout.Resolve(cfg.Name, nameSnake, cfg.Layout, cfg.ProjectType)
	data := buildTemplateData(cfg, spec)
	root := cfg.ProjectPath()

	return &Generator{
		cfg:     cfg,
		data:    data,
		spec:    spec,
		root:    root,
		verbose: verbose,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Generate — main orchestrator
// ─────────────────────────────────────────────────────────────────────────────

// Generate executes all project generation steps in the correct order:
//
//  1. Creates the folder structure and source files
//  2. Generates the CMake files (CMakeLists.txt, presets, helpers)
//  3. Configures the package manager (VCPKG or FetchContent)
//  4. Generates the chosen IDE configurations
//  5. Generates .clangd and/or .clang-format
//  6. Initializes the Git repository and generates .gitignore / README
//
// At the end, prints a report of all executed steps.
// If any critical step fails, generation is immediately interrupted.
func (g *Generator) Generate() error {
	fmt.Printf("\n  Gerando projeto %q em %q...\n\n", g.cfg.Name, g.root)

	// Steps are executed in sequence; each one records its result.
	pipeline := []struct {
		label string
		fn    func() error
	}{
		{"Estrutura de pastas e arquivos fonte", g.runStructure},
		{"Arquivos CMake", g.runCMake},
		{"Gerenciador de pacotes", g.runPackages},
		{"Configuração da IDE", g.runIDE},
		{"Ferramentas Clang", g.runClang},
		{"Git e metadados do repositório", g.runGit},
	}

	for _, step := range pipeline {
		err := step.fn()
		g.steps = append(g.steps, stepResult{
			label:   step.label,
			success: err == nil,
			err:     err,
		})

		if err != nil {
			g.printStepReport()
			return fmt.Errorf("falha em %q: %w", step.label, err)
		}
	}

	g.printStepReport()
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Pipeline steps
// ─────────────────────────────────────────────────────────────────────────────

// runStructure creates the directory hierarchy and initial C++ source files.
func (g *Generator) runStructure() error {
	return generateStructure(g.root, g.data, g.spec, g.verbose)
}

// runCMake generates all CMake files for the project:
//   - root CMakeLists.txt
//   - src/CMakeLists.txt
//   - tests/CMakeLists.txt
//   - cmake/CompilerWarnings.cmake
//   - CMakePresets.json
func (g *Generator) runCMake() error {
	return generateCMake(g.root, g.data, g.verbose)
}

// runPackages configures the chosen package manager.
// If none was selected (PkgNone), the step is silently skipped.
func (g *Generator) runPackages() error {
	switch g.cfg.PackageManager {
	case config.PkgVCPKG:
		return packages.GenerateVCPKG(g.root, g.verbose)
	case config.PkgFetchContent:
		return packages.GenerateFetchContent(g.root, g.verbose)
	default:
		// No manager selected — nothing to do.
		return nil
	}
}

// runIDE generates the configurations specific to the chosen IDE.
// If IDENone was selected, the step is silently skipped.
func (g *Generator) runIDE() error {
	ideData := &ide.Data{
		ProjectName:  g.data.Name,
		NameUpper:    g.data.NameUpper,
		IsExecutable: g.data.IsExecutable,
		UseVCPKG:     g.data.UseVCPKG,
	}

	switch g.cfg.IDE {
	case config.IDEVSCode:
		return ide.GenerateVSCode(g.root, ideData, g.verbose)
	case config.IDECLion:
		return ide.GenerateCLion(g.root, ideData, g.verbose)
	case config.IDENvim:
		return ide.GenerateNvim(g.root, ideData, g.verbose)
	case config.IDEZed:
		return ide.GenerateZed(g.root, ideData, g.verbose)
	default:
		return nil
	}
}

// runClang generates the configuration files for Clang tools:
//   - .clangd  (if UseClangd == true)
//   - .clang-format (if UseClangFormat == true)
func (g *Generator) runClang() error {
	return generateClang(g.root, g.data, g.verbose)
}

// runGit initializes the Git repository, generates .gitignore and README.md.
// If UseGit == false, only the README is created (without git init).
func (g *Generator) runGit() error {
	return generateGit(g.root, g.data, g.verbose)
}

// ─────────────────────────────────────────────────────────────────────────────
// Step report
// ─────────────────────────────────────────────────────────────────────────────

// printStepReport prints to standard output a summary of all executed steps,
// indicating success or failure with visual icons.
func (g *Generator) printStepReport() {
	checkOK := "  ✓"
	checkFail := "  ✗"

	for _, s := range g.steps {
		if s.success {
			fmt.Printf("%s  %s\n", checkOK, s.label)
		} else {
			fmt.Printf("%s  %s — %v\n", checkFail, s.label, s.err)
		}
	}
	fmt.Println()
}

// ─────────────────────────────────────────────────────────────────────────────
// buildTemplateData — template data derivation
// ─────────────────────────────────────────────────────────────────────────────

// buildTemplateData converts a ProjectConfig + layout.Spec into TemplateData,
// calculating all derived name forms, boolean flags and layout fields
// needed for the conditional logic in templates.
func buildTemplateData(cfg *config.ProjectConfig, spec *layout.Spec) *TemplateData {
	return &TemplateData{
		// Name forms
		Name:       cfg.Name,
		NameUpper:  toUpperSnake(cfg.Name),
		NameSnake:  toSnakeCase(cfg.Name),
		NamePascal: toPascalCase(cfg.Name),

		// Metadata
		Description: cfg.Description,
		Author:      cfg.Author,
		Version:     cfg.Version,
		Year:        fmt.Sprintf("%d", time.Now().Year()),

		// Technical
		Standard: string(cfg.Standard),

		// Project type
		IsExecutable: cfg.ProjectType == config.TypeExecutable,
		IsStaticLib:  cfg.ProjectType == config.TypeStaticLib,
		IsHeaderOnly: cfg.ProjectType == config.TypeHeaderOnly,

		// Package managers
		UseVCPKG:        cfg.PackageManager == config.PkgVCPKG,
		UseFetchContent: cfg.PackageManager == config.PkgFetchContent,

		// IDEs
		IsVSCode: cfg.IDE == config.IDEVSCode,
		IsCLion:  cfg.IDE == config.IDECLion,
		IsNvim:   cfg.IDE == config.IDENvim,
		IsZed:    cfg.IDE == config.IDEZed,

		// Optional tools
		UseGit:           cfg.UseGit,
		UseClangd:        cfg.UseClangd,
		UseClangFormat:   cfg.UseClangFormat,
		ClangFormatStyle: string(cfg.ClangFormatStyle),

		// Folder layout — derived from the resolved layout.Spec
		Layout:                      string(spec.Kind),
		LayoutNote:                  spec.LayoutNote,
		LayoutCMakeSubdir:           spec.CMakeSubdir,
		LayoutCMakeIncludeBlock:     spec.CMakeIncludeBlock,
		LayoutCMakeTestIncludeBlock: spec.CMakeTestIncludeBlock,
		LayoutIncludePrefix:         spec.IncludePrefix,
		LayoutIsModular:             spec.Kind == config.LayoutModular,
		LayoutModularLibDir:         spec.CMakeModularLibDir,
		LayoutSpec:                  spec,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Name transformation utilities
// ─────────────────────────────────────────────────────────────────────────────

// toUpperSnake converts a project name to UPPER_SNAKE_CASE.
//
// Examples:
//
//	"my-project"   → "MY_PROJECT"
//	"my.lib.core"  → "MY_LIB_CORE"
func toUpperSnake(name string) string {
	replacer := strings.NewReplacer("-", "_", ".", "_", " ", "_")
	return strings.ToUpper(replacer.Replace(name))
}

// toSnakeCase converts a project name to snake_case.
//
// Examples:
//
//	"my-project" → "my_project"
//	"My-Lib"     → "my_lib"
func toSnakeCase(name string) string {
	replacer := strings.NewReplacer("-", "_", ".", "_", " ", "_")
	return strings.ToLower(replacer.Replace(name))
}

// toPascalCase converts a project name to PascalCase.
// Recognized delimiters: hyphen, underscore, dot and space.
//
// Examples:
//
//	"my-project"   → "MyProject"
//	"my_lib_core"  → "MyLibCore"
func toPascalCase(name string) string {
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

// ─────────────────────────────────────────────────────────────────────────────
// I/O utilities shared among sub-generators
// ─────────────────────────────────────────────────────────────────────────────

// writeFile creates (or overwrites) a file at the given path with the provided
// content. Automatically creates all necessary parent directories.
// If verbose is true, prints the path of the created file.
func writeFile(path, content string, verbose bool) error {
	// Ensures the parent directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("criar diretório %q: %w", filepath.Dir(path), err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("escrever arquivo %q: %w", path, err)
	}

	if verbose {
		fmt.Printf("    + %s\n", path)
	}

	return nil
}

// renderTemplate processes a Go template (text/template) with the provided data
// and returns the result as a string. Returns an error if the template is invalid
// or if the data does not satisfy the referenced fields.
func renderTemplate(name, tmpl string, data any) (string, error) {
	t, err := template.New(name).Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("parse do template %q: %w", name, err)
	}

	var sb strings.Builder
	if err := t.Execute(&sb, data); err != nil {
		return "", fmt.Errorf("execução do template %q: %w", name, err)
	}

	return sb.String(), nil
}

// writeTemplate is a combination of renderTemplate + writeFile:
// processes the template and writes the result to the indicated file.
func writeTemplate(path, name, tmpl string, data any, verbose bool) error {
	content, err := renderTemplate(name, tmpl, data)
	if err != nil {
		return err
	}
	return writeFile(path, content, verbose)
}
