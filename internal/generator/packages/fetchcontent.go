// Package packages contém os geradores de configuração de gerenciadores
// de pacotes C++ suportados pelo cpp-gen.
package packages

import (
	"fmt"
	"path/filepath"
)

// ─────────────────────────────────────────────────────────────────────────────
// GenerateFetchContent — ponto de entrada público
// ─────────────────────────────────────────────────────────────────────────────

// GenerateFetchContent gera os arquivos necessários para gerenciar dependências
// C++ usando o módulo FetchContent nativo do CMake (CMake 3.11+).
//
// FetchContent é a abordagem "sem ferramentas externas" para dependências C++:
// o CMake baixa, configura e compila as dependências automaticamente durante
// o configure do projeto, sem necessidade de instalar nada previamente.
//
// Vantagens:
//   - Sem dependências externas (apenas CMake)
//   - Funciona em qualquer ambiente de CI/CD sem configuração prévia
//   - Controle total sobre as versões via tags/commits Git
//   - As dependências são compiladas junto com o projeto (mesmo compilador/flags)
//
// Desvantagens vs VCPKG:
//   - Recompila as dependências a cada `cmake --build` limpo
//   - Sem cache global entre projetos (cada projeto compila suas deps)
//   - Gerenciamento de versões manual (sem arquivo de lock automático)
//
// Arquivos gerados:
//
//   - cmake/Dependencies.cmake — declaração de todas as dependências via
//     FetchContent_Declare() e FetchContent_MakeAvailable()
//
// Uso no CMakeLists.txt (já configurado pelo cmake.go):
//
//	include(FetchContent)
//	include(cmake/Dependencies.cmake)
//
// Referência: https://cmake.org/cmake/help/latest/module/FetchContent.html
func GenerateFetchContent(root string, verbose bool) error {
	depsPath := filepath.Join(root, "cmake", "Dependencies.cmake")

	if err := writePkgFile(depsPath, tmplFetchContentDependencies, verbose); err != nil {
		return fmt.Errorf("gerar cmake/Dependencies.cmake: %w", err)
	}

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Template: cmake/Dependencies.cmake
// ─────────────────────────────────────────────────────────────────────────────

// tmplFetchContentDependencies é o conteúdo do arquivo cmake/Dependencies.cmake.
//
// Este arquivo centraliza TODAS as dependências externas do projeto declaradas
// via FetchContent. A separação em um arquivo dedicado mantém o CMakeLists.txt
// raiz limpo e facilita o gerenciamento de dependências.
//
// Padrão recomendado para cada dependência:
//
//	FetchContent_Declare(
//	    <nome>                         # identificador único (lowercase)
//	    GIT_REPOSITORY <url>           # URL do repositório Git
//	    GIT_TAG        <tag/commit>    # tag, branch ou commit hash (SHA para reprodutibilidade)
//	    GIT_SHALLOW    TRUE            # baixa apenas o commit especificado (mais rápido)
//	)
//	FetchContent_MakeAvailable(<nome>) # baixa, configura e disponibiliza o target
//
// Após FetchContent_MakeAvailable(), use o target no seu CMakeLists.txt:
//
//	target_link_libraries(${PROJECT_NAME} PRIVATE fmt::fmt)
//
// Para velocidade em CI/CD, considere usar CPM.cmake como wrapper do FetchContent:
//
//	https://github.com/cpm-cmake/CPM.cmake
//
// Referência: https://cmake.org/cmake/help/latest/module/FetchContent.html
const tmplFetchContentDependencies = `# =============================================================================
# cmake/Dependencies.cmake — Dependências externas via FetchContent
# =============================================================================
#
# Este arquivo é incluído automaticamente pelo CMakeLists.txt raiz:
#   include(FetchContent)
#   include(cmake/Dependencies.cmake)
#
# Para adicionar uma dependência:
#   1. Descomente ou adicione um bloco FetchContent_Declare() abaixo
#   2. Chame FetchContent_MakeAvailable(<nome>)
#   3. Adicione o target em target_link_libraries() no src/CMakeLists.txt
#
# Documentação completa: https://cmake.org/cmake/help/latest/module/FetchContent.html
# =============================================================================

# ── Configurações globais do FetchContent ─────────────────────────────────────

# Desabilita a exibição de mensagens de progresso das dependências no configure.
# Remova ou comente para ver o progresso detalhado de download/configure.
set(FETCHCONTENT_QUIET ON)

# Desabilita a atualização automática das dependências já baixadas.
# Quando ON, o CMake não busca por novas versões após o primeiro download.
# Recomendado para builds reprodutíveis em CI/CD.
# Defina como OFF durante desenvolvimento se quiser atualizações automáticas.
set(FETCHCONTENT_UPDATES_DISCONNECTED ON)

# Diretório onde as dependências serão armazenadas.
# Por padrão: ${CMAKE_BINARY_DIR}/_deps
# Personalize para compartilhar o cache entre builds:
# set(FETCHCONTENT_BASE_DIR "${CMAKE_SOURCE_DIR}/.cmake-deps")

# ── Função auxiliar ───────────────────────────────────────────────────────────

# declare_dep() é um wrapper conveniente em torno de FetchContent_Declare
# que adiciona mensagens de status formatadas durante o configure.
#
# Uso:
#   declare_dep(fmt GIT_REPOSITORY https://github.com/fmtlib/fmt GIT_TAG 10.2.1)
#   FetchContent_MakeAvailable(fmt)
#
macro(declare_dep NAME)
    message(STATUS "FetchContent: declarando dependência '${NAME}'")
    FetchContent_Declare(${NAME} ${ARGN})
endmacro()

# ── Dependências ──────────────────────────────────────────────────────────────
# Descomente as dependências que deseja utilizar no projeto.
# Para builds reprodutíveis, SEMPRE use um commit hash completo (SHA-1) em vez
# de um nome de branch (main, master) que pode mudar ao longo do tempo.
# =============================================================================

# =============================================================================
# {fmt} — Formatação de strings moderna e type-safe
# =============================================================================
# Uso: #include <fmt/core.h>
#      fmt::print("Hello, {}!\n", name);
# CMake target: fmt::fmt
# Docs: https://fmt.dev
# =============================================================================
# declare_dep(fmt
#     GIT_REPOSITORY https://github.com/fmtlib/fmt.git
#     GIT_TAG        10.2.1   # ou: e69e5f977d458f2650bb346dadf2ad30c5320281
#     GIT_SHALLOW    TRUE
# )
# # Desabilita testes e docs do fmt para acelerar o build
# set(FMT_DOC   OFF CACHE BOOL "" FORCE)
# set(FMT_TEST  OFF CACHE BOOL "" FORCE)
# FetchContent_MakeAvailable(fmt)

# =============================================================================
# spdlog — Logging assíncrono de alta performance
# =============================================================================
# Uso: #include <spdlog/spdlog.h>
#      spdlog::info("Hello, {}!", name);
# CMake target: spdlog::spdlog
# Docs: https://github.com/gabime/spdlog
# =============================================================================
# declare_dep(spdlog
#     GIT_REPOSITORY https://github.com/gabime/spdlog.git
#     GIT_TAG        v1.13.0
#     GIT_SHALLOW    TRUE
# )
# # Usa o {fmt} já baixado como backend do spdlog (evita download duplicado)
# set(SPDLOG_FMT_EXTERNAL     ON  CACHE BOOL "" FORCE)
# set(SPDLOG_BUILD_EXAMPLES   OFF CACHE BOOL "" FORCE)
# set(SPDLOG_BUILD_BENCH      OFF CACHE BOOL "" FORCE)
# set(SPDLOG_BUILD_TESTS      OFF CACHE BOOL "" FORCE)
# FetchContent_MakeAvailable(spdlog)

# =============================================================================
# nlohmann/json — Biblioteca header-only de JSON para C++ moderno
# =============================================================================
# Uso: #include <nlohmann/json.hpp>
#      nlohmann::json j = {{"key", "value"}};
# CMake target: nlohmann_json::nlohmann_json
# Docs: https://json.nlohmann.me
# =============================================================================
# declare_dep(nlohmann_json
#     GIT_REPOSITORY https://github.com/nlohmann/json.git
#     GIT_TAG        v3.11.3
#     GIT_SHALLOW    TRUE
# )
# # Instala apenas os headers, sem compilar testes ou exemplos.
# set(JSON_BuildTests      OFF CACHE BOOL "" FORCE)
# set(JSON_Install         OFF CACHE BOOL "" FORCE)
# set(JSON_MultipleHeaders OFF CACHE BOOL "" FORCE)
# FetchContent_MakeAvailable(nlohmann_json)

# =============================================================================
# CLI11 — Parser de argumentos de linha de comando (header-only)
# =============================================================================
# Uso: #include <CLI/CLI.hpp>
#      CLI::App app{"Descrição"};
#      app.add_option("--name", name, "Nome");
#      CLI11_PARSE(app, argc, argv);
# CMake target: CLI11::CLI11
# Docs: https://cliutils.gitlab.io/CLI11Tutorial
# =============================================================================
# declare_dep(cli11
#     GIT_REPOSITORY https://github.com/CLIUtils/CLI11.git
#     GIT_TAG        v2.4.1
#     GIT_SHALLOW    TRUE
# )
# set(CLI11_BUILD_TESTS   OFF CACHE BOOL "" FORCE)
# set(CLI11_BUILD_EXAMPLES OFF CACHE BOOL "" FORCE)
# FetchContent_MakeAvailable(cli11)

# =============================================================================
# Catch2 — Framework de testes BDD/TDD moderno e header-only
# =============================================================================
# Uso: #define CATCH_CONFIG_MAIN
#      #include <catch2/catch_all.hpp>
#      TEST_CASE("nome", "[tag]") { REQUIRE(1 + 1 == 2); }
# CMake target: Catch2::Catch2WithMain  (ou Catch2::Catch2 para main próprio)
# Docs: https://github.com/catchorg/Catch2/blob/devel/docs/cmake-integration.md
# =============================================================================
# declare_dep(Catch2
#     GIT_REPOSITORY https://github.com/catchorg/Catch2.git
#     GIT_TAG        v3.5.2
#     GIT_SHALLOW    TRUE
# )
# set(CATCH_BUILD_TESTING OFF CACHE BOOL "" FORCE)
# set(CATCH_INSTALL_DOCS  OFF CACHE BOOL "" FORCE)
# FetchContent_MakeAvailable(Catch2)
# # Habilita integração automática com CTest via catch_discover_tests()
# include(Catch)

# =============================================================================
# GoogleTest — Framework de testes da Google (amplamente adotado)
# =============================================================================
# Uso: #include <gtest/gtest.h>
#      TEST(SuiteNome, TestNome) { EXPECT_EQ(1 + 1, 2); }
# CMake targets: GTest::gtest, GTest::gtest_main, GTest::gmock
# Docs: https://google.github.io/googletest/
# =============================================================================
# declare_dep(googletest
#     GIT_REPOSITORY https://github.com/google/googletest.git
#     GIT_TAG        v1.14.0
#     GIT_SHALLOW    TRUE
# )
# # Evita sobrescrever as configurações do compilador no Windows
# set(gtest_force_shared_crt ON CACHE BOOL "" FORCE)
# set(BUILD_GMOCK             ON CACHE BOOL "" FORCE)
# set(INSTALL_GTEST           OFF CACHE BOOL "" FORCE)
# FetchContent_MakeAvailable(googletest)
# # Habilita integração automática com CTest via gtest_discover_tests()
# include(GoogleTest)

# =============================================================================
# doctest — Framework de testes ultraleve e header-only
# =============================================================================
# Uso: #define DOCTEST_CONFIG_IMPLEMENT_WITH_MAIN
#      #include <doctest/doctest.h>
#      TEST_CASE("nome") { CHECK(1 + 1 == 2); }
# CMake target: doctest::doctest
# Docs: https://github.com/doctest/doctest
# =============================================================================
# declare_dep(doctest
#     GIT_REPOSITORY https://github.com/doctest/doctest.git
#     GIT_TAG        v2.4.11
#     GIT_SHALLOW    TRUE
# )
# FetchContent_MakeAvailable(doctest)

# =============================================================================
# Eigen3 — Álgebra linear de alta performance (header-only)
# =============================================================================
# Uso: #include <Eigen/Dense>
#      Eigen::Matrix3d m = Eigen::Matrix3d::Identity();
# CMake target: Eigen3::Eigen
# Docs: https://eigen.tuxfamily.org/dox/
# =============================================================================
# declare_dep(Eigen3
#     GIT_REPOSITORY https://gitlab.com/libeigen/eigen.git
#     GIT_TAG        3.4.0
#     GIT_SHALLOW    TRUE
# )
# set(BUILD_TESTING    OFF CACHE BOOL "" FORCE)
# set(EIGEN_BUILD_DOC  OFF CACHE BOOL "" FORCE)
# set(EIGEN_BUILD_PKGCONFIG OFF CACHE BOOL "" FORCE)
# FetchContent_MakeAvailable(Eigen3)

# =============================================================================
# Google Benchmark — Microbenchmarking de alta precisão
# =============================================================================
# Uso: #include <benchmark/benchmark.h>
#      BENCHMARK(BM_MinhaFuncao)->Range(8, 8<<10);
#      BENCHMARK_MAIN();
# CMake target: benchmark::benchmark, benchmark::benchmark_main
# Docs: https://github.com/google/benchmark
# =============================================================================
# declare_dep(benchmark
#     GIT_REPOSITORY https://github.com/google/benchmark.git
#     GIT_TAG        v1.8.3
#     GIT_SHALLOW    TRUE
# )
# set(BENCHMARK_ENABLE_TESTING    OFF CACHE BOOL "" FORCE)
# set(BENCHMARK_ENABLE_GTEST_TESTS OFF CACHE BOOL "" FORCE)
# set(BENCHMARK_ENABLE_INSTALL    OFF CACHE BOOL "" FORCE)
# FetchContent_MakeAvailable(benchmark)

# =============================================================================
# absl (Abseil) — Biblioteca de utilitários C++ do Google
# =============================================================================
# Uso: #include <absl/strings/str_format.h>
#      absl::StrFormat("Hello, %s!", name);
# CMake targets: absl::strings, absl::status, absl::time, etc.
# Docs: https://abseil.io/docs/cpp
# =============================================================================
# declare_dep(absl
#     GIT_REPOSITORY https://github.com/abseil/abseil-cpp.git
#     GIT_TAG        20240116.2
#     GIT_SHALLOW    TRUE
# )
# set(ABSL_PROPAGATE_CXX_STD ON CACHE BOOL "" FORCE)
# set(ABSL_BUILD_TESTING     OFF CACHE BOOL "" FORCE)
# FetchContent_MakeAvailable(absl)

# =============================================================================
# Resumo das dependências ativas
# =============================================================================
# Adicione aqui uma mensagem de status listando as dependências ativas,
# para facilitar o diagnóstico durante o configure.
# =============================================================================

message(STATUS "")
message(STATUS "=== Dependências (FetchContent) ===")
# message(STATUS "  fmt          : habilitado")
# message(STATUS "  spdlog       : habilitado")
# message(STATUS "  nlohmann_json: habilitado")
message(STATUS "  (nenhuma dependência ativa — descomente as desejadas acima)")
message(STATUS "")
`
