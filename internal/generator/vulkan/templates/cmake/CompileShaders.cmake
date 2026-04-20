# =============================================================================
# cmake/CompileShaders.cmake
# =============================================================================
# Provê target_compile_shaders() para compilar GLSL → SPIR-V usando glslc.
#
# Uso em CMakeLists.txt:
#   include(CompileShaders)
#
#   target_compile_shaders(${PROJECT_NAME}
#       OUTPUT_DIR "${CMAKE_RUNTIME_OUTPUT_DIRECTORY}/shaders"
#       SOURCES
#           shaders/triangle.vert
#           shaders/triangle.frag
#   )
# =============================================================================

# ── Localizar glslc (vem com o Vulkan SDK) ───────────────────────────────────
find_program(GLSLC_EXECUTABLE
    NAMES glslc
    HINTS
        "$ENV{VULKAN_SDK}/bin"
        "$ENV{VK_SDK_PATH}/bin"
    PATHS
        /usr/bin
        /usr/local/bin
)

if(GLSLC_EXECUTABLE)
  message(STATUS "glslc encontrado: ${GLSLC_EXECUTABLE}")
else()
  message(WARNING
        "[CompileShaders] glslc não encontrado!\n"
        "  Instale o Vulkan SDK: https://vulkan.lunarg.com/sdk/home\n"
        "  ou defina a variável de ambiente VULKAN_SDK=/caminho/para/sdk\n"
        "  Os shaders NÃO serão compilados automaticamente."
    )
endif()

# =============================================================================
# target_compile_shaders(<TARGET>
#     [OUTPUT_DIR <dir>]    — padrão: ${CMAKE_RUNTIME_OUTPUT_DIRECTORY}/shaders
#     SOURCES <arq1> ...    — arquivos .vert, .frag, .comp, .geom, etc.
# )
#
# Para cada shader:
#   glslc <fonte> -o <OUTPUT_DIR>/<nome>.spv
#
# Os .spv ficam junto ao executável em bin/shaders/, prontos para LoadModule().
# =============================================================================
function(target_compile_shaders TARGET)
  cmake_parse_arguments(PARSE_ARGV 1
        ARG
        ""              # flags booleanas (nenhuma)
        "OUTPUT_DIR"    # valor único
        "SOURCES"       # lista
    )

  if(NOT GLSLC_EXECUTABLE)
    message(WARNING "[CompileShaders] glslc indisponível — shaders de '${TARGET}' ignorados")
    return()
  endif()

  if(NOT ARG_SOURCES)
    message(WARNING "[CompileShaders] nenhum SOURCES fornecido para '${TARGET}'")
    return()
  endif()

  # Diretório de saída padrão: junto ao executável
  if(NOT ARG_OUTPUT_DIR)
    set(ARG_OUTPUT_DIR "${CMAKE_RUNTIME_OUTPUT_DIRECTORY}/shaders")
  endif()

  file(MAKE_DIRECTORY "${ARG_OUTPUT_DIR}")

  set(SPIRV_OUTPUTS "")

  foreach(SHADER_SOURCE ${ARG_SOURCES})
    # Resolver caminho absoluto
    if(NOT IS_ABSOLUTE "${SHADER_SOURCE}")
      set(SHADER_SOURCE "${CMAKE_CURRENT_SOURCE_DIR}/${SHADER_SOURCE}")
    endif()

    get_filename_component(SHADER_NAME "${SHADER_SOURCE}" NAME)
    set(SPIRV_OUTPUT "${ARG_OUTPUT_DIR}/${SHADER_NAME}.spv")

    add_custom_command(
            OUTPUT  "${SPIRV_OUTPUT}"
            COMMAND "${GLSLC_EXECUTABLE}" "${SHADER_SOURCE}" -o "${SPIRV_OUTPUT}"
            DEPENDS "${SHADER_SOURCE}"
            COMMENT "GLSL → SPIR-V: ${SHADER_NAME}"
            VERBATIM
        )

    list(APPEND SPIRV_OUTPUTS "${SPIRV_OUTPUT}")
    message(STATUS "  [shader] ${SHADER_NAME} → ${SPIRV_OUTPUT}")
  endforeach()

  if(SPIRV_OUTPUTS)
    set(SHADER_TARGET "${TARGET}_shaders")
    add_custom_target("${SHADER_TARGET}" ALL DEPENDS ${SPIRV_OUTPUTS})
    add_dependencies("${TARGET}" "${SHADER_TARGET}")
    message(STATUS "[CompileShaders] '${SHADER_TARGET}' configurado (${TARGET})")
  endif()
endfunction()
