// Package generator contains all C++ project generation logic for cpp-gen.
package generator

import (
	"fmt"
	"path/filepath"
)

// ─────────────────────────────────────────────────────────────────────────────
// generateCMake — entry point
// ─────────────────────────────────────────────────────────────────────────────

// generateCMake generates all CMake files for the project in the following order:
//
//  1. root CMakeLists.txt    — main project configuration
//  2. src/CMakeLists.txt     — main target definition (executable or library)
//  3. tests/CMakeLists.txt   — test target definitions with CTest
//  4. cmake/CompilerWarnings.cmake — reusable warning flags
//  5. CMakePresets.json      — configure/build/test presets for all IDEs
func generateCMake(root string, data *TemplateData, verbose bool) error {
	// The main target subdirectory varies with the layout:
	//   separate / flat / two-root → src/
	//   merged                     → <name>/
	//   modular                    → libs/<name>/
	srcCMakeRelPath := filepath.Join(data.LayoutCMakeSubdir, "CMakeLists.txt")

	// The tests subdirectory is always "tests", but we read it from the Spec for consistency.
	testsCMakeRelPath := filepath.Join(data.LayoutSpec.CMakeTestsDir, "CMakeLists.txt")

	steps := []struct {
		relPath  string
		tmplName string
		tmpl     string
	}{
		{
			relPath:  "CMakeLists.txt",
			tmplName: "root_cmake",
			tmpl:     tmplRootCMake,
		},
		{
			relPath:  srcCMakeRelPath,
			tmplName: "src_cmake",
			tmpl:     tmplSrcCMake,
		},
		{
			relPath:  testsCMakeRelPath,
			tmplName: "tests_cmake",
			tmpl:     tmplTestsCMake,
		},
		{
			relPath:  filepath.Join("cmake", "CompilerWarnings.cmake"),
			tmplName: "compiler_warnings",
			tmpl:     tmplCompilerWarnings,
		},
		{
			relPath:  "CMakePresets.json",
			tmplName: "cmake_presets",
			tmpl:     tmplCMakePresets,
		},
	}

	for _, s := range steps {
		path := filepath.Join(root, s.relPath)
		if err := writeTemplate(path, s.tmplName, s.tmpl, data, verbose); err != nil {
			return fmt.Errorf("gerar %s: %w", s.relPath, err)
		}
	}

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Template: root CMakeLists.txt
// ─────────────────────────────────────────────────────────────────────────────

// tmplRootCMake is the template for the project's main CMakeLists.txt.
//
// Responsibilities:
//   - Declares the project with version, description and language
//   - Configures the C++ standard and global compilation options
//   - Enables CMAKE_EXPORT_COMPILE_COMMANDS for Clangd / clang-tidy
//   - Includes auxiliary CMake modules (cmake/)
//   - Integrates VCPKG or FetchContent according to configuration
//   - Registers subdirectories according to the chosen layout ({{.Layout}})
//   - In the modular layout with executable, defines the app target directly here
const tmplRootCMake = `cmake_minimum_required(VERSION 3.20)

# =============================================================================
# Declaração do projeto
# =============================================================================
# A versão aqui definida é automaticamente acessível via
# {{.NameUpper}}_VERSION_MAJOR, {{.NameUpper}}_VERSION_MINOR, etc.
project(
    {{.Name}}
    VERSION {{.Version}}
    DESCRIPTION "{{if .Description}}{{.Description}}{{else}}Projeto C++ gerado pelo cpp-gen{{end}}"
    LANGUAGES CXX
)

# Impede builds in-source, que misturam artefatos com o código fonte.
if(CMAKE_SOURCE_DIR STREQUAL CMAKE_BINARY_DIR)
    message(FATAL_ERROR
        "Build in-source detectado!\n"
        "Por favor, crie um diretório separado para o build:\n"
        "  cmake -B build\n"
    )
endif()

# =============================================================================
# Padrão e configurações C++
# =============================================================================

set(CMAKE_CXX_STANDARD {{.Standard}})
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Desabilita extensões específicas do compilador (ex: gnu++20) para garantir
# portabilidade entre GCC, Clang e MSVC.
set(CMAKE_CXX_EXTENSIONS OFF)

# Exporta compile_commands.json para uso com Clangd, clang-tidy e IDEs.
# O arquivo é gerado no diretório de build e vinculado à raiz pelo .clangd.
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

# Organiza os binários em subdiretórios padrão dentro do build.
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/bin")
set(CMAKE_LIBRARY_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/lib")
set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/lib")

# =============================================================================
# Módulos CMake auxiliares
# =============================================================================

# Adiciona cmake/ ao caminho de busca de módulos, permitindo usar
# include(CompilerWarnings) e outros módulos do projeto.
list(APPEND CMAKE_MODULE_PATH "${CMAKE_CURRENT_SOURCE_DIR}/cmake")
{{if .UseVCPKG}}
# =============================================================================
# Integração VCPKG
# =============================================================================
# O VCPKG é integrado via toolchain file, que deve ser definido ANTES do
# primeiro projeto(). A forma recomendada é usar CMakePresets.json (já
# configurado neste projeto) ou passar na linha de comando:
#
#   cmake -B build --preset vcpkg-debug
#
# Variável de ambiente necessária:
#   VCPKG_ROOT=/caminho/para/vcpkg
#
# Para instalar dependências listadas em vcpkg.json, execute:
#   vcpkg install
# (ou deixe o CMake fazer automaticamente via manifest mode)
if(DEFINED ENV{VCPKG_ROOT} AND NOT DEFINED CMAKE_TOOLCHAIN_FILE)
    set(CMAKE_TOOLCHAIN_FILE
        "$ENV{VCPKG_ROOT}/scripts/buildsystems/vcpkg.cmake"
        CACHE STRING "VCPKG toolchain file"
    )
    message(STATUS "VCPKG: usando toolchain em ${CMAKE_TOOLCHAIN_FILE}")
endif()
{{end}}{{if .UseFetchContent}}
# =============================================================================
# FetchContent — dependências externas via CMake nativo
# =============================================================================
# As dependências são declaradas em cmake/Dependencies.cmake.
# FetchContent baixa, configura e constrói as dependências automaticamente
# durante o configure do CMake.
include(FetchContent)
include(cmake/Dependencies.cmake)
{{end}}
# =============================================================================
# Módulos internos
# =============================================================================

# Flags de warning do compilador definidas em cmake/CompilerWarnings.cmake.
# Aplique ao seu target com: target_link_libraries(<target> PRIVATE project_warnings)
include(CompilerWarnings)

# =============================================================================
# Subdiretórios
# =============================================================================
{{.LayoutNote}}

{{if .LayoutIsModular}}
# Layout modular: a biblioteca vive em {{.LayoutModularLibDir}}/
# O CMakeLists.txt em {{.LayoutModularLibDir}}/ define o target ${PROJECT_NAME}.
add_subdirectory({{.LayoutCMakeSubdir}})
{{if .IsExecutable}}
# O executável (apps/main.cpp) é definido aqui e linka à biblioteca acima.
add_executable(${PROJECT_NAME}_app
    apps/main.cpp
)
target_include_directories(${PROJECT_NAME}_app
    PRIVATE
        ${PROJECT_SOURCE_DIR}/{{.LayoutModularLibDir}}/include
)
target_link_libraries(${PROJECT_NAME}_app
    PRIVATE
        project_warnings
        ${PROJECT_NAME}::${PROJECT_NAME}
)
set_target_properties(${PROJECT_NAME}_app PROPERTIES
    CXX_STANDARD          {{.Standard}}
    CXX_STANDARD_REQUIRED ON
    CXX_EXTENSIONS        OFF
    OUTPUT_NAME           "{{.Name}}"
    RUNTIME_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/bin"
)
{{end}}
{{else}}
add_subdirectory({{.LayoutCMakeSubdir}})
{{end}}

# =============================================================================
# Testes
# =============================================================================

# Opção de build que pode ser desabilitada externamente:
#   cmake -B build -D{{.NameUpper}}_BUILD_TESTS=OFF
option({{.NameUpper}}_BUILD_TESTS "Compilar os testes de {{.Name}}" ON)

if({{.NameUpper}}_BUILD_TESTS)
    message(STATUS "{{.Name}}: testes habilitados")
    enable_testing()
    add_subdirectory(tests)
endif()

# =============================================================================
# Informações de configuração (exibidas ao final do cmake)
# =============================================================================

message(STATUS "")
message(STATUS "=== {{.Name}} v${PROJECT_VERSION} ===")
message(STATUS "  Compilador       : ${CMAKE_CXX_COMPILER_ID} ${CMAKE_CXX_COMPILER_VERSION}")
message(STATUS "  Padrão C++       : C++${CMAKE_CXX_STANDARD}")
message(STATUS "  Tipo de build    : ${CMAKE_BUILD_TYPE}")
message(STATUS "  Diretório de build: ${CMAKE_BINARY_DIR}")
message(STATUS "  Layout           : {{.Layout}}")
message(STATUS "  Testes           : {{printf "${%s_BUILD_TESTS}" .NameUpper}}")
message(STATUS "")
`

// ─────────────────────────────────────────────────────────────────────────────
// Template: src/CMakeLists.txt
// ─────────────────────────────────────────────────────────────────────────────

// tmplSrcCMake is the template for the CMakeLists.txt of the sources subdirectory.
//
// This file is generated in the subdirectory indicated by the layout:
//   - separate / flat / two-root: src/CMakeLists.txt
//   - merged:                     <name>/CMakeLists.txt
//   - modular:                    libs/<name>/CMakeLists.txt
//
// The {{.LayoutCMakeIncludeBlock}} field contains the pre-formatted
// target_include_directories() block correct for the chosen layout.
//
// The generated content depends on the project type:
//   - Executable:         add_executable()
//   - Static library:     add_library(STATIC)
//   - Header-only:        add_library(INTERFACE)
const tmplSrcCMake = `# =============================================================================
# =============================================================================
# CMakeLists.txt — Target principal de {{.Name}}
# =============================================================================
# Layout ativo: {{.Layout}}
# Para detalhes sobre o layout, veja o CMakeLists.txt raiz.
# =============================================================================
{{if .IsExecutable}}
# ── Executável ────────────────────────────────────────────────────────────────
{{if .LayoutIsModular}}
# No layout modular, o executável (apps/main.cpp) é definido no CMakeLists.txt
# raiz para que possa linkar com o target desta biblioteca.
# Este arquivo define apenas a biblioteca reutilizável.
add_library(${PROJECT_NAME} STATIC
    src/{{.NameSnake}}.cpp
    # Adicione mais arquivos de implementação aqui.
)

add_library({{.Name}}::{{.Name}} ALIAS ${PROJECT_NAME})

target_include_directories(${PROJECT_NAME}
{{.LayoutCMakeIncludeBlock}}
)

target_link_libraries(${PROJECT_NAME}
    PRIVATE
        project_warnings
)

set_target_properties(${PROJECT_NAME} PROPERTIES
    CXX_STANDARD          {{.Standard}}
    CXX_STANDARD_REQUIRED ON
    CXX_EXTENSIONS        OFF
)
{{else}}
add_executable(${PROJECT_NAME}
    main.cpp
    # Adicione outros arquivos .cpp aqui conforme o projeto crescer.
)

target_include_directories(${PROJECT_NAME}
{{.LayoutCMakeIncludeBlock}}
)

target_link_libraries(${PROJECT_NAME}
    PRIVATE
        project_warnings
        # Adicione dependências aqui:
        # fmt::fmt
        # nlohmann_json::nlohmann_json
)

set_target_properties(${PROJECT_NAME} PROPERTIES
    CXX_STANDARD          {{.Standard}}
    CXX_STANDARD_REQUIRED ON
    CXX_EXTENSIONS        OFF
)
{{end}}
{{end}}{{if .IsStaticLib}}
# ── Biblioteca Estática ───────────────────────────────────────────────────────
{{if .LayoutIsModular}}
add_library(${PROJECT_NAME} STATIC
    src/{{.NameSnake}}.cpp
    # Adicione mais arquivos de implementação aqui.
)
{{else}}
add_library(${PROJECT_NAME} STATIC
    {{.NameSnake}}.cpp
    # Adicione mais arquivos de implementação aqui.
)
{{end}}

# Alias com namespace — permite uso como {{.Name}}::{{.Name}} via find_package()
# ou FetchContent sem precisar conhecer o nome interno do target.
add_library({{.Name}}::{{.Name}} ALIAS ${PROJECT_NAME})

target_include_directories(${PROJECT_NAME}
{{.LayoutCMakeIncludeBlock}}
)

target_link_libraries(${PROJECT_NAME}
    PRIVATE
        project_warnings
        # Dependências privadas (implementação) aqui.
    PUBLIC
        # Dependências públicas (propagadas à API) aqui.
)

set_target_properties(${PROJECT_NAME} PROPERTIES
    CXX_STANDARD          {{.Standard}}
    CXX_STANDARD_REQUIRED ON
    CXX_EXTENSIONS        OFF
    OUTPUT_NAME           "{{.NameSnake}}"
)
{{end}}{{if .IsHeaderOnly}}
# ── Biblioteca Header-Only (INTERFACE) ────────────────────────────────────────
# Targets INTERFACE não compilam fontes; apenas propagam includes e flags.

add_library(${PROJECT_NAME} INTERFACE)

add_library({{.Name}}::{{.Name}} ALIAS ${PROJECT_NAME})

target_include_directories(${PROJECT_NAME}
{{.LayoutCMakeIncludeBlock}}
)

# Propaga o requisito de padrão C++ para quem consumir esta biblioteca.
target_compile_features(${PROJECT_NAME}
    INTERFACE cxx_std_{{.Standard}}
)
{{end}}
`

// ─────────────────────────────────────────────────────────────────────────────
// Template: tests/CMakeLists.txt
// ─────────────────────────────────────────────────────────────────────────────

// tmplTestsCMake is the template for tests/CMakeLists.txt.
// Registers test executables with CTest via add_test().
const tmplTestsCMake = `# =============================================================================
# =============================================================================
# tests/CMakeLists.txt — Testes de {{.Name}}
# =============================================================================
# Layout ativo: {{.Layout}}
#
# Os testes são executados com:
#   cmake --preset debug && ctest --preset test-debug --output-on-failure
#   cd build && ctest -V                  (verbose)
#   cd build && ctest -R <nome>           (filtrar por nome)
# =============================================================================

# ── Target de teste principal ─────────────────────────────────────────────────

add_executable({{.NameSnake}}_tests
    {{if eq .Layout "merged"}}driver.cpp{{else}}test_main.cpp{{end}}
    # Adicione arquivos de teste adicionais aqui:
    # test_feature_a.cpp
    # test_feature_b.cpp
)

target_include_directories({{.NameSnake}}_tests
    PRIVATE
{{.LayoutCMakeTestIncludeBlock}}
        ${CMAKE_CURRENT_SOURCE_DIR}
)

target_link_libraries({{.NameSnake}}_tests
    PRIVATE
        project_warnings
{{- if not .IsExecutable}}
        # Linka contra o target principal da biblioteca.
        ${PROJECT_NAME}
{{- end}}
        # Adicione frameworks de teste aqui se necessário:
        # Catch2::Catch2WithMain
        # GTest::gtest_main
        # doctest::doctest
)

set_target_properties({{.NameSnake}}_tests PROPERTIES
    CXX_STANDARD          {{.Standard}}
    CXX_STANDARD_REQUIRED ON
    CXX_EXTENSIONS        OFF
)

# ── Registro no CTest ─────────────────────────────────────────────────────────

# Registra o executável como um teste no CTest.
# O executável deve retornar 0 para indicar sucesso.
add_test(
    NAME    {{.Name}}_unit_tests
    COMMAND {{.NameSnake}}_tests
)

# Propriedades do teste (timeout em segundos, variáveis de ambiente, etc.)
set_tests_properties({{.Name}}_unit_tests PROPERTIES
    TIMEOUT 30
    LABELS  "unit"
)

# ── Dica: Adicionando testes individuais ──────────────────────────────────────
# Para granularidade maior (cada função de teste separada no CTest), use:
#
# add_test(NAME {{.Name}}_test_greet   COMMAND {{.NameSnake}}_tests greet)
# add_test(NAME {{.Name}}_test_feature COMMAND {{.NameSnake}}_tests feature)
#
# Isso requer que seu executável de testes aceite o nome do teste como argumento.
`

// ─────────────────────────────────────────────────────────────────────────────
// Template: cmake/CompilerWarnings.cmake
// ─────────────────────────────────────────────────────────────────────────────

// tmplCompilerWarnings is the template for cmake/CompilerWarnings.cmake.
// Defines a `project_warnings` interface target with warning flags
// appropriate for GCC, Clang and MSVC, which can be linked via:
//
//	target_link_libraries(<your_target> PRIVATE project_warnings)
const tmplCompilerWarnings = `# =============================================================================
# cmake/CompilerWarnings.cmake
# =============================================================================
# Define o target de interface ` + "`project_warnings`" + ` com um conjunto abrangente
# de flags de warning para GCC, Clang e MSVC.
#
# Uso:
#   target_link_libraries(<seu_target> PRIVATE project_warnings)
#
# Referências:
#   - https://gcc.gnu.org/onlinedocs/gcc/Warning-Options.html
#   - https://clang.llvm.org/docs/DiagnosticsReference.html
#   - https://docs.microsoft.com/cpp/build/reference/compiler-options
# =============================================================================

add_library(project_warnings INTERFACE)

# ── Flags para GCC e Clang ────────────────────────────────────────────────────
if(CMAKE_CXX_COMPILER_ID MATCHES "GNU|Clang|AppleClang")
    target_compile_options(project_warnings INTERFACE
        -Wall                   # Warnings padrão essenciais
        -Wextra                 # Warnings adicionais além de -Wall
        -Wpedantic              # Exige conformidade estrita com o padrão ISO
        -Wshadow                # Variável local oculta outra de escopo externo
        -Wnon-virtual-dtor      # Destrutor não-virtual em classe com métodos virtuais
        -Wold-style-cast        # Cast estilo C (use static_cast/reinterpret_cast)
        -Wcast-align            # Cast que aumenta o alinhamento do ponteiro
        -Wunused                # Variáveis, funções e parâmetros não utilizados
        -Woverloaded-virtual    # Função virtual oculta sobrecarga da base
        -Wconversion            # Conversão implícita que pode perder dados
        -Wsign-conversion       # Conversão entre tipos com/sem sinal
        -Wmisleading-indentation # Indentação enganosa (sem chaves)
        -Wduplicated-cond       # Condição duplicada em if/else-if (GCC)
        -Wduplicated-branches   # Ramos idênticos em if/else (GCC)
        -Wlogical-op            # Operadores lógicos suspeitos (GCC)
        -Wnull-dereference      # Possível derreferência de ponteiro nulo
        -Wdouble-promotion      # Float promovido implicitamente para double
        -Wformat=2              # Verificações rigorosas de printf/scanf
        -Wimplicit-fallthrough  # Case sem break em switch (requer [[fallthrough]])
    )

    # Clang tem warnings adicionais úteis
    if(CMAKE_CXX_COMPILER_ID MATCHES "Clang|AppleClang")
        target_compile_options(project_warnings INTERFACE
            -Wno-unknown-warning-option # Ignora warnings não reconhecidos por versões antigas
        )
    endif()

    # GCC >= 8 suporta Wduplicated-cond e Wduplicated-branches
    # Versões mais antigas ignorarão silenciosamente via Wno-unknown-warning-option
endif()

# ── Flags para MSVC ───────────────────────────────────────────────────────────
if(MSVC)
    target_compile_options(project_warnings INTERFACE
        /W4         # Nível 4 de warnings (mais alto sem ser /Wall)
        /w14242     # Conversão: possível perda de dados
        /w14254     # Operador: possível perda de dados em conversão
        /w14263     # Função membro não sobrescreve virtual da classe base
        /w14265     # Classe com funções virtuais mas sem destrutor virtual
        /w14287     # Constante negativa usada como operando sem sinal
        /we4289     # Variável de loop usada fora do for
        /w14296     # Expressão sempre verdadeira ou falsa
        /w14311     # Truncamento de ponteiro: de 64-bit para 32-bit
        /w14545     # Expressão antes da vírgula sem efeito colateral
        /w14546     # Chamada de função antes da vírgula sem efeito colateral
        /w14547     # Operador sem efeito colateral antes da vírgula
        /w14549     # Operador sem efeito colateral antes da vírgula
        /w14555     # Expressão sem efeito colateral
        /w14619     # #pragma warning: número de warning inexistente
        /w14640     # Objeto local estático não é thread-safe
        /w14826     # Conversão implícita com sinal/sem sinal
        /w14905     # String literal convertida para LPSTR
        /w14906     # String literal convertida para LPWSTR
        /w14928     # Inicialização de cópia ilegal
        /permissive- # Desabilita extensões não-padrão do MSVC
    )
endif()

# ── Modo de tratamento de warnings ────────────────────────────────────────────
# Descomente para tratar todos os warnings como erros (-Werror / /WX).
# Recomendado em CI/CD, mas pode ser inconveniente durante desenvolvimento.
#
# if(CMAKE_CXX_COMPILER_ID MATCHES "GNU|Clang|AppleClang")
#     target_compile_options(project_warnings INTERFACE -Werror)
# elseif(MSVC)
#     target_compile_options(project_warnings INTERFACE /WX)
# endif()
`

// ─────────────────────────────────────────────────────────────────────────────
// Template: CMakePresets.json
// ─────────────────────────────────────────────────────────────────────────────

// tmplCMakePresets is the template for CMakePresets.json.
//
// CMakePresets.json (CMake 3.20+) is the modern standard for defining portable
// build configurations across developers, IDEs and CI systems.
//
// IDE support:
//   - VSCode CMake Tools: reads automatically and lists the presets
//   - CLion 2022.1+:      native support in the configuration UI
//   - Visual Studio 2019+: native support
//   - CLI:                cmake --preset <name>
//
// Generated presets:
//
//	Configure: base, debug, release, release-with-debug[, vcpkg-debug, vcpkg-release]
//	Build:     build-debug, build-release
//	Test:      test-debug, test-release
const tmplCMakePresets = `{
    "version": 6,
    "cmakeMinimumRequired": {
        "major": 3,
        "minor": 20,
        "patch": 0
    },

    "configurePresets": [
        {
            "name": "base",
            "hidden": true,
            "description": "Configurações base compartilhadas por todos os presets",
            "binaryDir": "${sourceDir}/build/${presetName}",
            "installDir": "${sourceDir}/install/${presetName}",
            "generator": "Ninja",
            "cacheVariables": {
                "CMAKE_EXPORT_COMPILE_COMMANDS": "ON",
                "{{.NameUpper}}_BUILD_TESTS": "ON"
            }
        },
        {
            "name": "debug",
            "displayName": "Debug",
            "description": "Build de desenvolvimento com símbolos de debug e sanitizers",
            "inherits": "base",
            "cacheVariables": {
                "CMAKE_BUILD_TYPE": "Debug",
                "CMAKE_CXX_FLAGS_DEBUG": "-g3 -O0 -fno-omit-frame-pointer"
            }
        },
        {
            "name": "release",
            "displayName": "Release",
            "description": "Build otimizado para produção",
            "inherits": "base",
            "cacheVariables": {
                "CMAKE_BUILD_TYPE": "Release",
                "{{.NameUpper}}_BUILD_TESTS": "OFF"
            }
        },
        {
            "name": "release-with-debug",
            "displayName": "RelWithDebInfo",
            "description": "Release com informações de debug (profiling)",
            "inherits": "base",
            "cacheVariables": {
                "CMAKE_BUILD_TYPE": "RelWithDebInfo"
            }
        },
        {
            "name": "sanitize",
            "displayName": "Debug + Sanitizers",
            "description": "Debug com AddressSanitizer e UndefinedBehaviorSanitizer",
            "inherits": "debug",
            "cacheVariables": {
                "CMAKE_CXX_FLAGS": "-fsanitize=address,undefined -fno-sanitize-recover=all"
            }
        }{{if .UseVCPKG}},
        {
            "name": "vcpkg-base",
            "hidden": true,
            "description": "Base para presets com VCPKG",
            "inherits": "base",
            "toolchainFile": "$env{VCPKG_ROOT}/scripts/buildsystems/vcpkg.cmake",
            "cacheVariables": {
                "VCPKG_BUILD_TYPE": "debug"
            }
        },
        {
            "name": "vcpkg-debug",
            "displayName": "VCPKG Debug",
            "description": "Debug com dependências VCPKG (requer VCPKG_ROOT)",
            "inherits": ["debug", "vcpkg-base"]
        },
        {
            "name": "vcpkg-release",
            "displayName": "VCPKG Release",
            "description": "Release com dependências VCPKG (requer VCPKG_ROOT)",
            "inherits": ["release", "vcpkg-base"],
            "cacheVariables": {
                "VCPKG_BUILD_TYPE": "release"
            }
        }{{end}}
    ],

    "buildPresets": [
        {
            "name": "build-debug",
            "displayName": "Build Debug",
            "configurePreset": "debug",
            "configuration": "Debug"
        },
        {
            "name": "build-release",
            "displayName": "Build Release",
            "configurePreset": "release",
            "configuration": "Release"
        },
        {
            "name": "build-sanitize",
            "displayName": "Build Sanitizers",
            "configurePreset": "sanitize",
            "configuration": "Debug"
        }{{if .UseVCPKG}},
        {
            "name": "build-vcpkg-debug",
            "displayName": "Build VCPKG Debug",
            "configurePreset": "vcpkg-debug",
            "configuration": "Debug"
        },
        {
            "name": "build-vcpkg-release",
            "displayName": "Build VCPKG Release",
            "configurePreset": "vcpkg-release",
            "configuration": "Release"
        }{{end}}
    ],

    "testPresets": [
        {
            "name": "test-debug",
            "displayName": "Testes Debug",
            "configurePreset": "debug",
            "configuration": "Debug",
            "output": {
                "outputOnFailure": true,
                "verbosity": "default"
            },
            "execution": {
                "stopOnFailure": false,
                "timeout": 60
            }
        },
        {
            "name": "test-release",
            "displayName": "Testes Release",
            "configurePreset": "release-with-debug",
            "configuration": "RelWithDebInfo",
            "output": {
                "outputOnFailure": true
            }
        }{{if .UseVCPKG}},
        {
            "name": "test-vcpkg-debug",
            "displayName": "Testes VCPKG Debug",
            "configurePreset": "vcpkg-debug",
            "configuration": "Debug",
            "output": {
                "outputOnFailure": true
            }
        }{{end}}
    ]
}
`
