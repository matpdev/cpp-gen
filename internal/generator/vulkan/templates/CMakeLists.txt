cmake_minimum_required(VERSION 3.20)

# =============================================================================
# Declaração do projeto
# =============================================================================
# A versão aqui definida é automaticamente acessível via
# {{.NameUpper}}_VERSION_MAJOR, {{.NameUpper}}_VERSION_MINOR, etc.
project(
    {{.Name}}
    VERSION {{.Version}}
    DESCRIPTION "{{if .Description}}{{.Description}}{{else}}Projeto Vulkan C++ gerado pelo cpp-gen{{end}}"
    LANGUAGES CXX
)

# Impede builds in-source, que misturam artefatos com o código fonte.
if(CMAKE_SOURCE_DIR STREQUAL CMAKE_BINARY_DIR)
  message(FATAL_ERROR
        "Build in-source detectado!
"
        "Por favor, crie um diretório separado para o build:
"
        "  cmake -B build
"
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
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

# Organiza os binários em subdiretórios padrão dentro do build.
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/bin")
set(CMAKE_LIBRARY_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/lib")
set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/lib")

# =============================================================================
# Módulos CMake auxiliares
# =============================================================================

list(APPEND CMAKE_MODULE_PATH "${CMAKE_CURRENT_SOURCE_DIR}/cmake")
{{if .UseVCPKG}}
# =============================================================================
# Integração VCPKG
# =============================================================================
# O VCPKG é integrado via toolchain file, definido ANTES do primeiro project().
# A forma recomendada é usar CMakePresets.json (já configurado) ou:
#
#   cmake -B build --preset vcpkg-debug
#
# Variável de ambiente necessária: VCPKG_ROOT=/caminho/para/vcpkg
if(DEFINED ENV{VCPKG_ROOT} AND NOT DEFINED CMAKE_TOOLCHAIN_FILE)
  set(CMAKE_TOOLCHAIN_FILE
        "$ENV{VCPKG_ROOT}/scripts/buildsystems/vcpkg.cmake"
        CACHE STRING "VCPKG toolchain file"
    )
  message(STATUS "VCPKG: usando toolchain em ${CMAKE_TOOLCHAIN_FILE}")
endif()
{{end}}
# =============================================================================
# Módulos internos
# =============================================================================

include(CompilerWarnings)
include(CompileShaders)

# =============================================================================
# Dependências externas
# =============================================================================

# Vulkan — busca no dispositivo do sistema ou via Vulkan SDK
find_package(Vulkan REQUIRED)

# vk-bootstrap — sempre via FetchContent (não disponível no vcpkg com esta config)
include(FetchContent)

set(VK_BOOTSTRAP_TEST OFF CACHE BOOL "" FORCE)

FetchContent_Declare(
    fetch_vk_bootstrap
    GIT_REPOSITORY https://github.com/charles-lunarg/vk-bootstrap
    GIT_TAG        v1.4.349
)
FetchContent_MakeAvailable(fetch_vk_bootstrap)

# Demais dependências (via VCPKG ou pacotes do sistema)
find_package(VulkanMemoryAllocator CONFIG REQUIRED)
find_package(glfw3 CONFIG REQUIRED)
find_package(glm CONFIG REQUIRED)
find_package(Stb REQUIRED)
find_package(tinyobjloader CONFIG REQUIRED)
find_package(imgui CONFIG REQUIRED)

# =============================================================================
# Subdiretórios
# =============================================================================
# Estrutura do projeto:
#   vklib/  — biblioteca de abstração Vulkan (engine)
#   src/    — código da sua aplicação

add_subdirectory(vklib)
add_subdirectory(src)

# =============================================================================
# Testes
# =============================================================================

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
message(STATUS "  Compilador        : ${CMAKE_CXX_COMPILER_ID} ${CMAKE_CXX_COMPILER_VERSION}")
message(STATUS "  Padrão C++        : C++${CMAKE_CXX_STANDARD}")
message(STATUS "  Tipo de build     : ${CMAKE_BUILD_TYPE}")
message(STATUS "  Diretório de build: ${CMAKE_BINARY_DIR}")
message(STATUS "  VCPKG             : {{if .UseVCPKG}}habilitado{{else}}desabilitado (pacotes do sistema){{end}}")
message(STATUS "  Testes            : {{printf "${%s_BUILD_TESTS}" .NameUpper}}")
message(STATUS "")
