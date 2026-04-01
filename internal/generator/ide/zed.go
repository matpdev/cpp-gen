// Package ide contém os geradores de configuração de IDE para projetos C++.
package ide

import (
	"fmt"
	"path/filepath"
)

// ─────────────────────────────────────────────────────────────────────────────
// generateZed — ponto de entrada
// ─────────────────────────────────────────────────────────────────────────────

// generateZed cria o diretório .zed/ com as configurações completas para o
// editor Zed (https://zed.dev), um editor moderno e de alta performance
// escrito em Rust.
//
// Arquivos gerados:
//
//   - .zed/settings.json  — configurações do workspace: clangd LSP,
//     clang-format, preferências de editor, inlay hints e DAP (debug)
//   - .zed/tasks.json     — tarefas de cmake configure, build, clean,
//     test e execução do binário, acessíveis via task runner do Zed
//
// Compatibilidade: Zed 0.140+ (suporte estável a LSP C++ e tasks).
// Referência: https://zed.dev/docs/configuring-zed
func generateZed(root string, data *Data, verbose bool) error {
	zedDir := filepath.Join(root, ".zed")

	files := []struct {
		name     string
		relPath  string
		tmplName string
		tmpl     string
	}{
		{
			name:     "settings.json",
			relPath:  filepath.Join(zedDir, "settings.json"),
			tmplName: "zed_settings",
			tmpl:     tmplZedSettings,
		},
		{
			name:     "tasks.json",
			relPath:  filepath.Join(zedDir, "tasks.json"),
			tmplName: "zed_tasks",
			tmpl:     tmplZedTasks,
		},
	}

	for _, f := range files {
		if err := writeIDETemplate(f.relPath, f.tmplName, f.tmpl, data, verbose); err != nil {
			return fmt.Errorf("gerar .zed/%s: %w", f.name, err)
		}
	}

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Template: .zed/settings.json
// ─────────────────────────────────────────────────────────────────────────────

// tmplZedSettings é o template para .zed/settings.json.
//
// O arquivo settings.json do Zed define configurações de workspace que
// sobrepõem as configurações globais do usuário (~/.config/zed/settings.json)
// apenas para este projeto. Isso garante que todos os colaboradores usem
// as mesmas configurações de formatação, LSP e editor.
//
// Seções configuradas:
//
//	Linguagens C e C++:
//	  - Formatação automática ao salvar via clang-format
//	  - Tab size de 4 espaços, sem tabs
//	  - Inlay hints (tipos de variáveis, nomes de parâmetros)
//
//	LSP — Clangd:
//	  - compile-commands-dir apontando para o preset de debug
//	  - clang-tidy integrado
//	  - header-insertion via IWYU
//	  - completion all-scopes
//
//	Editor geral:
//	  - Exclusão de diretórios de build das buscas e árvore de arquivos
//	  - Fim de linha Unix (LF)
//	  - Trim de espaços em branco ao salvar
//
//	DAP — Debug Adapter Protocol:
//	  - Configurações de debug para LLDB (Linux/macOS) e CodeLLDB
//	  - Apontam para o binário gerado pelo preset de debug
//
// Referência: https://zed.dev/docs/configuring-zed
//
//	https://zed.dev/docs/languages/c
const tmplZedSettings = `{
  // ============================================================================
  // .zed/settings.json — Configurações de workspace para {{ .ProjectName }}
  // ============================================================================
  // Estas configurações aplicam-se APENAS a este workspace e sobrepõem
  // as configurações globais do usuário em ~/.config/zed/settings.json.
  //
  // Documentação: https://zed.dev/docs/configuring-zed
  // ============================================================================

  // ── Configurações de linguagem ────────────────────────────────────────────

  "languages": {

    // ── C++ ───────────────────────────────────────────────────────────────────
    "C++": {
      // Tamanho de indentação (consistente com .clang-format)
      "tab_size": 4,

      // Usa espaços em vez de tabs
      "hard_tabs": false,

      // Comprimento máximo de linha para guia visual (consistente com .clang-format)
      "preferred_line_length": 100,

      // Formata o arquivo ao salvar usando o clang-format instalado no sistema.
      // O Zed usa o .clang-format da raiz do projeto automaticamente.
      "format_on_save": "on",
      "formatter": {
        "external": {
          "command": "clang-format",
          // Lê o arquivo da stdin e escreve na stdout (-style=file usa .clang-format)
          "arguments": ["--style=file", "-"]
        }
      },

      // Inlay hints: exibe tipos de variáveis 'auto' e nomes de parâmetros inline.
      // Requer suporte do servidor Clangd (habilitado na seção "lsp" abaixo).
      "inlay_hints": {
        "enabled": true,
        "show_parameter_hints": true,
        "show_type_hints": true,
        "show_other_hints": true
      },

      // Fim de linha Unix (LF) — evita problemas de diff em ambientes mistos
      "line_ending": "unix"
    },

    // ── C (compartilha as mesmas configurações do C++) ─────────────────────
    "C": {
      "tab_size": 4,
      "hard_tabs": false,
      "preferred_line_length": 100,
      "format_on_save": "on",
      "formatter": {
        "external": {
          "command": "clang-format",
          "arguments": ["--style=file", "-"]
        }
      },
      "inlay_hints": {
        "enabled": true,
        "show_parameter_hints": true,
        "show_type_hints": true,
        "show_other_hints": true
      },
      "line_ending": "unix"
    }
  },

  // ── LSP — Clangd ──────────────────────────────────────────────────────────
  //
  // Configura o servidor de linguagem Clangd para este projeto.
  // O Zed usa o Clangd como LSP padrão para C/C++ e o gerencia automaticamente.
  //
  // O Clangd encontra o compile_commands.json gerado pelo CMake no diretório
  // especificado em "compile-commands-dir". O preset de debug é usado como
  // padrão pois é o ambiente de desenvolvimento mais comum.
  //
  // Referência: https://clangd.llvm.org/config.html

  "lsp": {
    "clangd": {
      "binary": {
        // Caminho para o binário clangd. "clangd" usa o do PATH.
        // Descomente e ajuste se precisar de uma versão específica:
        // "path": "/usr/bin/clangd-18",
        "path": "clangd",

        // Argumentos passados ao servidor Clangd ao iniciar.
        "arguments": [
          // Aponta para o compile_commands.json do preset de debug.
          // Gerado automaticamente pelo CMake (CMAKE_EXPORT_COMPILE_COMMANDS=ON).
          "--compile-commands-dir=build/{{if .UseVCPKG}}vcpkg-debug{{else}}debug{{end}}",

          // Número de threads para indexação em background.
          // Ajuste conforme o número de cores disponíveis.
          "--j=4",

          // Inclui símbolos de todos os escopos no autocompletar,
          // não apenas do escopo atual.
          "--all-scopes-completion=true",

          // Sugere incluir headers automaticamente ao completar
          // símbolos desconhecidos (usa IWYU — Include What You Use).
          "--header-insertion=iwyu",

          // Habilita diagnósticos do clang-tidy inline no editor.
          // As regras são configuradas no .clangd da raiz do projeto.
          "--clang-tidy",

          // Habilita inlay hints no protocolo LSP.
          "--inlay-hints",

          // Habilita leitura do arquivo .clangd do projeto.
          "--enable-config",

          // Nível de log do servidor (error | warning | info | verbose).
          "--log=error",

          // Habilita cache em disco para acelerar reinicializações.
          "--pch-storage=memory"
        ]
      },

      // Configurações de inicialização do Clangd (initializationOptions).
      "initialization_options": {
        // Habilita o índice em background de todo o projeto.
        "index": {
          "enableBackgroundIndexing": true
        },
        // Exibe a documentação em hover no formato Markdown.
        "hover": {
          "showAKA": true
        }
      }
    }
  },

  // ── Editor ────────────────────────────────────────────────────────────────

  // Remove espaços em branco no final das linhas ao salvar.
  "remove_trailing_whitespace_on_save": true,

  // Insere uma linha em branco no final de cada arquivo ao salvar.
  "ensure_final_newline_on_save": true,

  // Exibe a régua vertical na coluna 100 (consistente com .clang-format).
  "rulers": [100],

  // Exibe números de linha relativos (útil para navegação com hjkl/Vim mode).
  // Altere para "absolute" se preferir números absolutos.
  "relative_line_numbers": false,

  // Comportamento do git blame inline.
  "git": {
    "inline_blame": {
      "enabled": true,
      // Exibe o blame apenas após o cursor ficar parado por 600ms.
      "delay_ms": 600
    }
  },

  // ── Arquivo — Exclusões ───────────────────────────────────────────────────
  //
  // Oculta diretórios gerados da árvore de arquivos do Zed e das buscas.
  // NOTA: isto não afeta o .gitignore — apenas a visualização no editor.

  "file_scan_exclusions": [
    "**/.git",
    "**/build",
    "**/build-*",
    "**/install",
    "**/dist",
    "**/vcpkg_installed",
    "**/.cache",
    "**/*.o",
    "**/*.a",
    "**/*.so",
    "**/*.dylib",
    "**/*.d"
  ],

  // ── Terminal integrado ────────────────────────────────────────────────────

  "terminal": {
    // Diretório de trabalho padrão do terminal integrado.
    // "current_project_directory" abre sempre na raiz do workspace.
    "working_directory": "current_project_directory",

    // Shell padrão (null = usa o shell do sistema definido em $SHELL).
    "shell": { "with_arguments": { "program": "bash", "args": [] } }
  },

  // ── DAP — Debug Adapter Protocol ─────────────────────────────────────────
  //
  // Configurações de debug integrado do Zed via DAP.
  // Requer o adaptador de debug instalado:
  //   - LLDB: instalado junto com o Xcode CLT (macOS) ou lldb (Linux)
  //   - CodeLLDB: extensão autônoma (https://github.com/vadimcn/codelldb)
  //
  // Para iniciar o debug: Ctrl+Shift+D (ou pelo menu Run > Debug)
  //
  // Referência: https://zed.dev/docs/debugger

  "debugger": {
    // Adapters disponíveis para C++.
    // O Zed detecta automaticamente o adapter disponível no sistema.
    "adapters": {
      // LLDB nativo (macOS/Linux com clang)
      "lldb": {
        "type": "lldb"
      }
    }
  }
}
`

// ─────────────────────────────────────────────────────────────────────────────
// Template: .zed/tasks.json
// ─────────────────────────────────────────────────────────────────────────────

// tmplZedTasks é o template para .zed/tasks.json.
//
// Define as tarefas de build acessíveis no Zed via:
//   - Ctrl+Shift+B        — executa a tarefa padrão de build
//   - Ctrl+Shift+P → "task: spawn"  — lista todas as tarefas
//   - Menu: Run > Task    — lista e executa tarefas
//
// Variáveis disponíveis nos campos "command" e "args":
//
//	$ZED_WORKTREE_ROOT  — caminho absoluto da raiz do projeto
//	$ZED_FILE           — caminho absoluto do arquivo aberto
//	$ZED_FILENAME       — nome do arquivo aberto (sem diretório)
//	$ZED_COLUMN         — coluna atual do cursor
//	$ZED_ROW            — linha atual do cursor
//
// Referência: https://zed.dev/docs/tasks
const tmplZedTasks = `[
  // ============================================================================
  // .zed/tasks.json — Tarefas de build para {{ .ProjectName }}
  // ============================================================================
  // Acesso via:
  //   Ctrl+Shift+P → "task: spawn"    lista e executa qualquer tarefa
  //   Ctrl+Shift+B                    re-executa a última tarefa usada
  //
  // Variáveis disponíveis:
  //   $ZED_WORKTREE_ROOT  raiz do projeto (onde este arquivo está)
  //   $ZED_FILE           arquivo aberto no momento
  //
  // Documentação: https://zed.dev/docs/tasks
  // ============================================================================

  // ── Configure ──────────────────────────────────────────────────────────────

  {
    "label": "CMake: Configure (Debug)",
    "command": "cmake",
    "args": [
      "--preset",
      "{{if .UseVCPKG}}vcpkg-debug{{else}}debug{{end}}"
    ],
    "cwd": "$ZED_WORKTREE_ROOT",
    // Sempre exibe o painel de saída para acompanhar o progresso.
    "reveal": "always",
    // Não permite execuções simultâneas do mesmo configure.
    "allow_concurrent_runs": false,
    "use_new_terminal": false
  },

  {
    "label": "CMake: Configure (Release)",
    "command": "cmake",
    "args": [
      "--preset",
      "{{if .UseVCPKG}}vcpkg-release{{else}}release{{end}}"
    ],
    "cwd": "$ZED_WORKTREE_ROOT",
    "reveal": "always",
    "allow_concurrent_runs": false,
    "use_new_terminal": false
  },

  {
    "label": "CMake: Configure (Sanitizers)",
    "command": "cmake",
    "args": ["--preset", "sanitize"],
    "cwd": "$ZED_WORKTREE_ROOT",
    "reveal": "always",
    "allow_concurrent_runs": false,
    "use_new_terminal": false
  },

  // ── Build ──────────────────────────────────────────────────────────────────

  {
    "label": "CMake: Build (Debug)",
    "command": "cmake",
    "args": [
      "--build",
      "--preset",
      "{{if .UseVCPKG}}build-vcpkg-debug{{else}}build-debug{{end}}"
    ],
    "cwd": "$ZED_WORKTREE_ROOT",
    "reveal": "always",
    "allow_concurrent_runs": false,
    "use_new_terminal": false
  },

  {
    "label": "CMake: Build (Release)",
    "command": "cmake",
    "args": [
      "--build",
      "--preset",
      "{{if .UseVCPKG}}build-vcpkg-release{{else}}build-release{{end}}"
    ],
    "cwd": "$ZED_WORKTREE_ROOT",
    "reveal": "always",
    "allow_concurrent_runs": false,
    "use_new_terminal": false
  },

  {
    "label": "CMake: Build (Sanitizers)",
    "command": "cmake",
    "args": ["--build", "--preset", "build-sanitize"],
    "cwd": "$ZED_WORKTREE_ROOT",
    "reveal": "always",
    "allow_concurrent_runs": false,
    "use_new_terminal": false
  },

  // ── Clean ──────────────────────────────────────────────────────────────────

  {
    "label": "CMake: Clean (Debug)",
    "command": "cmake",
    "args": [
      "--build",
      "--preset",
      "{{if .UseVCPKG}}build-vcpkg-debug{{else}}build-debug{{end}}",
      "--target", "clean"
    ],
    "cwd": "$ZED_WORKTREE_ROOT",
    "reveal": "always",
    "allow_concurrent_runs": false,
    "use_new_terminal": false
  },

  {
    "label": "CMake: Delete build/",
    "command": "rm",
    "args": ["-rf", "build"],
    "cwd": "$ZED_WORKTREE_ROOT",
    // "no_focus" exibe a saída mas não muda o foco para o terminal.
    "reveal": "no_focus",
    "allow_concurrent_runs": false,
    "use_new_terminal": false
  },

  // ── Testes ─────────────────────────────────────────────────────────────────

  {
    "label": "CMake: Run Tests (Debug)",
    "command": "ctest",
    "args": [
      "--preset",
      "{{if .UseVCPKG}}test-vcpkg-debug{{else}}test-debug{{end}}",
      "--output-on-failure"
    ],
    "cwd": "$ZED_WORKTREE_ROOT",
    "reveal": "always",
    "allow_concurrent_runs": false,
    "use_new_terminal": false
  },

  {
    "label": "CMake: Run Tests (verbose)",
    "command": "ctest",
    "args": [
      "--preset",
      "{{if .UseVCPKG}}test-vcpkg-debug{{else}}test-debug{{end}}",
      "--output-on-failure",
      "--verbose"
    ],
    "cwd": "$ZED_WORKTREE_ROOT",
    "reveal": "always",
    "allow_concurrent_runs": false,
    "use_new_terminal": false
  },
{{if .IsExecutable}}
  // ── Executar ───────────────────────────────────────────────────────────────

  {
    "label": "Run: {{ .ProjectName }} (Debug)",
    "command": "./build/{{if .UseVCPKG}}vcpkg-debug{{else}}debug{{end}}/bin/{{ .ProjectName }}",
    "args": [],
    "cwd": "$ZED_WORKTREE_ROOT",
    // "always" abre um terminal dedicado para ver a saída do programa.
    "reveal": "always",
    // Permite rodar o binário várias vezes simultaneamente se necessário.
    "allow_concurrent_runs": true,
    // Abre em um terminal separado para não misturar com saída do build.
    "use_new_terminal": true
  },

  {
    "label": "Build & Run: {{ .ProjectName }} (Debug)",
    "command": "bash",
    "args": [
      "-c",
      "cmake --build --preset {{if .UseVCPKG}}build-vcpkg-debug{{else}}build-debug{{end}} && echo '\\n─── Executando {{ .ProjectName }} ───\\n' && ./build/{{if .UseVCPKG}}vcpkg-debug{{else}}debug{{end}}/bin/{{ .ProjectName }}"
    ],
    "cwd": "$ZED_WORKTREE_ROOT",
    "reveal": "always",
    "allow_concurrent_runs": false,
    "use_new_terminal": true
  },
{{end}}
  // ── Ferramentas ─────────────────────────────────────────────────────────────

  {
    "label": "Clang-Format: Format file",
    "command": "clang-format",
    "args": ["-i", "$ZED_FILE"],
    "cwd": "$ZED_WORKTREE_ROOT",
    "reveal": "no_focus",
    "allow_concurrent_runs": false,
    "use_new_terminal": false
  },

  {
    "label": "Clang-Format: Format project",
    "command": "bash",
    "args": [
      "-c",
      "find src include tests -name '*.cpp' -o -name '*.hpp' -o -name '*.h' -o -name '*.cxx' | xargs clang-format -i && echo 'Formatação concluída.'"
    ],
    "cwd": "$ZED_WORKTREE_ROOT",
    "reveal": "always",
    "allow_concurrent_runs": false,
    "use_new_terminal": false
  },

  {
    "label": "CMake: Symlink compile_commands.json",
    "command": "bash",
    "args": [
      "-c",
      "ln -sf build/{{if .UseVCPKG}}vcpkg-debug{{else}}debug{{end}}/compile_commands.json compile_commands.json && echo 'compile_commands.json vinculado.'"
    ],
    "cwd": "$ZED_WORKTREE_ROOT",
    "reveal": "no_focus",
    "allow_concurrent_runs": false,
    "use_new_terminal": false
  }
]
`
