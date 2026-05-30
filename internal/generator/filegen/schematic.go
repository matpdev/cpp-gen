package filegen

import (
	"bytes"
	"text/template"
)

// SchematicFileSpec describes one file within a schematic.
type SchematicFileSpec struct {
	// Role is a semantic label: "interface", "impl_header", "impl", "test", "types", etc.
	Role string

	// PathTmpl is a Go text/template string for the output path.
	// Available variables (via SchematicData):
	//   .NameSnake   .NamePascal   .NameUpper   .ClassName
	//   .Layer       .Namespace    .ProjectName
	//   .Src         .Include      .Tests
	PathTmpl string

	// ContentTmpl is the name of a built-in content template key:
	//   "class_hpp" | "class_cpp" | "struct_hpp" | "interface_hpp" |
	//   "free_hpp"  | "free_cpp"  | "test_catch2" | "test_gtest" | "test_none"
	ContentTmpl string

	// Brief is an optional fixed @brief for this specific file.
	// If empty, SchematicRequest.Brief is used.
	Brief string

	// ClassSuffix is appended to NamePascal to form the C++ class name for
	// this specific file. For example, with NamePascal="User" and
	// ClassSuffix="Service", the generated class will be named "UserService".
	// The interface prefix "I" is handled by ClassPrefix.
	// Leave empty to use NamePascal as-is.
	ClassSuffix string

	// ClassPrefix is prepended to NamePascal (after any suffix) to form the
	// class name. Typically "I" for interface files (e.g. "IUserService").
	// Leave empty for no prefix.
	ClassPrefix string
}

// Schematic groups a set of files that belong to a single architectural concept.
type Schematic struct {
	// Name is the schematic identifier used on the CLI (e.g. "service").
	Name string

	// Description is shown in cpp-gen generate --list.
	Description string

	// Files is the ordered list of file specs this schematic produces.
	Files []SchematicFileSpec
}

// SchematicData holds the template variables available in PathTmpl strings
// and content templates. ClassName is the only field that varies per
// SchematicFileSpec; all others are constant for the whole schematic invocation.
type SchematicData struct {
	NameSnake   string
	NamePascal  string
	NameUpper   string
	Layer       string // active architectural layer, empty when not applicable
	Namespace   string
	ProjectName string
	Src         string // from cfg.Paths.Src
	Include     string // from cfg.Paths.Include; equals Src when HeaderOnly
	Tests       string // from cfg.Paths.Tests

	// ClassName is the fully-qualified C++ class name for the current file.
	// It is recomputed for each SchematicFileSpec using ClassPrefix + NamePascal + ClassSuffix.
	// Example: prefix="I", pascal="User", suffix="Service" → "IUserService".
	// Templates can use {{.ClassName}} to get the role-specific class name.
	ClassName string
}

// SchematicRegistry is an ordered, named collection of schematics.
// Built-in schematics are registered at init time; custom schematics
// (from .cppgenrc.json) are merged at runtime and take precedence.
type SchematicRegistry struct {
	entries map[string]Schematic
	order   []string // insertion order, for --list display
}

// NewRegistry returns an empty SchematicRegistry.
func NewRegistry() *SchematicRegistry {
	return &SchematicRegistry{
		entries: make(map[string]Schematic),
	}
}

// Register adds or replaces a schematic. The order slice preserves insertion
// order so that --list output is deterministic.
func (r *SchematicRegistry) Register(s Schematic) {
	if _, exists := r.entries[s.Name]; !exists {
		r.order = append(r.order, s.Name)
	}
	r.entries[s.Name] = s
}

// Get returns the schematic with the given name, or false if not found.
func (r *SchematicRegistry) Get(name string) (Schematic, bool) {
	s, ok := r.entries[name]
	return s, ok
}

// All returns all schematics in registration order.
func (r *SchematicRegistry) All() []Schematic {
	out := make([]Schematic, 0, len(r.order))
	for _, name := range r.order {
		out = append(out, r.entries[name])
	}
	return out
}

// Names returns the names of all registered schematics in order.
func (r *SchematicRegistry) Names() []string {
	names := make([]string, len(r.order))
	copy(names, r.order)
	return names
}

// ResolvePath evaluates a PathTmpl string with the given SchematicData.
// Returns the resolved path string or an error.
func ResolvePath(pathTmpl string, data SchematicData) (string, error) {
	tmpl, err := template.New("path").Parse(pathTmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
