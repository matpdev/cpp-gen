# =============================================================================
# tests/CMakeLists.txt — Testes de {{.Name}}
# =============================================================================
# Layout ativo: two-root
#
# Os testes são executados com:
#   cmake --preset debug && ctest --preset test-debug --output-on-failure
#   cd build && ctest -V                  (verbose)
#   cd build && ctest -R <nome>           (filtrar por nome)
# =============================================================================

# ── Target de teste principal ──────────────────────────────────────────────────────

add_executable({{.NameSnake}}_tests
    test_main.cpp
    # Adicione arquivos de teste adicionais aqui:
    # test_feature_a.cpp
    # test_feature_b.cpp
)

target_include_directories({{.NameSnake}}_tests
    PRIVATE
        ${PROJECT_SOURCE_DIR}/include
        ${CMAKE_CURRENT_SOURCE_DIR}
)

target_link_libraries({{.NameSnake}}_tests
    PRIVATE
        project_warnings
        vklib
        # Adicione frameworks de teste aqui se necessário:
        # Catch2::Catch2WithMain
        # GTest::gtest_main
        # doctest::doctest
)

set_target_properties({{.NameSnake}}_tests PROPERTIES
    CXX_STANDARD          23
    CXX_STANDARD_REQUIRED ON
    CXX_EXTENSIONS        OFF
)

# ── Registro no CTest ─────────────────────────────────────────────────────────────

add_test(
    NAME    {{.Name}}_unit_tests
    COMMAND {{.NameSnake}}_tests
)

set_tests_properties({{.Name}}_unit_tests PROPERTIES
    TIMEOUT 30
    LABELS  "unit"
)
