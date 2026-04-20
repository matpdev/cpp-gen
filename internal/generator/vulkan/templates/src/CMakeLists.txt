# =============================================================================
# src/CMakeLists.txt — Executável principal da aplicação
# =============================================================================
# Adicione aqui apenas os arquivos da sua aplicação.
# A biblioteca vklib (abstração Vulkan) está em vklib/ e é linkada abaixo.
# =============================================================================

add_executable(${PROJECT_NAME}
    main.cpp
    # Adicione outros arquivos .cpp da sua aplicação aqui:
    # renderer.cpp
    # scene.cpp
)

target_compile_definitions(${PROJECT_NAME} PRIVATE
    SHADER_DIR="${CMAKE_RUNTIME_OUTPUT_DIRECTORY}/shaders/"
)

# Compila shaders GLSL → SPIR-V automaticamente no build
target_compile_shaders(${PROJECT_NAME}
    OUTPUT_DIR "${CMAKE_RUNTIME_OUTPUT_DIRECTORY}/shaders"
    SOURCES
        ${CMAKE_SOURCE_DIR}/shaders/triangle.vert
        ${CMAKE_SOURCE_DIR}/shaders/triangle.frag
)

target_include_directories(${PROJECT_NAME}
    PRIVATE
        ${PROJECT_SOURCE_DIR}/vklib/include
        ${CMAKE_CURRENT_SOURCE_DIR}
        ${Stb_INCLUDE_DIR}
)

target_link_libraries(${PROJECT_NAME}
    PRIVATE
        vklib
        project_warnings
        # ── Assets ──────────────────────────────────────────────────────────
        tinyobjloader::tinyobjloader
        # ── UI de Debug ─────────────────────────────────────────────────────
        imgui::imgui
)

set_target_properties(${PROJECT_NAME} PROPERTIES
    CXX_STANDARD          23
    CXX_STANDARD_REQUIRED ON
    CXX_EXTENSIONS        OFF
)
