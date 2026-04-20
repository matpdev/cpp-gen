# =============================================================================
# cmake/CompilerWarnings.cmake
# =============================================================================
# Define o target de interface `project_warnings` com um conjunto abrangente
# de flags de warning para GCC, Clang e MSVC.
#
# Uso:
#   target_link_libraries(<seu_target> PRIVATE project_warnings)
#
# Referências:
#   - https://gcc.gnu.org/onlinedocs/gcc/Warning-Options.html
#   - https://clang.llvm.org/docs/DiagnosticsReference.html
#   - https://docs.microsoft.com/cpp/build/reference/compiler-options
# =============================================================================

add_library(project_warnings INTERFACE)

# ── Flags para GCC e Clang ────────────────────────────────────────────────────
if(CMAKE_CXX_COMPILER_ID MATCHES "GNU|Clang|AppleClang")
    target_compile_options(project_warnings INTERFACE
        -Wall                   # Warnings padrão essenciais
        -Wextra                 # Warnings adicionais além de -Wall
        -Wpedantic              # Exige conformidade estrita com o padrão ISO
        -Wshadow                # Variável local oculta outra de escopo externo
        -Wnon-virtual-dtor      # Destrutor não-virtual em classe com métodos virtuais
        -Wold-style-cast        # Cast estilo C (use static_cast/reinterpret_cast)
        -Wcast-align            # Cast que aumenta o alinhamento do ponteiro
        -Wunused                # Variáveis, funções e parâmetros não utilizados
        -Woverloaded-virtual    # Função virtual oculta sobrecarga da base
        -Wconversion            # Conversão implícita que pode perder dados
        -Wsign-conversion       # Conversão entre tipos com/sem sinal
        -Wmisleading-indentation # Indentação enganosa (sem chaves)
        -Wduplicated-cond       # Condição duplicada em if/else-if (GCC)
        -Wduplicated-branches   # Ramos idênticos em if/else (GCC)
        -Wlogical-op            # Operadores lógicos suspeitos (GCC)
        -Wnull-dereference      # Possível derreferência de ponteiro nulo
        -Wdouble-promotion      # Float promovido implicitamente para double
        -Wformat=2              # Verificações rigorosas de printf/scanf
        -Wimplicit-fallthrough  # Case sem break em switch (requer [[fallthrough]])
    )

    # Clang tem warnings adicionais úteis
    if(CMAKE_CXX_COMPILER_ID MATCHES "Clang|AppleClang")
        target_compile_options(project_warnings INTERFACE
            -Wno-unknown-warning-option # Ignora warnings não reconhecidos por versões antigas
        )
    endif()

    # GCC >= 8 suporta Wduplicated-cond e Wduplicated-branches
    # Versões mais antigas ignorarão silenciosamente via Wno-unknown-warning-option
endif()

# ── Flags para MSVC ───────────────────────────────────────────────────────────
if(MSVC)
    target_compile_options(project_warnings INTERFACE
        /W4         # Nível 4 de warnings (mais alto sem ser /Wall)
        /w14242     # Conversão: possível perda de dados
        /w14254     # Operador: possível perda de dados em conversão
        /w14263     # Função membro não sobrescreve virtual da classe base
        /w14265     # Classe com funções virtuais mas sem destrutor virtual
        /w14287     # Constante negativa usada como operando sem sinal
        /we4289     # Variável de loop usada fora do for
        /w14296     # Expressão sempre verdadeira ou falsa
        /w14311     # Truncamento de ponteiro: de 64-bit para 32-bit
        /w14545     # Expressão antes da vírgula sem efeito colateral
        /w14546     # Chamada de função antes da vírgula sem efeito colateral
        /w14547     # Operador sem efeito colateral antes da vírgula
        /w14549     # Operador sem efeito colateral antes da vírgula
        /w14555     # Expressão sem efeito colateral
        /w14619     # #pragma warning: número de warning inexistente
        /w14640     # Objeto local estático não é thread-safe
        /w14826     # Conversão implícita com sinal/sem sinal
        /w14905     # String literal convertida para LPSTR
        /w14906     # String literal convertida para LPWSTR
        /w14928     # Inicialização de cópia ilegal
        /permissive- # Desabilita extensões não-padrão do MSVC
    )
endif()

# ── Modo de tratamento de warnings ────────────────────────────────────────────
# Descomente para tratar todos os warnings como erros (-Werror / /WX).
# Recomendado em CI/CD, mas pode ser inconveniente durante desenvolvimento.
#
# if(CMAKE_CXX_COMPILER_ID MATCHES "GNU|Clang|AppleClang")
#     target_compile_options(project_warnings INTERFACE -Werror)
# elseif(MSVC)
#     target_compile_options(project_warnings INTERFACE /WX)
# endif()
