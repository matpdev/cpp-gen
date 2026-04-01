// Package generator contém toda a lógica de geração de projetos C++ do cpp-gen.
package generator

import (
	"fmt"
	"os"
	"path/filepath"
)

// ─────────────────────────────────────────────────────────────────────────────
// generateStructure — ponto de entrada
// ─────────────────────────────────────────────────────────────────────────────

// generateStructure cria a hierarquia de diretórios do projeto e os arquivos
// fonte C++ iniciais de acordo com o tipo de projeto configurado.
//
// Estrutura gerada (executável como exemplo):
//
//	<nome>/
//	├── src/
//	│   └── main.cpp
//	├── include/
//	│   └── <nome>/         ← namespace de includes do projeto
//	├── tests/
//	│   └── test_main.cpp
//	├── cmake/              ← módulos CMake auxiliares (populado por cmake.go)
//	└── docs/               ← documentação (vazio, referenciado no README)
func generateStructure(root string, data *TemplateData, verbose bool) error {
	// ── 1. Criar diretórios base ──────────────────────────────────────────────
	if err := createDirectories(root, data); err != nil {
		return fmt.Errorf("criar diretórios: %w", err)
	}

	// ── 2. Gerar arquivos fonte C++ de acordo com o tipo de projeto ───────────
	if err := generateSourceFiles(root, data, verbose); err != nil {
		return fmt.Errorf("gerar arquivos fonte: %w", err)
	}

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Criação de diretórios
// ─────────────────────────────────────────────────────────────────────────────

// createDirectories cria todos os diretórios necessários para o projeto.
// Os diretórios cmake/ e docs/ são sempre criados; src/, include/ e tests/
// são criados para todos os tipos de projeto.
func createDirectories(root string, data *TemplateData) error {
	dirs := []string{
		"cmake",                             // módulos CMake auxiliares
		"docs",                              // documentação do projeto
		"tests",                             // testes unitários/integração
		filepath.Join("include", data.Name), // headers públicos sob namespace
	}

	// src/ só é criado para projetos não header-only (header-only não tem .cpp)
	if !data.IsHeaderOnly {
		dirs = append(dirs, "src")
	}

	for _, d := range dirs {
		path := filepath.Join(root, d)
		if err := mkdirAll(path); err != nil {
			return fmt.Errorf("mkdir %q: %w", d, err)
		}
	}

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Geração de arquivos fonte
// ─────────────────────────────────────────────────────────────────────────────

// generateSourceFiles escolhe qual conjunto de arquivos C++ gerar
// baseado no tipo de projeto (executável, biblioteca ou header-only).
func generateSourceFiles(root string, data *TemplateData, verbose bool) error {
	switch {
	case data.IsExecutable:
		return generateExecutableFiles(root, data, verbose)
	case data.IsStaticLib:
		return generateStaticLibFiles(root, data, verbose)
	case data.IsHeaderOnly:
		return generateHeaderOnlyFiles(root, data, verbose)
	default:
		return fmt.Errorf("tipo de projeto desconhecido")
	}
}

// ── Executável ────────────────────────────────────────────────────────────────

// generateExecutableFiles gera os arquivos iniciais para um projeto executável:
//   - src/main.cpp          — ponto de entrada com exemplo de uso
//   - tests/test_main.cpp   — esqueleto de testes
func generateExecutableFiles(root string, data *TemplateData, verbose bool) error {
	files := []struct {
		relPath  string
		template string
		tmplName string
	}{
		{
			relPath:  filepath.Join("src", "main.cpp"),
			tmplName: "main.cpp",
			template: tmplExecutableMain,
		},
		{
			relPath:  filepath.Join("tests", "test_main.cpp"),
			tmplName: "test_main.cpp",
			template: tmplTestMain,
		},
	}

	for _, f := range files {
		path := filepath.Join(root, f.relPath)
		if err := writeTemplate(path, f.tmplName, f.template, data, verbose); err != nil {
			return fmt.Errorf("gerar %s: %w", f.relPath, err)
		}
	}

	return nil
}

// ── Biblioteca estática ───────────────────────────────────────────────────────

// generateStaticLibFiles gera os arquivos iniciais para uma biblioteca estática:
//   - src/<nome>.cpp                 — implementação da biblioteca
//   - include/<nome>/<nome>.hpp      — API pública da biblioteca
//   - tests/test_main.cpp            — esqueleto de testes
func generateStaticLibFiles(root string, data *TemplateData, verbose bool) error {
	files := []struct {
		relPath  string
		template string
		tmplName string
	}{
		{
			relPath:  filepath.Join("src", data.NameSnake+".cpp"),
			tmplName: "lib.cpp",
			template: tmplLibraryCpp,
		},
		{
			relPath:  filepath.Join("include", data.Name, data.NameSnake+".hpp"),
			tmplName: "lib.hpp",
			template: tmplLibraryHpp,
		},
		{
			relPath:  filepath.Join("tests", "test_main.cpp"),
			tmplName: "test_main.cpp",
			template: tmplTestMain,
		},
	}

	for _, f := range files {
		path := filepath.Join(root, f.relPath)
		if err := writeTemplate(path, f.tmplName, f.template, data, verbose); err != nil {
			return fmt.Errorf("gerar %s: %w", f.relPath, err)
		}
	}

	return nil
}

// ── Header-only ───────────────────────────────────────────────────────────────

// generateHeaderOnlyFiles gera os arquivos iniciais para uma biblioteca header-only:
//   - include/<nome>/<nome>.hpp  — implementação completa no header
//   - tests/test_main.cpp        — esqueleto de testes
func generateHeaderOnlyFiles(root string, data *TemplateData, verbose bool) error {
	files := []struct {
		relPath  string
		template string
		tmplName string
	}{
		{
			relPath:  filepath.Join("include", data.Name, data.NameSnake+".hpp"),
			tmplName: "header_only.hpp",
			template: tmplHeaderOnly,
		},
		{
			relPath:  filepath.Join("tests", "test_main.cpp"),
			tmplName: "test_main.cpp",
			template: tmplTestMain,
		},
	}

	for _, f := range files {
		path := filepath.Join(root, f.relPath)
		if err := writeTemplate(path, f.tmplName, f.template, data, verbose); err != nil {
			return fmt.Errorf("gerar %s: %w", f.relPath, err)
		}
	}

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// mkdirAll — wrapper com mensagem de erro contextualizada
// ─────────────────────────────────────────────────────────────────────────────

// mkdirAll cria o diretório e todos os pais necessários com permissão 0755.
// É um wrapper fino sobre os.MkdirAll com mensagem de erro contextualizada.
func mkdirAll(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("criar diretório %q: %w", path, err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Templates C++
// ─────────────────────────────────────────────────────────────────────────────

// tmplExecutableMain é o template para src/main.cpp de projetos executáveis.
// Gera um "Hello, World!" com estilo C++ moderno como ponto de partida.
const tmplExecutableMain = `/**
 * @file main.cpp
 * @brief Ponto de entrada do projeto {{.Name}}.
 *
 * @author {{if .Author}}{{.Author}}{{else}}Autor{{end}}
 * @version {{.Version}}
 * @date {{.Year}}
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

// tmplLibraryCpp é o template para src/<nome>.cpp de bibliotecas estáticas.
// Contém a implementação inicial com namespace baseado no nome do projeto.
const tmplLibraryCpp = `/**
 * @file {{.NameSnake}}.cpp
 * @brief Implementação da biblioteca {{.Name}}.
 *
 * @author {{if .Author}}{{.Author}}{{else}}Autor{{end}}
 * @version {{.Version}}
 * @date {{.Year}}
 */

#include "{{.Name}}/{{.NameSnake}}.hpp"

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

// tmplLibraryHpp é o template para include/<nome>/<nome>.hpp de bibliotecas estáticas.
// Define a API pública com include guard e namespace.
const tmplLibraryHpp = `/**
 * @file {{.NameSnake}}.hpp
 * @brief API pública da biblioteca {{.Name}}.
 *
 * @author {{if .Author}}{{.Author}}{{else}}Autor{{end}}
 * @version {{.Version}}
 * @date {{.Year}}
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
// Toda a implementação fica no arquivo .hpp dentro do namespace.
const tmplHeaderOnly = `/**
 * @file {{.NameSnake}}.hpp
 * @brief Biblioteca header-only {{.Name}}.
 *
 * Inclua este header diretamente no seu projeto:
 * @code
 *   #include <{{.Name}}/{{.NameSnake}}.hpp>
 * @endcode
 *
 * @author {{if .Author}}{{.Author}}{{else}}Autor{{end}}
 * @version {{.Version}}
 * @date {{.Year}}
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

// tmplTestMain é o template para tests/test_main.cpp.
// Usa a abordagem de testes simples com assert — compatível com CTest
// sem requerer dependência externa de framework de testes.
const tmplTestMain = `/**
 * @file test_main.cpp
 * @brief Testes de {{.Name}}.
 *
 * Testes registrados com CTest via add_test() no CMakeLists.txt de tests/.
 * Para adicionar um framework de testes (Catch2, GoogleTest, doctest),
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
 */

#include <cassert>
#include <iostream>
#include <stdexcept>
#include <string>
{{- if not .IsExecutable}}
#include "{{.Name}}/{{.NameSnake}}.hpp"
{{- end}}

// ─────────────────────────────────────────────────────────────────────────────
// Macro auxiliar para testes
// ─────────────────────────────────────────────────────────────────────────────

/**
 * @brief Macro para verificar uma condição e encerrar com mensagem de erro.
 *
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
    std::cout << "=== Testes de {{.Name}} ===\n\n";

    int failures = 0;
{{- if not .IsExecutable}}
    failures += test_greet_valid();
    failures += test_greet_empty_throws();
{{- else}}
    // Adicione seus testes aqui.
    // Exemplo:
    //   failures += test_minha_funcao();
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
