cmake_minimum_required(VERSION 3.20)

# =============================================================================
# {{.Name}}
# =============================================================================
# Gerado a partir de um template customizado do cpp-gen. As variáveis abaixo
# ({{.Name}}, {{.Standard}}, etc.) são substituídas quando alguém roda
# `cpp-gen new <projeto> --template <esta-pasta>`. Veja TEMPLATE_GUIDE.md
# pra referência completa das variáveis disponíveis.

project(
    {{.Name}}
    VERSION {{.Version}}
    DESCRIPTION "{{if .Description}}{{.Description}}{{else}}Projeto gerado a partir de um template customizado do cpp-gen{{end}}"
    LANGUAGES CXX
)

# Impede builds in-source, que misturam artefatos com o código fonte.
if(CMAKE_SOURCE_DIR STREQUAL CMAKE_BINARY_DIR)
    message(FATAL_ERROR "Build in-source detectado! Use: cmake -B build")
endif()

set(CMAKE_CXX_STANDARD {{.Standard}})
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/bin")

# =============================================================================
# Dependências — adapte ao gerenciador de pacotes escolhido em `cpp-gen new`
# =============================================================================
# Apague este bloco (e o arquivo vcpkg.json condicional abaixo) se seu
# template não tiver dependências externas.
{{if .UseVCPKG}}
# VCPKG ativo — dependências declaradas em vcpkg.json.
if(DEFINED ENV{VCPKG_ROOT} AND NOT DEFINED CMAKE_TOOLCHAIN_FILE)
    set(CMAKE_TOOLCHAIN_FILE
        "$ENV{VCPKG_ROOT}/scripts/buildsystems/vcpkg.cmake"
        CACHE STRING "VCPKG toolchain file"
    )
endif()
# find_package(alguma-lib CONFIG REQUIRED)
{{else if .UseFetchContent}}
# FetchContent ativo — declare dependências aqui com include(FetchContent) +
# FetchContent_Declare(...) + FetchContent_MakeAvailable(...).
{{else}}
# Nenhum gerenciador de pacotes selecionado — assuma pacotes do sistema, se
# este template precisar de algum (find_package(... REQUIRED)).
{{end}}

# =============================================================================
# Executável
# =============================================================================

add_executable({{.Name}} src/{{.NameSnake}}.cpp)
target_compile_features({{.Name}} PRIVATE cxx_std_{{.Standard}})

message(STATUS "{{.Name}} v${PROJECT_VERSION} — C++${CMAKE_CXX_STANDARD}")
