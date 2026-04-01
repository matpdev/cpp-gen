// Package layout defines the folder structure specifications for C++ projects.
//
// Exported kind constants for use by other packages without importing config:
//
//	layout.KindSeparate, layout.KindMerged, layout.KindFlat, layout.KindModular, layout.KindTwoRoot
//
// Each layout (organization pattern) determines:
//   - Which directories to create
//   - Where source files, headers and tests are located
//   - Which include directories CMake uses in target_include_directories()
//   - The prefix used in #include directives of generated C++ files
//
// Available layouts:
//
//	separate  — include/<name>/ + src/ (classic CMake)
//	merged    — <name>/ with headers and sources together (Pitchfork / P1204R0)
//	flat      — everything in src/ without separation
//	modular   — libs/<name>/ for multi-modules (Pitchfork libs/)
//	two-root  — include/ + src/ without a namespace subdirectory
package layout

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/matpdev/cpp-gen/internal/config"
)

// ─────────────────────────────────────────────────────────────────────────────
// Exported kind constants
// ─────────────────────────────────────────────────────────────────────────────
// Aliases of config.FolderLayout constants exposed directly by the layout
// package so other packages (e.g.: structure.go) can compare spec.Kind
// without needing to import the config package separately.

const (
	KindSeparate = config.LayoutSeparate
	KindMerged   = config.LayoutMerged
	KindFlat     = config.LayoutFlat
	KindModular  = config.LayoutModular
	KindTwoRoot  = config.LayoutTwoRoot
)

// ─────────────────────────────────────────────────────────────────────────────
// Spec — complete layout specification
// ─────────────────────────────────────────────────────────────────────────────

// Spec describes all parameterizable aspects of a C++ folder layout.
// It is calculated once per project in Resolve() and used by all generators
// (structure.go, cmake.go) to know where to create files and how to configure CMake.
type Spec struct {
	// Kind is the layout identifier (e.g.: "separate", "merged").
	Kind config.FolderLayout

	// ── Directories to create ────────────────────────────────────────────────────

	// Dirs lists all directories to create (relative to the project root).
	// Includes all necessary intermediates — generators use MkdirAll.
	Dirs []string

	// ── Generated file paths ──────────────────────────────────────────────────
	// All paths are relative to the project root.

	// MainCPP is the path to the executable entry file (main.cpp).
	MainCPP string

	// LibCPP is the path to the library implementation file.
	LibCPP string

	// PublicHPP is the path to the library's main public header.
	PublicHPP string

	// TestCPP is the path to the main test file.
	TestCPP string

	// ── CMake configuration ───────────────────────────────────────────────────

	// CMakeSubdir is the subdirectory passed to add_subdirectory() in the
	// root CMakeLists.txt for the main target (e.g.: "src", "mylib", "libs/mylib").
	CMakeSubdir string

	// CMakeTestsDir is the tests subdirectory (almost always "tests").
	CMakeTestsDir string

	// CMakeIncludeBlock is the pre-formatted block of target_include_directories()
	// for the main target, ready for direct insertion into the CMake template.
	// Example:
	//     PUBLIC
	//         $<BUILD_INTERFACE:${PROJECT_SOURCE_DIR}/include>
	//         $<INSTALL_INTERFACE:include>
	//     PRIVATE
	//         ${CMAKE_CURRENT_SOURCE_DIR}
	CMakeIncludeBlock string

	// CMakeTestIncludeBlock is the include dirs block for the test target.
	CMakeTestIncludeBlock string

	// CMakeModularLibDir is the library directory in the modular layout.
	// Populated only when Kind == LayoutModular; empty otherwise.
	CMakeModularLibDir string

	// ── Include C++ ───────────────────────────────────────────────────────────

	// IncludePrefix is the directory prefix used in the #include directives
	// of generated C++ files.
	//
	// Examples:
	//   "mylib/"   → #include "mylib/mylib.hpp"   (separate, merged, modular)
	//   ""         → #include "mylib.hpp"          (flat, two-root)
	IncludePrefix string

	// ── Documentation notes ───────────────────────────────────────────────────

	// LayoutNote is a comment inserted in the generated CMakeLists.txt explaining
	// briefly the layout convention adopted and where to find the documentation.
	LayoutNote string
}

