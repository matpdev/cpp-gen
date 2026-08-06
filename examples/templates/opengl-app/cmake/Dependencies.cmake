# =============================================================================
# Dependências — GLFW + GLEW
# =============================================================================
# Resolvida de acordo com o gerenciador de pacotes escolhido ao gerar o
# projeto (--pkg vcpkg | fetchcontent | none).
{{if .UseFetchContent}}
# FetchContent: baixa e compila GLFW e GLEW junto com o projeto, sem exigir
# nenhuma ferramenta externa além do CMake.
include(FetchContent)
set(FETCHCONTENT_QUIET OFF)

# ── GLFW ──────────────────────────────────────────────────────────────────
set(GLFW_BUILD_EXAMPLES OFF CACHE BOOL "" FORCE)
set(GLFW_BUILD_TESTS OFF CACHE BOOL "" FORCE)
set(GLFW_BUILD_DOCS OFF CACHE BOOL "" FORCE)
set(GLFW_INSTALL OFF CACHE BOOL "" FORCE)

FetchContent_Declare(
    glfw
    GIT_REPOSITORY https://github.com/glfw/glfw.git
    GIT_TAG        3.4
)
FetchContent_MakeAvailable(glfw)

# ── GLEW ──────────────────────────────────────────────────────────────────
# glew-cmake (fork da Perlmint) empacota o GLEW com um CMakeLists.txt
# utilizável via FetchContent — o projeto oficial do GLEW não tem.
set(glew-cmake_BUILD_SHARED OFF CACHE BOOL "" FORCE)
set(glew-cmake_BUILD_STATIC ON CACHE BOOL "" FORCE)
set(ONLY_LIBS ON CACHE BOOL "" FORCE)

FetchContent_Declare(
    glew
    GIT_REPOSITORY https://github.com/Perlmint/glew-cmake.git
    GIT_TAG        glew-cmake-2.2.0
)

# O CMakeLists.txt do glew-cmake declara um cmake_minimum_required antigo,
# incompatível com CMake 4+ (que removeu suporte a < 3.5). Relaxa a policy
# mínima só para este subprojeto.
set(CMAKE_POLICY_VERSION_MINIMUM 3.5)

# glew-cmake linka incondicionalmente contra o framework AGL no macOS, que
# não existe mais em SDKs recentes (find_library ainda "encontra" a pasta
# do framework, mas o link falha: "framework 'AGL' not found"). Pré-cacheia
# a variável como vazia para transformar aquele find_library em um no-op.
if(APPLE)
    set(AGL_LIBRARY "" CACHE FILEPATH "AGL removido dos SDKs modernos do macOS" FORCE)
endif()

FetchContent_MakeAvailable(glew)
unset(CMAKE_POLICY_VERSION_MINIMUM)
{{else}}
# VCPKG ou pacote do sistema: ambos expõem GLFW/GLEW via find_package()
# (VCPKG em modo config; pacotes do sistema via os módulos Find* do CMake).
find_package(glfw3 CONFIG REQUIRED)
find_package(GLEW REQUIRED)
{{end}}
