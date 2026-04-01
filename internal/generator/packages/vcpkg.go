// Package packages contém os geradores de configuração de gerenciadores
// de pacotes C++ suportados pelo cpp-gen.
//
// Gerenciadores suportados:
//   - vcpkg.go         — Microsoft VCPKG (manifest mode via vcpkg.json)
//   - fetchcontent.go  — CMake FetchContent (dependências via cmake/Dependencies.cmake)
package packages

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// ─────────────────────────────────────────────────────────────────────────────
// GenerateVCPKG — ponto de entrada público
// ─────────────────────────────────────────────────────────────────────────────

// GenerateVCPKG gera os arquivos necessários para integrar o VCPKG ao projeto
// em modo manifesto (manifest mode), que é a forma recomendada para projetos
// modernos pois versiona as dependências junto com o código-fonte.
//
// Arquivos gerados:
//
//   - vcpkg.json           — manifesto de dependências do projeto
//   - vcpkg-configuration.json — configuração de baseline e registros
//   - cmake/Vcpkg.cmake    — módulo CMake auxiliar de documentação/fallback
//
// Modo manifesto (manifest mode):
//
//	No modo manifesto, o VCPKG lê o vcpkg.json na raiz do projeto e instala
//	as dependências automaticamente durante o configure do CMake. As dependências
//	ficam em vcpkg_installed/ (listado no .gitignore).
//
//	Para adicionar uma dependência:
//	  1. Adicione ao array "dependencies" no vcpkg.json
//	  2. Execute: cmake --preset debug  (instala automaticamente)
//	     Ou manualmente: vcpkg install
//
//	Requerimento: variável de ambiente VCPKG_ROOT deve estar definida,
//	ou o toolchain file deve ser passado via CMakePresets.json (já configurado).
//
// Referência: https://learn.microsoft.com/vcpkg/concepts/manifest-mode
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
			data:     nil, // sem dados de template — conteúdo estático
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
// Utilitários internos do pacote
// ─────────────────────────────────────────────────────────────────────────────

// writePkgFile cria (ou sobrescreve) um arquivo no caminho dado.
// Cria todos os diretórios pai necessários automaticamente.
// Se verbose for true, imprime o caminho do arquivo criado.
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

// renderPkgTemplate processa um template Go (text/template) com os dados
// fornecidos e retorna o resultado como string.
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

// tmplVCPKGManifest é o conteúdo do arquivo vcpkg.json.
//
// O vcpkg.json (manifesto) é o arquivo central do VCPKG em manifest mode.
// Ele declara:
//   - Metadados do projeto (name, version, description)
//   - Dependências diretas com versões mínimas opcionais
//   - Features opcionais que podem ser habilitadas pelo consumidor
//   - Overrides de versão para forçar uma versão específica de uma dep
//
// Boas práticas:
//   - Use "version>=" para especificar versão mínima (preferível a "version")
//   - Adicione "builtin-baseline" no vcpkg-configuration.json para reprodutibilidade
//   - Mantenha vcpkg_installed/ no .gitignore (gerado automaticamente)
//   - Versione vcpkg.json E vcpkg-configuration.json para builds reprodutíveis
//
// Para encontrar pacotes disponíveis:
//
//	vcpkg search <nome>
//	https://vcpkg.io/en/packages
//
// Referência: https://learn.microsoft.com/vcpkg/reference/vcpkg-json
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

// tmplVCPKGConfiguration é o conteúdo do arquivo vcpkg-configuration.json.
//
// O vcpkg-configuration.json define:
//   - builtin-baseline: hash de commit do repositório vcpkg que garante
//     que todos os desenvolvedores usem exatamente as mesmas versões
//     das dependências (reprodutibilidade de builds).
//   - registries: fontes adicionais de pacotes (registros privados ou
//     mirrors do registro oficial).
//
// IMPORTANTE: O campo "builtin-baseline" deve ser atualizado periodicamente
// para receber correções de segurança e novas versões das dependências.
//
// Para obter o baseline atual:
//
//	cd $VCPKG_ROOT && git rev-parse HEAD
//
// Para atualizar o baseline (e consequentemente as dependências):
//
//	cd $VCPKG_ROOT && git pull
//	# Copie o novo hash para "builtin-baseline" abaixo
//	# Execute: cmake --preset debug  (reinstala com as novas versões)
//
// Referência: https://learn.microsoft.com/vcpkg/reference/vcpkg-configuration-json
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

// tmplCMakeVcpkg é o conteúdo do arquivo cmake/Vcpkg.cmake.
//
// Este módulo CMake serve como:
//  1. Documentação das opções de integração VCPKG para a equipe
//  2. Fallback para configurar o toolchain quando CMakePresets.json não é usado
//  3. Helper para diagnóstico de problemas de configuração
//
// NOTA: A integração principal do VCPKG é feita via toolchain file no
// CMakePresets.json (já configurado pelo cmake.go). Este arquivo é
// complementar e não é incluído automaticamente — para usá-lo como fallback,
// inclua-o no CMakeLists.txt principal antes do primeiro project().
//
// IMPORTANTE: O toolchain file do VCPKG deve ser configurado ANTES da chamada
// project() no CMakeLists.txt. Com CMakePresets.json, isso é feito
// automaticamente via o campo "toolchainFile" nos presets.
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
