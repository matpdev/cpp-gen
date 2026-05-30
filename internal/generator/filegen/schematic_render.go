package filegen

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/matpdev/cpp-gen/internal/localconfig"
)

// SchematicRequest describes what cpp-gen generate should produce
// when invoked with a schematic name.
type SchematicRequest struct {
	// Name is the user-supplied identifier (any case). Will be normalised.
	Name string

	// SchematicName is the schematic to use (e.g. "service", "repository").
	SchematicName string

	// Layer overrides cfg.Architecture.DefaultLayer for this invocation.
	// Empty means use the default layer from the config.
	Layer string

	// Brief is an optional short description for the header @brief.
	Brief string

	// NoTest skips test file generation even if the schematic includes one.
	NoTest bool
}

// SchematicListEntry holds display info for a single schematic.
type SchematicListEntry struct {
	Name        string
	Description string
	IsCustom    bool // true = defined in .cppgenrc.json
	FileCount   int
}

// GenerateSchematic resolves and executes a schematic, writing all files to disk.
// It merges built-in schematics with any custom schematics defined in cfg.Schematics,
// with custom entries taking precedence.
//
// Returns the list of generated (or skipped) files.
func GenerateSchematic(req SchematicRequest, cfg *localconfig.LocalConfig) ([]GeneratedFile, error) {
	// 1. Build the merged registry (built-ins + custom overrides)
	registry := buildMergedRegistry(cfg)

	// 2. Look up the schematic by name
	schematic, ok := registry.Get(req.SchematicName)
	if !ok {
		return nil, fmt.Errorf("schematic %q not found", req.SchematicName)
	}

	// 3. Resolve SchematicData from cfg + req
	data := buildSchematicData(req, cfg)

	baseMeta := HeaderMeta{
		Author:       cfg.Author,
		Organization: cfg.Organization,
		License:      cfg.License,
		ProjectName:  cfg.ProjectName,
	}

	var results []GeneratedFile

	// 4. For each SchematicFileSpec in the schematic
	for _, spec := range schematic.Files {
		// 4b. Skip test files when req.NoTest
		if req.NoTest && isTestSpec(spec) {
			continue
		}

		// Compute the role-specific class name for this file.
		// ClassName = ClassPrefix + NamePascal + ClassSuffix
		// e.g. prefix="I", pascal="User", suffix="Service" → "IUserService"
		className := spec.ClassPrefix + data.NamePascal + spec.ClassSuffix
		if className == "" {
			className = data.NamePascal
		}
		data.ClassName = className

		// 4a. Resolve path via ResolvePath (ClassName is now available in data)
		resolvedPath, err := ResolvePath(spec.PathTmpl, data)
		if err != nil {
			return results, fmt.Errorf("resolving path for role %q: %w", spec.Role, err)
		}

		fileName := filepath.Base(resolvedPath)

		// 4c. Build HeaderMeta
		brief := spec.Brief
		if brief == "" {
			brief = req.Brief
		}

		meta := baseMeta
		meta.FileName = fileName
		meta.Brief = brief

		// 4d. Render header
		header := BuildHeader(cfg.Header.Style, cfg.Header.Fields, cfg.Header.DateFormat, meta)

		// 4e. Build FileTemplateData — use ClassName as the C++ class name
		// so each file gets the role-specific name (e.g. IUserService, UserService).
		tmplData := FileTemplateData{
			Header:       header,
			Name:         className,
			NameSnake:    data.NameSnake,
			NameUpper:    data.NameUpper,
			Namespace:    data.Namespace,
			HasNamespace: cfg.Namespace.Enabled && data.Namespace != "",
			IncludePath:  filepath.ToSlash(filepath.Join(data.Namespace, data.NameSnake+".hpp")),
			ProjectName:  data.ProjectName,
		}

		// 4e. Render content template
		content, err := resolveContentTemplate(spec.ContentTmpl, tmplData)
		if err != nil {
			return results, fmt.Errorf("rendering content for role %q: %w", spec.Role, err)
		}

		// 4f. Write file
		gf, err := writeGeneratedFile(resolvedPath, content)
		if err != nil {
			return results, fmt.Errorf("writing file %q: %w", resolvedPath, err)
		}

		results = append(results, gf)
	}

	return results, nil
}

