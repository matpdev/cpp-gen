# =============================================================================
# Dependências — raylib
# =============================================================================
# Resolvida de acordo com o gerenciador de pacotes escolhido ao gerar o
# projeto (--pkg vcpkg | fetchcontent | none).
{{if .UseVCPKG}}
# VCPKG: a dependência já está declarada em vcpkg.json.
# find_package() aqui apenas expõe o target "raylib" ao CMake.
find_package(raylib CONFIG REQUIRED)
{{else if .UseFetchContent}}
# FetchContent: baixa e compila o raylib junto com o projeto, sem exigir
# nenhuma ferramenta externa além do CMake.
include(FetchContent)

set(FETCHCONTENT_QUIET OFF)
set(BUILD_EXAMPLES OFF CACHE BOOL "" FORCE)
set(BUILD_GAMES OFF CACHE BOOL "" FORCE)

FetchContent_Declare(
    raylib
    GIT_REPOSITORY https://github.com/raysan5/raylib.git
    GIT_TAG        5.5
)
FetchContent_MakeAvailable(raylib)
{{else}}
# Nenhum gerenciador de pacotes selecionado: assume raylib instalado no
# sistema (ex: `brew install raylib` ou `sudo apt install libraylib-dev`).
find_package(raylib CONFIG REQUIRED)
{{end}}
