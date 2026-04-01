// Package generator contém toda a lógica de geração de projetos C++ do cpp-gen.
package generator

import (
	"fmt"
	"os"
	"path/filepath"

	"cpp-gen/internal/generator/layout"
)

// ─────────────────────────────────────────────────────────────────────────────
// generateStructure — ponto de entrada
// ─────────────────────────────────────────────────────────────────────────────

// generateStructure cria a hierarquia de diretórios do projeto e os arquivos
// fonte C++ iniciais de acordo com o tipo de projeto e o layout de pastas
// configurados.
//
// O layout.Spec (calculado em layout.Resolve) determina:
//   - Quais diretórios criar
//   - Onde cada arquivo C++ gerado é colocado
//   - O prefixo usado nas diretivas #include dos arquivos gerados
//
// Estrutura exemplificada para layout "separate" + tipo "static-lib":
//
//	<nome>/
//	├── include/<nome>/  ← headers públicos
//	│   └── <nome>.hpp
//	├── src/             ← implementações
//	│   └── <nome>.cpp
//	├── tests/
//	│   └── test_main.cpp
//	├── cmake/
//	└── docs/
func generateStructure(root string, data *TemplateData, spec *layout.Spec, verbose bool) error {
	// ── 1. Criar diretórios ───────────────────────────────────────────────────
	if err := createDirectories(root, spec); err != nil {
		return fmt.Errorf("criar diretórios: %w", err)
	}

	// ── 2. Gerar arquivos fonte C++ ───────────────────────────────────────────
	if err := generateSourceFiles(root, data, spec, verbose); err != nil {
		return fmt.Errorf("gerar arquivos fonte: %w", err)
	}

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Criação de diretórios
// ─────────────────────────────────────────────────────────────────────────────

// createDirectories cria todos os diretórios definidos na spec do layout.
// Os diretórios são criados com os pais necessários (equivalente a mkdir -p).
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
// Geração de arquivos fonte
// ─────────────────────────────────────────────────────────────────────────────

// generateSourceFiles despacha para o gerador de arquivos correto com base
// no tipo de projeto (executável, biblioteca estática ou header-only).
//
// Os caminhos de destino são lidos diretamente do layout.Spec, portanto
// esta função não precisa saber qual layout está ativo — apenas qual tipo
// de artefato gerar.
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

// ── Executável ────────────────────────────────────────────────────────────────

// generateExecutableFiles gera os arquivos iniciais para projetos executáveis.
//
// Arquivos gerados (caminhos conforme o layout.Spec):
//   - <spec.MainCPP>   — ponto de entrada (main.cpp ou apps/main.cpp etc.)
//   - <spec.TestCPP>   — esqueleto de testes
//
// Para o layout Modular, além do apps/main.cpp, são gerados os arquivos da
// biblioteca reutilizável em libs/<nome>/ (LibCPP e PublicHPP), pois o
// executável é apenas um front-end fino que linka à lib.
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

	// No layout modular, o executável em apps/ é apenas um front-end fino.
	// A lógica reutilizável fica em libs/<nome>/, portanto também geramos
	// os arquivos da biblioteca (implementação + header público).
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

// ── Biblioteca estática ───────────────────────────────────────────────────────

// generateStaticLibFiles gera os arquivos iniciais para bibliotecas estáticas.
//
// Arquivos gerados (caminhos conforme o layout.Spec):
//   - <spec.LibCPP>    — implementação da biblioteca
//   - <spec.PublicHPP> — header público da API
//   - <spec.TestCPP>   — esqueleto de testes
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

// generateHeaderOnlyFiles gera os arquivos iniciais para bibliotecas header-only.
//
// Arquivos gerados (caminhos conforme o layout.Spec):
//   - <spec.PublicHPP> — implementação completa no header
//   - <spec.TestCPP>   — esqueleto de testes
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
// Templates C++ — os dados injetados vêm de TemplateData
// ─────────────────────────────────────────────────────────────────────────────

// tmplExecutableMain é o template para o ponto de entrada de projetos executáveis.
// O campo {{.LayoutIncludePrefix}} garante que o #include use o caminho correto
// independentemente do layout escolhido.
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

// tmplLibraryCpp é o template para o arquivo de implementação de bibliotecas
// estáticas. O #include usa {{.LayoutIncludePrefix}} para ser correto em
// qualquer layout (ex: "mylib/mylib.hpp" no separate, "mylib.hpp" no flat).
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

// tmplLibraryHpp é o template para o header público de bibliotecas estáticas.
// Usa #pragma once como include guard moderno e define o namespace baseado
// no nome do projeto em snake_case.
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

// tmplHeaderOnly é o template para bibliotecas header-only.
// Toda a implementação fica inline no arquivo .hpp.
// O path de include no Doxygen usa {{.LayoutIncludePrefix}} para ser
// correto independentemente do layout.
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

// tmplTestMain é o template para o arquivo de testes.
// Usa a macro CHECK() como framework mínimo compatível com CTest,
// sem requerer dependências externas.
// O #include usa {{.LayoutIncludePrefix}} para ser correto em qualquer layout.
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

// mkdirAll é um wrapper sobre os.MkdirAll para uso nos testes do pacote.
// O gerador principal usa os.MkdirAll diretamente via createDirectories.
func mkdirAll(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("criar diretório %q: %w", path, err)
	}
	return nil
}