// ListSchematics returns all schematics available for a given config,
// merging built-ins with custom schematics. Custom entries are marked.
func ListSchematics(cfg *localconfig.LocalConfig) []SchematicListEntry {
	builtinNames := make(map[string]bool)
	for _, s := range BuiltinRegistry().All() {
		builtinNames[s.Name] = true
	}

	registry := buildMergedRegistry(cfg)

	var entries []SchematicListEntry
	for _, s := range registry.All() {
		entries = append(entries, SchematicListEntry{
			Name:        s.Name,
			Description: s.Description,
			IsCustom:    !builtinNames[s.Name],
			FileCount:   len(s.Files),
		})
	}
	return entries
}

// buildMergedRegistry returns a registry with built-ins plus custom overrides from cfg.
func buildMergedRegistry(cfg *localconfig.LocalConfig) *SchematicRegistry {
	r := NewRegistry()
	// register all built-ins first
	for _, s := range BuiltinRegistry().All() {
		r.Register(s)
	}
	// then override/add custom schematics from cfg
	for name, cs := range cfg.Schematics {
		files := make([]SchematicFileSpec, len(cs.Files))
		for i, f := range cs.Files {
			files[i] = SchematicFileSpec{
				Role:        f.Role,
				PathTmpl:    f.Path,
				ContentTmpl: f.Template,
				Brief:       f.Brief,
			}
		}
		r.Register(Schematic{
			Name:        name,
			Description: cs.Description,
			Files:       files,
		})
	}
	return r
}

// buildSchematicData derives SchematicData from the request and config.
func buildSchematicData(req SchematicRequest, cfg *localconfig.LocalConfig) SchematicData {
	nameSnake := toSnakeCase(req.Name)
	namePascal := toPascalCase(req.Name)
	nameUpper := toUpperSnake(req.Name)

	layer := req.Layer
	if layer == "" {
		layer = cfg.Architecture.DefaultLayer
	}

	namespace := cfg.Namespace.Name
	if namespace == "" {
		namespace = cfg.ProjectName
	}

	include := cfg.Paths.Include
	if cfg.Architecture.HeaderOnly {
		include = cfg.Paths.Src
	}

	return SchematicData{
		NameSnake:   nameSnake,
		NamePascal:  namePascal,
		NameUpper:   nameUpper,
		Layer:       layer,
		Namespace:   namespace,
		ProjectName: cfg.ProjectName,
		Src:         cfg.Paths.Src,
		Include:     include,
		Tests:       cfg.Paths.Tests,
	}
}

// isTestSpec reports whether a file spec represents a test file.
func isTestSpec(spec SchematicFileSpec) bool {
	return spec.Role == "test" || strings.HasPrefix(spec.ContentTmpl, "test")
}

// resolveContentTemplate maps a ContentTmpl key to the actual Go template string
// and renders it with the given FileTemplateData.
func resolveContentTemplate(key string, data FileTemplateData) (string, error) {
	var tmpl string
	switch key {
	case "class_hpp":
		tmpl = tmplClassHpp
	case "class_cpp":
		tmpl = tmplClassCpp
	case "struct_hpp":
		tmpl = tmplStructHpp
	case "interface_hpp":
		tmpl = tmplInterfaceHpp
	case "free_hpp":
		tmpl = tmplFreeHpp
	case "free_cpp":
		tmpl = tmplFreeCpp
	case "test_catch2":
		tmpl = tmplTestCatch2
	case "test_gtest":
		tmpl = tmplTestGTest
	case "test_none", "test":
		// No cfg available here; default to catch2
		tmpl = tmplTestCatch2
	default:
		return "", fmt.Errorf("unknown content template %q", key)
	}
	return renderFileTemplate(tmpl, data)
}
