// Package generator contains all C++ project generation logic for cpp-gen.
package generator

import (
	"fmt"
	"os"
	"path/filepath"

	"cpp-gen/internal/generator/layout"
)

// ─────────────────────────────────────────────────────────────────────────────
// generateStructure — entry point
// ─────────────────────────────────────────────────────────────────────────────

// generateStructure creates the project directory hierarchy and initial C++
// source files according to the configured project type and folder layout.
//
// The layout.Spec (calculated in layout.Resolve) determines:
//   - Which directories to create
//   - Where each generated C++ file is placed
//   - The prefix used in #include directives of generated files
//
// Example structure for layout "separate" + type "static-lib":
//
//	<nome>/
//	├── include/<nome>/  ← public headers
//	│   └── <nome>.hpp
//	├── src/             ← implementations
//	│   └── <nome>.cpp
//	├── tests/
//	│   └── test_main.cpp
//	├── cmake/
//	└── docs/
func generateStructure(root string, data *TemplateData, spec *layout.Spec, verbose bool) error {
	// ── 1. Create directories ─────────────────────────────────────────────────
	if err := createDirectories(root, spec); err != nil {
		return fmt.Errorf("criar diretórios: %w", err)
	}

	// ── 2. Generate C++ source files ──────────────────────────────────────────
	if err := generateSourceFiles(root, data, spec, verbose); err != nil {
		return fmt.Errorf("gerar arquivos fonte: %w", err)
	}

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Directory creation
// ─────────────────────────────────────────────────────────────────────────────

// createDirectories creates all directories defined in the layout spec.
// Directories are created with all necessary parents (equivalent to mkdir -p).
func createDirectories(root string, spec *layout.Spec) error {
	for _, d := range spec.Dirs {
		path := filepath.Join(root, d)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("mkdir %q: %w", d, err)
		}
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Source file generation
// ─────────────────────────────────────────────────────────────────────────────

// generateSourceFiles dispatches to the correct file generator based on
// the project type (executable, static library or header-only).
//
// Destination paths are read directly from the layout.Spec, so
// this function does not need to know which layout is active — only which
// type of artifact to generate.
func generateSourceFiles(root string, data *TemplateData, spec *layout.Spec, verbose bool) error {
	switch {
	case data.IsExecutable:
		return generateExecutableFiles(root, data, spec, verbose)
	case data.IsStaticLib:
		return generateStaticLibFiles(root, data, spec, verbose)
	case data.IsHeaderOnly:
		return generateHeaderOnlyFiles(root, data, spec, verbose)
	default:
		return fmt.Errorf("tipo de projeto desconhecido")
	}
}

// ── Executable ───────────────────────────────────────────────────────────────

// generateExecutableFiles generates the initial files for executable projects.
//
// Generated files (paths per layout.Spec):
//   - <spec.MainCPP>   — entry point (main.cpp or apps/main.cpp etc.)
//   - <spec.TestCPP>   — test skeleton
//
// For the Modular layout, in addition to apps/main.cpp, the reusable library
// files in libs/<name>/ (LibCPP and PublicHPP) are also generated, since the
// executable is just a thin front-end that links to the lib.
func generateExecutableFiles(root string, data *TemplateData, spec *layout.Spec, verbose bool) error {
	files := []struct {
		path     string
		tmplName string
		tmpl     string
	}{
		{
			path:     filepath.Join(root, spec.MainCPP),
			tmplName: "main.cpp",
			tmpl:     tmplExecutableMain,
		},
		{
			path:     filepath.Join(root, spec.TestCPP),
			tmplName: "test_main.cpp",
			tmpl:     tmplTestMain,
		},
	}

	// In the modular layout, the executable in apps/ is just a thin front-end.
	// The reusable logic lives in libs/<name>/, so we also generate
	// the library files (implementation + public header).
	if spec.Kind == layout.KindModular {
		files = append(files,
			struct {
				path     string
				tmplName string
				tmpl     string
			}{
				path:     filepath.Join(root, spec.LibCPP),
				tmplName: "lib.cpp",
				tmpl:     tmplLibraryCpp,
			},
			struct {
				path     string
				tmplName string
				tmpl     string
			}{
				path:     filepath.Join(root, spec.PublicHPP),
				tmplName: "lib.hpp",
				tmpl:     tmplLibraryHpp,
			},
		)
	}

	for _, f := range files {
		if err := writeTemplate(f.path, f.tmplName, f.tmpl, data, verbose); err != nil {
			return fmt.Errorf("gerar %s: %w", f.path, err)
		}
	}

	return nil
}

// ── Static library ────────────────────────────────────────────────────────────

// generateStaticLibFiles generates the initial files for static libraries.
//
// Generated files (paths per layout.Spec):
//   - <spec.LibCPP>    — library implementation
//   - <spec.PublicHPP> — public API header
//   - <spec.TestCPP>   — test skeleton
func generateStaticLibFiles(root string, data *TemplateData, spec *layout.Spec, verbose bool) error {
	files := []struct {
		path     string
		tmplName string
		tmpl     string
	}{
		{
			path:     filepath.Join(root, spec.LibCPP),
			tmplName: "lib.cpp",
			tmpl:     tmplLibraryCpp,
		},
		{
			path:     filepath.Join(root, spec.PublicHPP),
			tmplName: "lib.hpp",
			tmpl:     tmplLibraryHpp,
		},
		{
			path:     filepath.Join(root, spec.TestCPP),
			tmplName: "test_main.cpp",
			tmpl:     tmplTestMain,
		},
	}

	for _, f := range files {
		if err := writeTemplate(f.path, f.tmplName, f.tmpl, data, verbose); err != nil {
			return fmt.Errorf("gerar %s: %w", f.path, err)
		}
	}

	return nil
}

// ── Header-only ───────────────────────────────────────────────────────────────

// generateHeaderOnlyFiles generates the initial files for header-only libraries.
//
// Generated files (paths per layout.Spec):
//   - <spec.PublicHPP> — complete implementation in the header
//   - <spec.TestCPP>   — test skeleton
func generateHeaderOnlyFiles(root string, data *TemplateData, spec *layout.Spec, verbose bool) error {
	files := []struct {
		path     string
		tmplName string
		tmpl     string
	}{
		{
			path:     filepath.Join(root, spec.PublicHPP),
			tmplName: "header_only.hpp",
			tmpl:     tmplHeaderOnly,
		},
		{
			path:     filepath.Join(root, spec.TestCPP),
			tmplName: "test_main.cpp",
			tmpl:     tmplTestMain,
		},
	}

	for _, f := range files {
		if err := writeTemplate(f.path, f.tmplName, f.tmpl, data, verbose); err != nil {
			return fmt.Errorf("gerar %s: %w", f.path, err)
		}
	}

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// C++ templates — injected data comes from TemplateData
// ─────────────────────────────────────────────────────────────────────────────

// tmplExecutableMain is the template for the entry point of executable projects.
// The {{.LayoutIncludePrefix}} field ensures that #include uses the correct path
// regardless of the chosen layout.
const tmplExecutableMain = `/**
 * @file main.cpp
 * @brief Ponto de entrada do projeto {{.Name}}.
 *
 * @author {{if .Author}}{{.Author}}{{else}}Autor{{end}}
 * @version {{.Version}}
 * @date {{.Year}}
 *
 * Layout: {{.Layout}}
 */

#include <iostream>
#include <string_view>

// ─────────────────────────────────────────────────────────────────────────────

/**
 * @brief Exibe uma mensagem de boas-vindas no terminal.
 * @param project_name Nome do projeto a ser exibido.
 */
void greet(std::string_view project_name) {
    std::cout << "Hello from " << project_name << "!\n";
}

// ─────────────────────────────────────────────────────────────────────────────

int main([[maybe_unused]] int argc, [[maybe_unused]] char* argv[]) {
    greet("{{.Name}}");
    return 0;
}
`

// tmplLibraryCpp is the template for the implementation file of static libraries.
// The #include uses {{.LayoutIncludePrefix}} to be correct in any layout
// (e.g. "mylib/mylib.hpp" in separate, "mylib.hpp" in flat).
const tmplLibraryCpp = `/**
 * @file {{.NameSnake}}.cpp
 * @brief Implementação da biblioteca {{.Name}}.
 *
 * @author {{if .Author}}{{.Author}}{{else}}Autor{{end}}
 * @version {{.Version}}
 * @date {{.Year}}
 *
 * Layout: {{.Layout}}
 */

#include "{{.LayoutIncludePrefix}}{{.NameSnake}}.hpp"

#include <stdexcept>

namespace {{.NameSnake}} {

// ─────────────────────────────────────────────────────────────────────────────
// Implementação de {{.NamePascal}}
// ─────────────────────────────────────────────────────────────────────────────

{{.NamePascal}}::{{.NamePascal}}() = default;

{{.NamePascal}}::~{{.NamePascal}}() = default;

std::string {{.NamePascal}}::greet(std::string_view name) const {
    if (name.empty()) {
        throw std::invalid_argument{"name cannot be empty"};
    }
    return "Hello from " + std::string{name} + "!";
}

} // namespace {{.NameSnake}}
`

// tmplLibraryHpp is the template for the public header of static libraries.
// Uses #pragma once as a modern include guard and defines the namespace based
// on the project name in snake_case.
const tmplLibraryHpp = `/**
 * @file {{.NameSnake}}.hpp
 * @brief API pública da biblioteca {{.Name}}.
 *
 * Inclua com:
 * @code
 *   #include "{{.LayoutIncludePrefix}}{{.NameSnake}}.hpp"
 * @endcode
 *
 * @author {{if .Author}}{{.Author}}{{else}}Autor{{end}}
 * @version {{.Version}}
 * @date {{.Year}}
 *
 * Layout: {{.Layout}}
 */

#pragma once

#include <string>
#include <string_view>

namespace {{.NameSnake}} {

// ─────────────────────────────────────────────────────────────────────────────

/**
 * @brief Classe principal da biblioteca {{.Name}}.
 *
 * Ponto de entrada para todas as funcionalidades da biblioteca.
 * Substitua esta documentação pela descrição real da sua API.
 */
class {{.NamePascal}} {
public:
    {{.NamePascal}}();
    ~{{.NamePascal}}();

    // Impede cópia acidental — habilite se necessário.
    {{.NamePascal}}(const {{.NamePascal}}&)            = delete;
    {{.NamePascal}}& operator=(const {{.NamePascal}}&) = delete;

    // Move é permitido por padrão.
    {{.NamePascal}}({{.NamePascal}}&&)            = default;
    {{.NamePascal}}& operator=({{.NamePascal}}&&) = default;

    /**
     * @brief Exemplo de método público.
     * @param name Nome a ser saudado.
     * @return String de saudação.
     * @throws std::invalid_argument se @p name for vazio.
     */
    [[nodiscard]] std::string greet(std::string_view name) const;
};

// ─────────────────────────────────────────────────────────────────────────────

} // namespace {{.NameSnake}}
`

// tmplHeaderOnly is the template for header-only libraries.
// The entire implementation is inline in the .hpp file.
// The Doxygen include path uses {{.LayoutIncludePrefix}} to be
// correct regardless of the layout.
const tmplHeaderOnly = `/**
 * @file {{.NameSnake}}.hpp
 * @brief Biblioteca header-only {{.Name}}.
 *
 * Inclua com:
 * @code
 *   #include "{{.LayoutIncludePrefix}}{{.NameSnake}}.hpp"
 * @endcode
 *
 * @author {{if .Author}}{{.Author}}{{else}}Autor{{end}}
 * @version {{.Version}}
 * @date {{.Year}}
 *
 * Layout: {{.Layout}}
 */

#pragma once

#include <string>
#include <string_view>
#include <stdexcept>

namespace {{.NameSnake}} {

// ─────────────────────────────────────────────────────────────────────────────

/**
 * @brief Classe principal da biblioteca header-only {{.Name}}.
 *
 * Por ser header-only, toda a implementação está inline neste arquivo.
 * Substitua pelo conteúdo real da sua biblioteca.
 */
class {{.NamePascal}} {
public:
    {{.NamePascal}}()  = default;
    ~{{.NamePascal}}() = default;

    /**
     * @brief Exemplo de método inline.
     * @param name Nome a ser saudado.
     * @return String de saudação.
     * @throws std::invalid_argument se @p name for vazio.
     */
    [[nodiscard]] std::string greet(std::string_view name) const {
        if (name.empty()) {
            throw std::invalid_argument{"name cannot be empty"};
        }
        return "Hello from " + std::string{name} + "!";
    }
};

// ─────────────────────────────────────────────────────────────────────────────

/**
 * @brief Função auxiliar de conveniência da biblioteca.
 * @param name Nome a ser saudado.
 * @return String de saudação.
 */
[[nodiscard]] inline std::string greet(std::string_view name) {
    return {{.NamePascal}}{}.greet(name);
}

} // namespace {{.NameSnake}}
`

// tmplTestMain is the template for the test file.
// Uses the CHECK() macro as a minimal framework compatible with CTest,
// without requiring external dependencies.
// The #include uses {{.LayoutIncludePrefix}} to be correct in any layout.
const tmplTestMain = `/**
 * @file test_main.cpp
 * @brief Testes de {{.Name}}.
 *
 * Testes registrados com CTest via add_test() no CMakeLists.txt de tests/.
 *
 * Para adicionar um framework externo (Catch2, GoogleTest, doctest),
 * configure a dependência no gerenciador de pacotes e adapte este arquivo.
 *
 * Execução:
 * @code
 *   cd build && ctest --output-on-failure
 * @endcode
 *
 * @author {{if .Author}}{{.Author}}{{else}}Autor{{end}}
 * @version {{.Version}}
 * @date {{.Year}}
 *
 * Layout: {{.Layout}}
 */

#include <cassert>
#include <iostream>
#include <stdexcept>
#include <string>
{{- if not .IsExecutable}}
#include "{{.LayoutIncludePrefix}}{{.NameSnake}}.hpp"
{{- end}}

// ─────────────────────────────────────────────────────────────────────────────
// Macro auxiliar para testes
// ─────────────────────────────────────────────────────────────────────────────

/**
 * @brief Verifica uma condição e encerra com mensagem de erro se falsa.
 * Diferente de assert(), não é removido em builds Release.
 */
#define CHECK(condition)                                                     \
    do {                                                                     \
        if (!(condition)) {                                                  \
            std::cerr << "[FALHOU] " << __FILE__ << ":" << __LINE__          \
                      << " — condição falsa: " #condition "\n";             \
            return 1;                                                        \
        }                                                                    \
        std::cout << "[OK]     " #condition "\n";                            \
    } while (false)

// ─────────────────────────────────────────────────────────────────────────────
// Testes
// ─────────────────────────────────────────────────────────────────────────────
{{- if not .IsExecutable}}

/**
 * @brief Testa o método greet() com uma entrada válida.
 */
static int test_greet_valid() {
    {{.NameSnake}}::{{.NamePascal}} obj;
    const auto result = obj.greet("World");
    CHECK(result == "Hello from World!");
    return 0;
}

/**
 * @brief Testa que greet() lança exceção para entrada vazia.
 */
static int test_greet_empty_throws() {
    {{.NameSnake}}::{{.NamePascal}} obj;
    bool threw = false;
    try {
        obj.greet("");
    } catch (const std::invalid_argument&) {
        threw = true;
    }
    CHECK(threw);
    return 0;
}
{{- end}}

// ─────────────────────────────────────────────────────────────────────────────

int main() {
    std::cout << "=== Testes de {{.Name}} (layout: {{.Layout}}) ===\n\n";

    int failures = 0;
{{- if not .IsExecutable}}
    failures += test_greet_valid();
    failures += test_greet_empty_throws();
{{- else}}
    // Adicione seus testes aqui.
    CHECK(1 + 1 == 2);
{{- end}}

    std::cout << "\n";
    if (failures == 0) {
        std::cout << "Todos os testes passaram.\n";
        return 0;
    }
    std::cout << failures << " teste(s) falharam.\n";
    return 1;
}
`

// mkdirAll is a wrapper around os.MkdirAll for use in package tests.
// The main generator uses os.MkdirAll directly via createDirectories.
func mkdirAll(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("criar diretório %q: %w", path, err)
	}
	return nil
}
