// Package packages contains the configuration generators for package
// managers for C++ supported by cpp-gen.
//
// Supported managers:
//   - vcpkg.go         — Microsoft VCPKG (manifest mode via vcpkg.json)
//   - fetchcontent.go  — CMake FetchContent (dependencies via cmake/Dependencies.cmake)
package packages

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// ─────────────────────────────────────────────────────────────────────────────
// GenerateVCPKG — public entry point
// ─────────────────────────────────────────────────────────────────────────────

// GenerateVCPKG generates the necessary files to integrate VCPKG into the project
// in manifest mode, which is the recommended approach for modern projects
// as it versions dependencies alongside source code.
//
// Generated files:
//
//   - vcpkg.json           — project dependencies manifest
//   - vcpkg-configuration.json — baseline and registry configuration
//   - cmake/Vcpkg.cmake    — auxiliary CMake module for documentation/fallback
//
// Manifest mode:
//
//	In manifest mode, VCPKG reads vcpkg.json at the project root and installs
//	dependencies automatically during CMake configure. Dependencies
//	are stored in vcpkg_installed/ (listed in .gitignore).
//
//	To add a dependency:
//	  1. Add to the "dependencies" array in vcpkg.json
//	  2. Run: cmake --preset debug  (installs automatically)
//	     Or manually: vcpkg install
//
//	Requirement: the VCPKG_ROOT environment variable must be defined,
//	or the toolchain file must be passed via CMakePresets.json (already configured).
//
// Reference: https://learn.microsoft.com/vcpkg/concepts/manifest-mode
func GenerateVCPKG(root string, verbose bool) error {
	steps := []struct {
		name     string
		relPath  string
		tmplName string
		tmpl     string
		data     any
	}{
		{
			name:     "vcpkg.json",
			relPath:  "vcpkg.json",
			tmplName: "vcpkg_manifest",
			tmpl:     tmplVCPKGManifest,
			data:     nil, // no template data — static content
		},
		{
			name:     "vcpkg-configuration.json",
			relPath:  "vcpkg-configuration.json",
			tmplName: "vcpkg_configuration",
			tmpl:     tmplVCPKGConfiguration,
			data:     nil,
		},
		{
			name:     "cmake/Vcpkg.cmake",
			relPath:  filepath.Join("cmake", "Vcpkg.cmake"),
			tmplName: "cmake_vcpkg",
			tmpl:     tmplCMakeVcpkg,
			data:     nil,
		},
	}

	for _, s := range steps {
		path := filepath.Join(root, s.relPath)
		var content string
		var err error

		if s.data != nil {
			content, err = renderPkgTemplate(s.tmplName, s.tmpl, s.data)
			if err != nil {
				return fmt.Errorf("renderizar template %s: %w", s.name, err)
			}
		} else {
			content = s.tmpl
		}

		if err := writePkgFile(path, content, verbose); err != nil {
			return fmt.Errorf("gerar %s: %w", s.name, err)
		}
	}

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Internal package utilities
// ─────────────────────────────────────────────────────────────────────────────

// writePkgFile creates (or overwrites) a file at the given path.
// Creates all necessary parent directories automatically.
// If verbose is true, prints the path of the created file.
func writePkgFile(path, content string, verbose bool) error {
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

// renderPkgTemplate processes a Go template (text/template) with the provided
// data and returns the result as a string.
func renderPkgTemplate(name, tmpl string, data any) (string, error) {
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

// ─────────────────────────────────────────────────────────────────────────────
// Template: vcpkg.json
// ─────────────────────────────────────────────────────────────────────────────

// tmplVCPKGManifest is the content of the vcpkg.json file.
//
// The vcpkg.json (manifest) is the central VCPKG file in manifest mode.
// It declares:
//   - Project metadata (name, version, description)
//   - Direct dependencies with optional minimum versions
//   - Optional features that can be enabled by the consumer
//   - Version overrides to force a specific version of a dependency
//
// Best practices:
//   - Use "version>=" to specify minimum version (preferable to "version")
//   - Add "builtin-baseline" in vcpkg-configuration.json for reproducibility
//   - Keep vcpkg_installed/ in .gitignore (generated automatically)
//   - Version both vcpkg.json AND vcpkg-configuration.json for reproducible builds
//
// To find available packages:
//
//	vcpkg search <name>
//	https://vcpkg.io/en/packages
//
// Reference: https://learn.microsoft.com/vcpkg/reference/vcpkg-json
const tmplVCPKGManifest = `{
    "$schema": "https://raw.githubusercontent.com/microsoft/vcpkg-tool/main/docs/vcpkg.schema.json",

    "name": "my-project",
    "version": "1.0.0",
    "description": "Projeto C++ com VCPKG",

    "dependencies": [
        // ── Exemplos de dependências populares ────────────────────────────────
        // Descomente as dependências que deseja utilizar e execute:
        //   cmake --preset debug   (instala automaticamente via manifest mode)
        //
        // ── Utilitários gerais ────────────────────────────────────────────────

        // {fmt} — Formatação de strings moderna, alternativa ao printf/iostream
        // https://github.com/fmtlib/fmt
        // {
        //     "name": "fmt",
        //     "version>=": "10.0.0"
        // },

        // spdlog — Logging assíncrono de alta performance
        // https://github.com/gabime/spdlog
        // {
        //     "name": "spdlog",
        //     "version>=": "1.12.0"
        // },

        // ── JSON ──────────────────────────────────────────────────────────────

        // nlohmann-json — Biblioteca header-only de JSON para C++
        // https://github.com/nlohmann/json
        // {
        //     "name": "nlohmann-json",
        //     "version>=": "3.11.0"
        // },

        // ── Linha de comando ──────────────────────────────────────────────────

        // CLI11 — Parser de argumentos de linha de comando
        // https://github.com/CLIUtils/CLI11
        // {
        //     "name": "cli11",
        //     "version>=": "2.3.0"
        // },

        // ── Testes ────────────────────────────────────────────────────────────

        // Catch2 — Framework de testes BDD/TDD moderno para C++
        // https://github.com/catchorg/Catch2
        // {
        //     "name": "catch2",
        //     "version>=": "3.4.0"
        // },

        // GoogleTest — Framework de testes do Google, muito usado em projetos grandes
        // https://github.com/google/googletest
        // {
        //     "name": "gtest",
        //     "version>=": "1.14.0"
        // },

        // doctest — Framework de testes leve e header-only
        // https://github.com/doctest/doctest
        // {
        //     "name": "doctest",
        //     "version>=": "2.4.11"
        // },

        // ── Rede ──────────────────────────────────────────────────────────────

        // libcurl — Transferência de dados com suporte a múltiplos protocolos
        // https://curl.se/libcurl/
        // {
        //     "name": "curl",
        //     "version>=": "8.0.0",
        //     "features": ["ssl"]
        // },

        // ── Banco de dados ────────────────────────────────────────────────────

        // SQLite3 — Banco de dados embarcado leve e sem servidor
        // https://www.sqlite.org/
        // {
        //     "name": "sqlite3",
        //     "version>=": "3.43.0"
        // },

        // ── Matemática e ciência ──────────────────────────────────────────────

        // Eigen3 — Álgebra linear de alta performance (header-only)
        // https://eigen.tuxfamily.org/
        // {
        //     "name": "eigen3",
        //     "version>=": "3.4.0"
        // },

        // ── Compressão ────────────────────────────────────────────────────────

        // zlib — Compressão de dados padrão da indústria
        // https://zlib.net/
        // {
        //     "name": "zlib",
        //     "version>=": "1.3.0"
        // }
    ],

    // ── Features opcionais ────────────────────────────────────────────────────
    // Features permitem habilitar dependências condicionalmente.
    // Para usar: cmake --preset debug -DVCPKG_MANIFEST_FEATURES="tests;docs"
    "features": {
        "tests": {
            "description": "Habilita dependências de framework de testes",
            "dependencies": [
                // Descomente o framework de testes preferido:
                // { "name": "catch2",  "version>=": "3.4.0" },
                // { "name": "gtest",   "version>=": "1.14.0" },
                // { "name": "doctest", "version>=": "2.4.11" }
            ]
        },
        "tools": {
            "description": "Ferramentas adicionais de desenvolvimento",
            "dependencies": [
                // { "name": "benchmark", "version>=": "1.8.0" }
            ]
        }
    }
}
`

// ─────────────────────────────────────────────────────────────────────────────
// Template: vcpkg-configuration.json
// ─────────────────────────────────────────────────────────────────────────────

// tmplVCPKGConfiguration is the content of the vcpkg-configuration.json file.
//
// The vcpkg-configuration.json defines:
//   - builtin-baseline: commit hash of the vcpkg repository that ensures
//     that all developers use exactly the same versions
//     of dependencies (build reproducibility).
//   - registries: additional package sources (private registries or
//     mirrors of the official registry).
//
// IMPORTANT: The "builtin-baseline" field must be updated periodically
// to receive security fixes and new versions of dependencies.
//
// To get the current baseline:
//
//	cd $VCPKG_ROOT && git rev-parse HEAD
//
// To update the baseline (and consequently the dependencies):
//
//	cd $VCPKG_ROOT && git pull
//	# Copy the new hash to "builtin-baseline" below
//	# Run: cmake --preset debug  (reinstalls with the new versions)
//
// Reference: https://learn.microsoft.com/vcpkg/reference/vcpkg-configuration-json
const tmplVCPKGConfiguration = `{
    "$schema": "https://raw.githubusercontent.com/microsoft/vcpkg-tool/main/docs/vcpkg-configuration.schema.json",

    // ── Baseline de versões ───────────────────────────────────────────────────
    // O builtin-baseline é um commit hash do repositório microsoft/vcpkg que
    // "congela" as versões de TODOS os pacotes do registro oficial.
    //
    // Isso garante builds reprodutíveis: todo desenvolvedor que clonar o projeto
    // obterá exatamente as mesmas versões das dependências, independente de quando
    // fizer o clone.
    //
    // !! AÇÃO NECESSÁRIA: Substitua este valor pelo hash atual do seu vcpkg !!
    //
    // Para obter o hash atual:
    //   cd $VCPKG_ROOT && git rev-parse HEAD
    //
    // Hash de exemplo (substitua pelo atual):
    "builtin-baseline": "a34c873a9717a888f58dc05268dea15592c2f0ff",

    // ── Registros adicionais ──────────────────────────────────────────────────
    // Por padrão, o VCPKG usa apenas o registro oficial (github.com/microsoft/vcpkg).
    // Adicione registros privados ou espelhos aqui se necessário.
    //
    // Exemplo de registro privado:
    // "registries": [
    //     {
    //         "kind": "git",
    //         "repository": "https://github.com/sua-org/vcpkg-registry",
    //         "baseline": "abc123...",
    //         "packages": ["seu-pacote-privado"]
    //     }
    // ],

    // ── Overlays ──────────────────────────────────────────────────────────────
    // Overlays permitem sobrescrever ou adicionar ports locais sem modificar
    // o repositório oficial do VCPKG. Úteis para patches ou ports não publicados.
    //
    // "overlay-ports": ["./vcpkg-ports"],
    // "overlay-triplets": ["./vcpkg-triplets"],

    // ── Triplet padrão ────────────────────────────────────────────────────────
    // Define a plataforma alvo padrão para compilação das dependências.
    // Valores comuns:
    //   x64-linux, x64-osx, x64-windows, arm64-osx, x64-windows-static
    //
    // Se não definido, o VCPKG detecta automaticamente a plataforma host.
    // "default-triplet": "x64-linux",
    // "host-triplet": "x64-linux"
}
`

// ─────────────────────────────────────────────────────────────────────────────
// Template: cmake/Vcpkg.cmake
// ─────────────────────────────────────────────────────────────────────────────

// tmplCMakeVcpkg is the content of the cmake/Vcpkg.cmake file.
//
// This CMake module serves as:
//  1. Documentation of VCPKG integration options for the team
//  2. Fallback to configure the toolchain when CMakePresets.json is not used
//  3. Helper for diagnosing configuration issues
//
// NOTE: The main VCPKG integration is done via toolchain file in
// CMakePresets.json (already configured by cmake.go). This file is
// complementary and not included automatically — to use it as fallback,
// include it in the main CMakeLists.txt before the first project().
//
// IMPORTANT: The VCPKG toolchain file must be configured BEFORE the
// project() call in CMakeLists.txt. With CMakePresets.json, this is done
// automatically via the "toolchainFile" field in the presets.
const tmplCMakeVcpkg = `# =============================================================================
# cmake/Vcpkg.cmake — Módulo auxiliar de integração VCPKG
# =============================================================================
#
# NOTA SOBRE USO:
# A integração principal do VCPKG neste projeto é feita via CMakePresets.json,
# através do campo "toolchainFile" nos presets de configure.
#
# Este arquivo serve como:
#   1. Documentação e referência para a equipe
#   2. Fallback para ambientes sem CMakePresets.json
#   3. Helper de diagnóstico para problemas de configuração
#
# Para usar como fallback (sem presets), inclua ANTES de project() no
# CMakeLists.txt principal:
#   include(cmake/Vcpkg.cmake)   # ANTES de project()!
#
# Referência: https://learn.microsoft.com/vcpkg/users/buildsystems/cmake-integration
# =============================================================================

# ── Verificação de pré-condições ──────────────────────────────────────────────

# Este módulo só deve ser processado uma vez (proteção contra inclusão dupla).
if(DEFINED _CPP_GEN_VCPKG_INCLUDED)
    return()
endif()
set(_CPP_GEN_VCPKG_INCLUDED TRUE)

# ── Configuração do Toolchain ─────────────────────────────────────────────────

# Tenta localizar o VCPKG_ROOT de múltiplas fontes, em ordem de prioridade:
#   1. Variável CMake: -DVCPKG_ROOT=/caminho/para/vcpkg
#   2. Variável de ambiente: export VCPKG_ROOT=/caminho/para/vcpkg
#   3. Locais comuns de instalação
if(NOT DEFINED VCPKG_ROOT)
    if(DEFINED ENV{VCPKG_ROOT})
        set(VCPKG_ROOT "$ENV{VCPKG_ROOT}" CACHE PATH "Raiz da instalação do VCPKG")
        message(STATUS "VCPKG: encontrado via variável de ambiente VCPKG_ROOT=${VCPKG_ROOT}")
    else()
        # Verifica locais padrão de instalação
        set(_vcpkg_default_paths
            "$ENV{HOME}/vcpkg"
            "$ENV{USERPROFILE}/vcpkg"
            "/opt/vcpkg"
            "C:/vcpkg"
            "C:/src/vcpkg"
        )
        foreach(_path IN LISTS _vcpkg_default_paths)
            if(EXISTS "${_path}/scripts/buildsystems/vcpkg.cmake")
                set(VCPKG_ROOT "${_path}" CACHE PATH "Raiz da instalação do VCPKG")
                message(STATUS "VCPKG: encontrado em localização padrão: ${VCPKG_ROOT}")
                break()
            endif()
        endforeach()
    endif()
endif()

# ── Configuração do Toolchain File ────────────────────────────────────────────

if(DEFINED VCPKG_ROOT)
    set(_vcpkg_toolchain "${VCPKG_ROOT}/scripts/buildsystems/vcpkg.cmake")

    if(EXISTS "${_vcpkg_toolchain}")
        # Só define CMAKE_TOOLCHAIN_FILE se ainda não foi definido por outro meio.
        # O CMakePresets.json tem precedência quando usado via preset.
        if(NOT DEFINED CMAKE_TOOLCHAIN_FILE)
            set(CMAKE_TOOLCHAIN_FILE "${_vcpkg_toolchain}"
                CACHE STRING "VCPKG Toolchain File"
                FORCE
            )
            message(STATUS "VCPKG: toolchain configurado: ${CMAKE_TOOLCHAIN_FILE}")
        else()
            message(STATUS "VCPKG: CMAKE_TOOLCHAIN_FILE já definido externamente: ${CMAKE_TOOLCHAIN_FILE}")
        endif()
    else()
        message(WARNING
            "VCPKG: VCPKG_ROOT definido como '${VCPKG_ROOT}', mas o toolchain "
            "file não foi encontrado em:\n  ${_vcpkg_toolchain}\n"
            "Verifique se o VCPKG foi corretamente instalado e faça o bootstrap:\n"
            "  Linux/macOS: ${VCPKG_ROOT}/bootstrap-vcpkg.sh\n"
            "  Windows:     ${VCPKG_ROOT}\\bootstrap-vcpkg.bat"
        )
    endif()
else()
    message(STATUS
        "VCPKG: VCPKG_ROOT não encontrado. Para habilitar o VCPKG, defina:\n"
        "  Via CMake:    cmake -DVCPKG_ROOT=/caminho/para/vcpkg\n"
        "  Via ambiente: export VCPKG_ROOT=/caminho/para/vcpkg\n"
        "  Via preset:   cmake --preset vcpkg-debug  (recomendado)\n"
        "\n"
        "Para instalar o VCPKG:\n"
        "  git clone https://github.com/microsoft/vcpkg.git ~/vcpkg\n"
        "  ~/vcpkg/bootstrap-vcpkg.sh"
    )
endif()

# ── Configurações opcionais do VCPKG ─────────────────────────────────────────

# Habilita o manifest mode (lê vcpkg.json na raiz do projeto).
# Quando ativo, o VCPKG instala automaticamente as dependências listadas
# no vcpkg.json durante o configure do CMake.
# Desabilite apenas se quiser gerenciar as dependências manualmente via CLI.
set(VCPKG_MANIFEST_MODE ON CACHE BOOL "Habilita VCPKG manifest mode (vcpkg.json)")

# Desabilita features opcionais por padrão (apenas dependências base são instaladas).
# Para habilitar features: cmake -DVCPKG_MANIFEST_FEATURES="tests;tools"
if(NOT DEFINED VCPKG_MANIFEST_FEATURES)
    set(VCPKG_MANIFEST_FEATURES "" CACHE STRING "Features do VCPKG a habilitar (separadas por ;)")
endif()

# ── Diagnóstico ───────────────────────────────────────────────────────────────

# Exibe um resumo da configuração VCPKG ao final do cmake configure.
cmake_language(DEFER CALL message STATUS
    "VCPKG configuração:\n"
    "  VCPKG_ROOT            : ${VCPKG_ROOT}\n"
    "  CMAKE_TOOLCHAIN_FILE  : ${CMAKE_TOOLCHAIN_FILE}\n"
    "  VCPKG_MANIFEST_MODE   : ${VCPKG_MANIFEST_MODE}\n"
    "  VCPKG_MANIFEST_FEATURES: ${VCPKG_MANIFEST_FEATURES}"
)
`
