/**
 * @file test_main.cpp
 * @brief Testes de vulkan-start-template.
 *
 * Testes registrados com CTest via add_test() no CMakeLists.txt de tests/.
 *
 * Para adicionar um framework externo (Catch2, GoogleTest, doctest),
 * configure a dependência no gerenciador de pacotes e adapte este arquivo.
 *
 * Execução:
 * @code
 *   cd build && ctest --output-on-failure
 * @endcode
 *
 * @author Matheus
 * @version 1.0.0
 * @date 2026
 *
 * Layout: two-root
 */

#include <cassert>
#include <iostream>
#include <stdexcept>
#include <string>

// ─────────────────────────────────────────────────────────────────────────────
// Macro auxiliar para testes
// ─────────────────────────────────────────────────────────────────────────────

/**
 * @brief Verifica uma condição e encerra com mensagem de erro se falsa.
 * Diferente de assert(), não é removido em builds Release.
 */
#define CHECK(condition)                                                     \
    do {                                                                     \
        if (!(condition)) {                                                  \
            std::cerr << "[FALHOU] " << __FILE__ << ":" << __LINE__          \
                      << " — condição falsa: " #condition "\n";             \
            return 1;                                                        \
        }                                                                    \
        std::cout << "[OK]     " #condition "\n";                            \
    } while (false)

// ─────────────────────────────────────────────────────────────────────────────
// Testes
// ─────────────────────────────────────────────────────────────────────────────

// ─────────────────────────────────────────────────────────────────────────────

int main() {
    std::cout << "=== Testes de vulkan-start-template (layout: two-root) ===\n\n";

    int failures = 0;
    // Adicione seus testes aqui.
    CHECK(1 + 1 == 2);

    std::cout << "\n";
    if (failures == 0) {
        std::cout << "Todos os testes passaram.\n";
        return 0;
    }
    std::cout << failures << " teste(s) falharam.\n";
    return 1;
}
