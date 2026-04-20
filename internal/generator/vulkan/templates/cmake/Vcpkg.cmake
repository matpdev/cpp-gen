# =============================================================================
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
