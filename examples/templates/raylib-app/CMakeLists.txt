cmake_minimum_required(VERSION 3.20)

# =============================================================================
# Declaração do projeto
# =============================================================================
project(
    {{.Name}}
    VERSION {{.Version}}
    DESCRIPTION "{{if .Description}}{{.Description}}{{else}}Projeto raylib gerado pelo cpp-gen{{end}}"
    LANGUAGES CXX
)

# Impede builds in-source, que misturam artefatos com o código fonte.
if(CMAKE_SOURCE_DIR STREQUAL CMAKE_BINARY_DIR)
    message(FATAL_ERROR "Build in-source detectado! Use: cmake -B build")
endif()

# =============================================================================
# Padrão e configurações C++
# =============================================================================

set(CMAKE_CXX_STANDARD {{.Standard}})
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/bin")
{{if .UseVCPKG}}
# =============================================================================
# Integração VCPKG
# =============================================================================
# Toolchain aplicado via CMakePresets.json (presets vcpkg-*) ou manualmente:
#   cmake -B build -DCMAKE_TOOLCHAIN_FILE=$VCPKG_ROOT/scripts/buildsystems/vcpkg.cmake
if(DEFINED ENV{VCPKG_ROOT} AND NOT DEFINED CMAKE_TOOLCHAIN_FILE)
    set(CMAKE_TOOLCHAIN_FILE
        "$ENV{VCPKG_ROOT}/scripts/buildsystems/vcpkg.cmake"
        CACHE STRING "VCPKG toolchain file"
    )
    message(STATUS "VCPKG: usando toolchain em ${CMAKE_TOOLCHAIN_FILE}")
endif()
{{end}}
# =============================================================================
# Dependências — raylib
# =============================================================================
# A forma de obter o raylib (VCPKG, FetchContent, ou pacote do sistema) é
# resolvida em cmake/Dependencies.cmake, conforme a escolha feita ao gerar
# o projeto com o cpp-gen.

include(cmake/Dependencies.cmake)

# =============================================================================
# Executável
# =============================================================================

add_executable({{.Name}} src/main.cpp)
target_link_libraries({{.Name}} PRIVATE raylib)

# Frameworks exigidos pelo raylib ao linkar estaticamente no macOS.
if(APPLE)
    target_link_libraries({{.Name}} PRIVATE
        "-framework IOKit"
        "-framework Cocoa"
        "-framework OpenGL"
    )
endif()

# =============================================================================
# Informações de configuração (exibidas ao final do cmake)
# =============================================================================

message(STATUS "")
message(STATUS "=== {{.Name}} v${PROJECT_VERSION} ===")
message(STATUS "  Padrão C++ : C++${CMAKE_CXX_STANDARD}")
message(STATUS "  Dependências: {{if .UseVCPKG}}VCPKG{{else if .UseFetchContent}}FetchContent{{else}}pacote do sistema{{end}}")
message(STATUS "")