// ─────────────────────────────────────────────────────────────────────────────
// Resolve — main factory
// ─────────────────────────────────────────────────────────────────────────────

// Resolve calculates and returns the complete Spec for the combination of project
// name, project type and chosen layout.
//
// All returned file paths are relative to the project root.
// CMake blocks are pre-formatted and ready for insertion into templates.
//
// Parameters:
//
//	name        — project name (e.g.: "minha-lib")
//	nameSnake   — name in snake_case   (e.g.: "minha_lib")
//	layout      — chosen FolderLayout constant
//	projectType — artifact type (executable, static-lib, header-only)
func Resolve(
	name string,
	nameSnake string,
	layout config.FolderLayout,
	projectType config.ProjectType,
) *Spec {
	switch layout {
	case config.LayoutMerged:
		return resolveMerged(name, nameSnake, projectType)
	case config.LayoutFlat:
		return resolveFlat(name, nameSnake, projectType)
	case config.LayoutModular:
		return resolveModular(name, nameSnake, projectType)
	case config.LayoutTwoRoot:
		return resolveTwoRoot(name, nameSnake, projectType)
	default: // config.LayoutSeparate — padrão e fallback
		return resolveSeparate(name, nameSnake, projectType)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Individual layouts
// ─────────────────────────────────────────────────────────────────────────────

// resolveSeparate returns the spec for the classic layout with include/<name>/ and src/ separation.
//
// Generated structure:
//
//	include/<name>/  ← public headers with namespace prefix
//	src/             ← implementations and private headers
//	tests/           ← unit and integration tests
//	cmake/           ← auxiliary CMake modules
//	docs/            ← documentation
//
// Example include: #include "<name>/file.hpp"
// CMake: PUBLIC include, PRIVATE src
func resolveSeparate(name, nameSnake string, pt config.ProjectType) *Spec {
	isHeaderOnly := pt == config.TypeHeaderOnly

	dirs := []string{
		filepath.Join("include", name),
		"cmake",
		"tests",
		"docs",
	}
	if !isHeaderOnly {
		dirs = append(dirs, "src")
	}

	mainCPP := filepath.Join("src", "main.cpp")
	libCPP := filepath.Join("src", nameSnake+".cpp")
	publicHPP := filepath.Join("include", name, nameSnake+".hpp")
	testCPP := filepath.Join("tests", "test_main.cpp")

	includeBlock := buildIncludeBlock(
		[]string{
			"$<BUILD_INTERFACE:${PROJECT_SOURCE_DIR}/include>",
			"$<INSTALL_INTERFACE:include>",
		},
		[]string{
			"${CMAKE_CURRENT_SOURCE_DIR}",
		},
		pt,
	)

	testIncludeBlock := indent(4, "${PROJECT_SOURCE_DIR}/include")

	return &Spec{
		Kind:                  config.LayoutSeparate,
		Dirs:                  dirs,
		MainCPP:               mainCPP,
		LibCPP:                libCPP,
		PublicHPP:             publicHPP,
		TestCPP:               testCPP,
		CMakeSubdir:           "src",
		CMakeTestsDir:         "tests",
		CMakeIncludeBlock:     includeBlock,
		CMakeTestIncludeBlock: testIncludeBlock,
		IncludePrefix:         name + "/",
		LayoutNote: fmt.Sprintf(
			"# Layout: Separate (clássico)\n"+
				"# Headers públicos em include/%s/, implementações em src/.\n"+
				"# Referência: padrão amplamente adotado em projetos CMake modernos.",
			name,
		),
	}
}

// resolveMerged returns the spec for the Pitchfork / P1204R0 layout (merged placement).
//
// Headers and implementations are in the same <name>/ directory, eliminating the
// need to navigate between include/ and src/. Unit tests remain as sibling
// files (*.<name>.test.cpp). Integration tests are in tests/.
//
// Generated structure:
//
//	<name>/           ← headers (.hpp) and sources (.cpp) together
//	<name>/*.test.cpp ← unit tests alongside modules (generated if library)
//	tests/            ← integration and functional tests
//	cmake/            ← auxiliary CMake modules
//	docs/             ← documentation
//
// Example include: #include "<name>/file.hpp"
// CMake: PUBLIC ${PROJECT_SOURCE_DIR} (root as include base)
func resolveMerged(name, nameSnake string, pt config.ProjectType) *Spec {
	dirs := []string{
		name,
		"tests",
		"cmake",
		"docs",
	}

	mainCPP := filepath.Join(name, "main.cpp")
	libCPP := filepath.Join(name, nameSnake+".cpp")
	publicHPP := filepath.Join(name, nameSnake+".hpp")
	testCPP := filepath.Join("tests", "driver.cpp")

	// In the merged layout, the CMake subdir is the project's own directory
	cmakeSubdir := name

	includeBlock := buildIncludeBlock(
		[]string{
			"$<BUILD_INTERFACE:${PROJECT_SOURCE_DIR}>",
			"$<INSTALL_INTERFACE:.>",
		},
		nil,
		pt,
	)

	testIncludeBlock := indent(4, "${PROJECT_SOURCE_DIR}")

	return &Spec{
		Kind:                  config.LayoutMerged,
		Dirs:                  dirs,
		MainCPP:               mainCPP,
		LibCPP:                libCPP,
		PublicHPP:             publicHPP,
		TestCPP:               testCPP,
		CMakeSubdir:           cmakeSubdir,
		CMakeTestsDir:         "tests",
		CMakeIncludeBlock:     includeBlock,
		CMakeTestIncludeBlock: testIncludeBlock,
		IncludePrefix:         name + "/",
		LayoutNote: fmt.Sprintf(
			"# Layout: Merged (Pitchfork / P1204R0)\n"+
				"# Headers e implementações juntos em %s/. Unit tests como arquivos irmãos.\n"+
				"# Referência: https://vector-of-bool.github.io/pitchfork (SG15 P1204R0)",
			name,
		),
	}
}

// resolveFlat returns the spec for the flat layout (everything in src/).
//
// Headers and implementations are together in src/ without any separation.
// It is the simplest layout, suitable for executables and small projects
// where the distinction between public and private API is not relevant.
//
// Generated structure:
//
//	src/   ← headers (.hpp) and sources (.cpp) in the same place
//	tests/ ← tests
//	cmake/ ← auxiliary CMake modules
//	docs/  ← documentation
//
// Example include: #include "file.hpp" (without namespace prefix)
// CMake: PRIVATE src
func resolveFlat(name, nameSnake string, pt config.ProjectType) *Spec {
	dirs := []string{
		"src",
		"tests",
		"cmake",
		"docs",
	}

	mainCPP := filepath.Join("src", "main.cpp")
	libCPP := filepath.Join("src", nameSnake+".cpp")
	publicHPP := filepath.Join("src", nameSnake+".hpp")
	testCPP := filepath.Join("tests", "test_main.cpp")

	// Flat: no PUBLIC includes — everything is private to the target.
	// For header-only, INTERFACE include points to src/.
	var includeBlock string
	if pt == config.TypeHeaderOnly {
		includeBlock = buildIncludeBlock(
			[]string{
				"$<BUILD_INTERFACE:${PROJECT_SOURCE_DIR}/src>",
				"$<INSTALL_INTERFACE:src>",
			},
			nil,
			pt,
		)
	} else {
		includeBlock = buildIncludeBlock(
			nil,
			[]string{"${CMAKE_CURRENT_SOURCE_DIR}"},
			pt,
		)
	}

	testIncludeBlock := indent(4, "${PROJECT_SOURCE_DIR}/src")

	return &Spec{
		Kind:                  config.LayoutFlat,
		Dirs:                  dirs,
		MainCPP:               mainCPP,
		LibCPP:                libCPP,
		PublicHPP:             publicHPP,
		TestCPP:               testCPP,
		CMakeSubdir:           "src",
		CMakeTestsDir:         "tests",
		CMakeIncludeBlock:     includeBlock,
		CMakeTestIncludeBlock: testIncludeBlock,
		IncludePrefix:         "",
		LayoutNote: "# Layout: Flat\n" +
			"# Headers e implementações juntos em src/. Ideal para executáveis e projetos simples.\n" +
			"# Referência: convenção common para projetos pequenos e scripts.",
	}
}

// resolveModular returns the spec for the Pitchfork multi-module layout with libs/.
//
// Each library/module lives in its own subdirectory within libs/.
// Executables are in apps/. This layout facilitates the future extraction of modules
// into separate projects and is suitable for large projects with multiple components.
//
// Generated structure:
//
//	libs/<name>/include/<name>/  ← public headers
//	libs/<name>/src/             ← implementations
//	apps/                        ← executables (entry points)
//	tests/                       ← tests
//	cmake/                       ← auxiliary CMake modules
//	docs/                        ← documentation
//
// Example include: #include "<name>/file.hpp"
// CMake: PUBLIC libs/<name>/include, PRIVATE libs/<name>/src
func resolveModular(name, nameSnake string, pt config.ProjectType) *Spec {
	libDir := filepath.Join("libs", name)

	dirs := []string{
		filepath.Join(libDir, "include", name),
		filepath.Join(libDir, "src"),
		"apps",
		"tests",
		"cmake",
		"docs",
	}

	// For modular, the executable is in apps/ and the lib in libs/<name>/
	mainCPP := filepath.Join("apps", "main.cpp")
	libCPP := filepath.Join(libDir, "src", nameSnake+".cpp")
	publicHPP := filepath.Join(libDir, "include", name, nameSnake+".hpp")
	testCPP := filepath.Join("tests", "test_main.cpp")

	// CMakeSubdir points to the library. The apps/main.cpp executable
	// is added directly in the root CMakeLists.txt (see cmake.go).
	cmakeSubdir := libDir

	// In the modular layout, the CMakeLists.txt in libs/<name>/ ALWAYS defines a
	// static library target — even when the project is an executable.
	// The executable (apps/main.cpp) is defined in the root CMakeLists.txt and links
	// to the lib. Therefore we force TypeStaticLib to generate correct PUBLIC/PRIVATE.
	includeBlock := buildIncludeBlock(
		[]string{
			"$<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>",
			"$<INSTALL_INTERFACE:include>",
		},
		[]string{
			"${CMAKE_CURRENT_SOURCE_DIR}/src",
		},
		config.TypeStaticLib, // always lib, regardless of the project type
	)

	testIncludeBlock := indent(4,
		fmt.Sprintf("${PROJECT_SOURCE_DIR}/%s/include", libDir),
	)

	return &Spec{
		Kind:                  config.LayoutModular,
		Dirs:                  dirs,
		MainCPP:               mainCPP,
		LibCPP:                libCPP,
		PublicHPP:             publicHPP,
		TestCPP:               testCPP,
		CMakeSubdir:           cmakeSubdir,
		CMakeTestsDir:         "tests",
		CMakeIncludeBlock:     includeBlock,
		CMakeTestIncludeBlock: testIncludeBlock,
		CMakeModularLibDir:    libDir,
		IncludePrefix:         name + "/",
		LayoutNote: fmt.Sprintf(
			"# Layout: Modular (Pitchfork libs/)\n"+
				"# Biblioteca em %s/include/%s/ e %s/src/. Executáveis em apps/.\n"+
				"# Referência: https://vector-of-bool.github.io/pitchfork (libs/ submodules)",
			libDir, name, libDir,
		),
	}
}

// resolveTwoRoot returns the spec for the two-root layout: include/ + src/ without a namespace subdir.
//
// Similar to Separate, but headers are directly in include/*.hpp
// without creating a namespace subdirectory (include/<name>/). Common in smaller
// projects or libraries with a single main header.
//
// Generated structure:
//
//	include/  ← public headers (without namespace subdir)
//	src/      ← implementations
//	tests/    ← tests
//	cmake/    ← auxiliary CMake modules
//	docs/     ← documentation
//
// Example include: #include "file.hpp" (without namespace prefix)
// CMake: PUBLIC include, PRIVATE src
func resolveTwoRoot(name, nameSnake string, pt config.ProjectType) *Spec {
	isHeaderOnly := pt == config.TypeHeaderOnly

	dirs := []string{
		"include",
		"cmake",
		"tests",
		"docs",
	}
	if !isHeaderOnly {
		dirs = append(dirs, "src")
	}

	mainCPP := filepath.Join("src", "main.cpp")
	libCPP := filepath.Join("src", nameSnake+".cpp")
	publicHPP := filepath.Join("include", nameSnake+".hpp")
	testCPP := filepath.Join("tests", "test_main.cpp")

	includeBlock := buildIncludeBlock(
		[]string{
			"$<BUILD_INTERFACE:${PROJECT_SOURCE_DIR}/include>",
			"$<INSTALL_INTERFACE:include>",
		},
		[]string{
			"${CMAKE_CURRENT_SOURCE_DIR}",
		},
		pt,
	)

	testIncludeBlock := indent(4, "${PROJECT_SOURCE_DIR}/include")

	return &Spec{
		Kind:                  config.LayoutTwoRoot,
		Dirs:                  dirs,
		MainCPP:               mainCPP,
		LibCPP:                libCPP,
		PublicHPP:             publicHPP,
		TestCPP:               testCPP,
		CMakeSubdir:           "src",
		CMakeTestsDir:         "tests",
		CMakeIncludeBlock:     includeBlock,
		CMakeTestIncludeBlock: testIncludeBlock,
		IncludePrefix:         "",
		LayoutNote: "# Layout: Two-Root\n" +
			"# Headers públicos em include/ (sem subdir de namespace), implementações em src/.\n" +
			"# Referência: convenção simples para bibliotecas com interface mínima.",
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// CMake block construction helpers
// ─────────────────────────────────────────────────────────────────────────────

// buildIncludeBlock builds the arguments block for target_include_directories()
// based on the provided public and private directories, adapted to the project type.
//
// For header-only projects (INTERFACE), the entire block uses INTERFACE instead of
// PUBLIC/PRIVATE, as INTERFACE targets cannot have PRIVATE properties.
//
// Example output for a static library:
//
//	PUBLIC
//	    $<BUILD_INTERFACE:${PROJECT_SOURCE_DIR}/include>
//	    $<INSTALL_INTERFACE:include>
//	PRIVATE
//	    ${CMAKE_CURRENT_SOURCE_DIR}
func buildIncludeBlock(public, private []string, pt config.ProjectType) string {
	var sb strings.Builder

	if pt == config.TypeHeaderOnly {
		// Header-only uses INTERFACE for PUBLIC and ignores PRIVATE
		if len(public) > 0 {
			sb.WriteString("    INTERFACE\n")
			for _, p := range public {
				sb.WriteString("        " + p + "\n")
			}
		}
		return strings.TrimRight(sb.String(), "\n")
	}

	// Executable: uses only PRIVATE (no installable public API)
	if pt == config.TypeExecutable {
		allPrivate := append(public, private...)
		if len(allPrivate) > 0 {
			sb.WriteString("    PRIVATE\n")
			for _, p := range allPrivate {
				sb.WriteString("        " + p + "\n")
			}
		}
		return strings.TrimRight(sb.String(), "\n")
	}

	// Static library: PUBLIC for the API, PRIVATE for implementation
	if len(public) > 0 {
		sb.WriteString("    PUBLIC\n")
		for _, p := range public {
			sb.WriteString("        " + p + "\n")
		}
	}
	if len(private) > 0 {
		sb.WriteString("    PRIVATE\n")
		for _, p := range private {
			sb.WriteString("        " + p + "\n")
		}
	}

	return strings.TrimRight(sb.String(), "\n")
}

// indent returns the string with n spaces of indentation on the left.
// Used to format include directories in the tests CMakeLists.txt block.
func indent(n int, s string) string {
	return strings.Repeat(" ", n) + s
}

// ─────────────────────────────────────────────────────────────────────────────
// Human-readable layout report (for diagnostics and --verbose)
// ─────────────────────────────────────────────────────────────────────────────

// Summary returns a multi-line string with a human-readable summary of the Spec,
// useful for display in verbose mode or diagnostic logs.
func (s *Spec) Summary() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Layout:        %s\n", s.Kind))
	sb.WriteString(fmt.Sprintf("CMake subdir:  %s\n", s.CMakeSubdir))
	sb.WriteString(fmt.Sprintf("Include prefix: %q\n", s.IncludePrefix))
	sb.WriteString("Directories:\n")
	for _, d := range s.Dirs {
		sb.WriteString(fmt.Sprintf("  %s/\n", d))
	}

	return sb.String()
}
