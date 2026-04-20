// Package ide contains the IDE configuration generators for C++ projects.
package ide

import (
	"fmt"
	"path/filepath"
)

// ─────────────────────────────────────────────────────────────────────────────
// generateVSCode — entry point
// ─────────────────────────────────────────────────────────────────────────────

// generateVSCode creates the .vscode/ directory with all configuration files
// necessary for a complete workflow in Visual Studio Code:
//
//   - tasks.json       — configure, build, clean and test tasks via CMake
//   - launch.json      — debug configurations with CodeLLDB and cppdbg (MS)
//   - settings.json    — workspace configurations (clangd, cmake-tools, editor)
//   - extensions.json  — recommended extensions for the project
//   - c_cpp_properties.json — IntelliSense configuration (fallback without clangd)
//
// All files are generated in .vscode/ at the project root and should be
// versioned with the code to ensure a consistent environment for the team.
func generateVSCode(root string, data *Data, verbose bool) error {
	vscodeDir := filepath.Join(root, ".vscode")

	files := []struct {
		name     string // file name (for error messages)
		relPath  string // path relative to root
		tmplName string // Go template name
		tmpl     string // template content
	}{
		{
			name:     "tasks.json",                           // file name (for error messages)
			relPath:  filepath.Join(vscodeDir, "tasks.json"), // path relative to root
			tmplName: "vscode_tasks",                         // Go template name
			tmpl:     tmplVSCodeTasks,                        // template content
		},
		{
			name:     "launch.json",
			relPath:  filepath.Join(vscodeDir, "launch.json"),
			tmplName: "vscode_launch",
			tmpl:     tmplVSCodeLaunch,
		},
		{
			name:     "settings.json",
			relPath:  filepath.Join(vscodeDir, "settings.json"),
			tmplName: "vscode_settings",
			tmpl:     tmplVSCodeSettings,
		},
		{
			name:     "extensions.json",
			relPath:  filepath.Join(vscodeDir, "extensions.json"),
			tmplName: "vscode_extensions",
			tmpl:     tmplVSCodeExtensions,
		},
		{
			name:     "c_cpp_properties.json",
			relPath:  filepath.Join(vscodeDir, "c_cpp_properties.json"),
			tmplName: "vscode_cpp_properties",
			tmpl:     tmplVSCodeCppProperties,
		},
	}

	for _, f := range files {
		if err := writeIDETemplate(f.relPath, f.tmplName, f.tmpl, data, verbose); err != nil {
			return fmt.Errorf("gerar .vscode/%s: %w", f.name, err)
		}
	}

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Template: .vscode/tasks.json
// ─────────────────────────────────────────────────────────────────────────────

// tmplVSCodeTasks is the template for .vscode/tasks.json.
//
// Defines build tasks integrated with VSCode, accessible via:
//   - Terminal > Run Task...         (Ctrl+Shift+P → "Run Task")
//   - Ctrl+Shift+B                   (default build task)
//
// Generated tasks:
//
//	┌─────────────────────────────────────────────────────────────────┐
//	│ CMake: Configure (Debug)   — cmake --preset debug               │
//	│ CMake: Configure (Release) — cmake --preset release             │
//	│ CMake: Build (Debug)       — cmake --build --preset build-debug │
//	│ CMake: Build (Release)     — cmake --build --preset build-release│
//	│ CMake: Clean               — cmake --build --target clean       │
//	│ CMake: Run Tests           — ctest --preset test-debug          │
//	│ CMake: Rebuild             — clean + build in sequence          │
//	└─────────────────────────────────────────────────────────────────┘
//
// Reference: https://code.visualstudio.com/docs/editor/tasks
const tmplVSCodeTasks = `{
    // ==========================================================================
    // .vscode/tasks.json — Tarefas de build para {{ .ProjectName }}
    // ==========================================================================
    // Documentação: https://code.visualstudio.com/docs/editor/tasks
    //
    // Atalhos:
    //   Ctrl+Shift+B         → Build Debug (tarefa padrão)
    //   Ctrl+Shift+P → "Run Task" → lista todas as tarefas
    // ==========================================================================
    "version": "2.0.0",
    "tasks": [

        // ── Configure ──────────────────────────────────────────────────────────

        {
            "label": "CMake: Configure (Debug)",
            "detail": "Configura o projeto CMake com o preset 'debug'",
            "type": "shell",
            "command": "cmake",
            "args": [
                "--preset",
                "{{if .UseVCPKG}}vcpkg-debug{{else}}debug{{end}}"
            ],
            "options": {
                "cwd": "${workspaceFolder}"
            },
            "group": "build",
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared",
                "showReuseMessage": false,
                "clear": false
            },
            "problemMatcher": ["$gcc", "$msCompile"]
        },

        {
            "label": "CMake: Configure (Release)",
            "detail": "Configura o projeto CMake com o preset 'release'",
            "type": "shell",
            "command": "cmake",
            "args": [
                "--preset",
                "{{if .UseVCPKG}}vcpkg-release{{else}}release{{end}}"
            ],
            "options": {
                "cwd": "${workspaceFolder}"
            },
            "group": "build",
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared",
                "showReuseMessage": false,
                "clear": false
            },
            "problemMatcher": ["$gcc", "$msCompile"]
        },

        // ── Build ──────────────────────────────────────────────────────────────

        {
            "label": "CMake: Build (Debug)",
            "detail": "Compila o projeto em modo Debug — atalho padrão: Ctrl+Shift+B",
            "type": "shell",
            "command": "cmake",
            "args": [
                "--build",
                "--preset",
                "{{if .UseVCPKG}}build-vcpkg-debug{{else}}build-debug{{end}}"
            ],
            "options": {
                "cwd": "${workspaceFolder}"
            },
            "group": {
                "kind": "build",
                "isDefault": true
            },
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared",
                "showReuseMessage": false,
                "clear": true
            },
            "problemMatcher": ["$gcc", "$msCompile"],
            "dependsOn": ["CMake: Configure (Debug)"]
        },

        {
            "label": "CMake: Build (Release)",
            "detail": "Compila o projeto em modo Release com otimizações",
            "type": "shell",
            "command": "cmake",
            "args": [
                "--build",
                "--preset",
                "{{if .UseVCPKG}}build-vcpkg-release{{else}}build-release{{end}}"
            ],
            "options": {
                "cwd": "${workspaceFolder}"
            },
            "group": "build",
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared",
                "showReuseMessage": false,
                "clear": true
            },
            "problemMatcher": ["$gcc", "$msCompile"],
            "dependsOn": ["CMake: Configure (Release)"]
        },

        {
            "label": "CMake: Build (Sanitizers)",
            "detail": "Compila com AddressSanitizer + UBSanitizer para detecção de bugs",
            "type": "shell",
            "command": "cmake",
            "args": [
                "--build",
                "--preset",
                "build-sanitize"
            ],
            "options": {
                "cwd": "${workspaceFolder}"
            },
            "group": "build",
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared",
                "showReuseMessage": false,
                "clear": true
            },
            "problemMatcher": ["$gcc"]
        },

        // ── Clean ──────────────────────────────────────────────────────────────

        {
            "label": "CMake: Clean",
            "detail": "Remove os artefatos compilados do último build Debug",
            "type": "shell",
            "command": "cmake",
            "args": [
                "--build",
                "--preset",
                "{{if .UseVCPKG}}build-vcpkg-debug{{else}}build-debug{{end}}",
                "--target",
                "clean"
            ],
            "options": {
                "cwd": "${workspaceFolder}"
            },
            "group": "build",
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared",
                "showReuseMessage": false,
                "clear": false
            },
            "problemMatcher": []
        },

        {
            "label": "CMake: Delete build dir",
            "detail": "Apaga completamente o diretório build/ para um configure limpo",
            "type": "shell",
            "command": "rm",
            "args": ["-rf", "${workspaceFolder}/build"],
            "windows": {
                "command": "Remove-Item",
                "args": ["-Recurse", "-Force", "${workspaceFolder}\\build"]
            },
            "options": {
                "cwd": "${workspaceFolder}"
            },
            "group": "build",
            "presentation": {
                "echo": true,
                "reveal": "silent",
                "focus": false,
                "panel": "shared",
                "showReuseMessage": false
            },
            "problemMatcher": []
        },

        // ── Rebuild ────────────────────────────────────────────────────────────

        {
            "label": "CMake: Rebuild (Debug)",
            "detail": "Clean + Build Debug em sequência",
            "group": "build",
            "dependsOrder": "sequence",
            "dependsOn": [
                "CMake: Clean",
                "CMake: Build (Debug)"
            ],
            "presentation": {
                "echo": false,
                "reveal": "always",
                "panel": "shared",
                "showReuseMessage": false
            },
            "problemMatcher": []
        },

        // ── Testes ─────────────────────────────────────────────────────────────

        {
            "label": "CMake: Run Tests",
            "detail": "Executa os testes com CTest (preset test-debug)",
            "type": "shell",
            "command": "ctest",
            "args": [
                "--preset",
                "{{if .UseVCPKG}}test-vcpkg-debug{{else}}test-debug{{end}}",
                "--output-on-failure"
            ],
            "options": {
                "cwd": "${workspaceFolder}"
            },
            "group": {
                "kind": "test",
                "isDefault": true
            },
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared",
                "showReuseMessage": false,
                "clear": true
            },
            "problemMatcher": [],
            "dependsOn": ["CMake: Build (Debug)"]
        },

        // ── Ferramentas ────────────────────────────────────────────────────────

        {
            "label": "Clang-Format: Format file",
            "detail": "Formata o arquivo atualmente aberto com clang-format",
            "type": "shell",
            "command": "clang-format",
            "args": ["-i", "${file}"],
            "options": {
                "cwd": "${workspaceFolder}"
            },
            "group": "none",
            "presentation": {
                "echo": true,
                "reveal": "silent",
                "focus": false,
                "panel": "shared",
                "showReuseMessage": false
            },
            "problemMatcher": []
        },

        {
            "label": "Clang-Format: Format project",
            "detail": "Formata todos os arquivos .cpp e .hpp do projeto",
            "type": "shell",
            "command": "find",
            "args": [
                "src", "include", "tests",
                "-name", "*.cpp",
                "-o", "-name", "*.hpp",
                "-o", "-name", "*.h",
                "|", "xargs", "clang-format", "-i"
            ],
            "options": {
                "cwd": "${workspaceFolder}",
                "shell": {
                    "executable": "/bin/sh",
                    "args": ["-c"]
                }
            },
            "group": "none",
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared",
                "showReuseMessage": false
            },
            "problemMatcher": []
        },

        {
            "label": "CMake: Symlink compile_commands.json",
            "detail": "Cria symlink de compile_commands.json na raiz (para Clangd)",
            "type": "shell",
            "command": "ln",
            "args": [
                "-sf",
                "build/{{if .UseVCPKG}}vcpkg-debug{{else}}debug{{end}}/compile_commands.json",
                "compile_commands.json"
            ],
            "options": {
                "cwd": "${workspaceFolder}"
            },
            "group": "none",
            "presentation": {
                "echo": true,
                "reveal": "silent",
                "focus": false,
                "panel": "shared",
                "showReuseMessage": false
            },
            "problemMatcher": []
        }
    ]
}
`

// ─────────────────────────────────────────────────────────────────────────────
// Template: .vscode/launch.json
// ─────────────────────────────────────────────────────────────────────────────

// tmplVSCodeLaunch é o template para .vscode/launch.json.
//
// Fornece configurações de debug para dois debuggers populares:
//
//  1. CodeLLDB (vadimcn.vscode-lldb)
//     - Recomendado para Linux e macOS com Clang
//     - Melhor integração com LLVM/Clang (tipos STL legíveis, etc.)
//     - Extensão: "vadimcn.vscode-lldb"
//
//  2. cppdbg — Microsoft C/C++ Extension (ms-vscode.cpptools)
//     - Suporta GDB e LLDB em todas as plataformas
//     - Preferido para Windows e projetos multi-plataforma
//     - Extensão: "ms-vscode.cpptools"
//
// Configurações incluídas:
//   - Debug (LLDB) — executa o binário com CodeLLDB
//   - Debug (GDB)  — executa o binário com cppdbg+GDB
//   - Debug Tests  — executa o binário de testes com debug
//   - Attach       — anexa a um processo em execução
//
// Referência: https://code.visualstudio.com/docs/cpp/launch-json-reference
const tmplVSCodeLaunch = `{
    // ==========================================================================
    // .vscode/launch.json — Configurações de debug para {{ .ProjectName }}
    // ==========================================================================
    // Documentação: https://code.visualstudio.com/docs/cpp/launch-json-reference
    //
    // Adaptador de debug configurado: {{if eq .DebugAdapter "both"}}LLDB + GDB{{else}}{{.DebugAdapter}}{{end}}
    //
{{- if .UseLLDB}}
    // ── CodeLLDB (LLDB) ───────────────────────────────────────────────────────
    // Extensão: vadimcn.vscode-lldb
    // Instale: Ctrl+Shift+X → pesquise "CodeLLDB"
{{- end}}
{{- if .UseGDB}}
    // ── C/C++ Extension (GDB) ─────────────────────────────────────────────────
    // Extensão: ms-vscode.cpptools
    // Instale: Ctrl+Shift+X → pesquise "C/C++"
{{- end}}
    //
    // Para iniciar o debug: F5  (usa a configuração ativa na barra de status)
    // ==========================================================================
    "version": "0.2.0",
    "configurations": [
{{if .IsExecutable}}
{{- if .UseLLDB}}
        // ── CodeLLDB — LLDB nativo (Linux/macOS com Clang) ────────────────────
        {
            "name": "Debug: {{ .ProjectName }} (LLDB)",
            "type": "lldb",
            "request": "launch",
            "program": "${workspaceFolder}/build/{{if .UseVCPKG}}vcpkg-debug{{else}}debug{{end}}/bin/{{ .ProjectName }}",
            "args": [],
            "cwd": "${workspaceFolder}",
            "env": {},
            "preLaunchTask": "CMake: Build (Debug)",
            "terminal": "integrated",
            "initCommands": [
                "settings set target.max-string-summary-length 256"
            ]
        },

        // ── Debug dos testes (LLDB) ────────────────────────────────────────────
        {
            "name": "Debug: Testes (LLDB)",
            "type": "lldb",
            "request": "launch",
            "program": "${workspaceFolder}/build/{{if .UseVCPKG}}vcpkg-debug{{else}}debug{{end}}/tests/{{ .ProjectName }}_tests",
            "args": [],
            "cwd": "${workspaceFolder}",
            "preLaunchTask": "CMake: Build (Debug)",
            "terminal": "integrated"
        },

        // ── Attach (LLDB) ──────────────────────────────────────────────────────
        {
            "name": "Debug: Attach to Process (LLDB)",
            "type": "lldb",
            "request": "attach",
            "pid": "${command:pickProcess}",
            "stopOnEntry": false
        },
{{- end}}
{{- if .UseGDB}}
        // ── cppdbg com GDB ────────────────────────────────────────────────────
        {
            "name": "Debug: {{ .ProjectName }} (GDB)",
            "type": "cppdbg",
            "request": "launch",
            "program": "${workspaceFolder}/build/{{if .UseVCPKG}}vcpkg-debug{{else}}debug{{end}}/bin/{{ .ProjectName }}",
            "args": [],
            "cwd": "${workspaceFolder}",
            "environment": [],
            "externalConsole": false,
            "MIMode": "gdb",
            "miDebuggerPath": "/usr/bin/gdb",
            "setupCommands": [
                {
                    "description": "Habilitar pretty-printing para GDB",
                    "text": "-enable-pretty-printing",
                    "ignoreFailures": true
                },
                {
                    "description": "Desabilitar paginação de saída",
                    "text": "set pagination off",
                    "ignoreFailures": true
                }
            ],
            "preLaunchTask": "CMake: Build (Debug)",
            "stopAtEntry": false,
            "logging": {
                "exceptions": true,
                "moduleLoad": false,
                "programOutput": true
            }
        },

        // ── Debug dos testes (GDB) ─────────────────────────────────────────────
        {
            "name": "Debug: Testes (GDB)",
            "type": "cppdbg",
            "request": "launch",
            "program": "${workspaceFolder}/build/{{if .UseVCPKG}}vcpkg-debug{{else}}debug{{end}}/tests/{{ .ProjectName }}_tests",
            "args": [],
            "cwd": "${workspaceFolder}",
            "externalConsole": false,
            "MIMode": "gdb",
            "setupCommands": [
                {
                    "description": "Habilitar pretty-printing",
                    "text": "-enable-pretty-printing",
                    "ignoreFailures": true
                }
            ],
            "preLaunchTask": "CMake: Build (Debug)"
        },
{{- end}}
{{else}}
        // ── Biblioteca: debug via testes ───────────────────────────────────────
{{- if .UseLLDB}}
        {
            "name": "Debug: Testes (LLDB)",
            "type": "lldb",
            "request": "launch",
            "program": "${workspaceFolder}/build/{{if .UseVCPKG}}vcpkg-debug{{else}}debug{{end}}/tests/{{ .ProjectName }}_tests",
            "args": [],
            "cwd": "${workspaceFolder}",
            "preLaunchTask": "CMake: Build (Debug)",
            "terminal": "integrated"
        },
{{- end}}
{{- if .UseGDB}}
        {
            "name": "Debug: Testes (GDB)",
            "type": "cppdbg",
            "request": "launch",
            "program": "${workspaceFolder}/build/{{if .UseVCPKG}}vcpkg-debug{{else}}debug{{end}}/tests/{{ .ProjectName }}_tests",
            "args": [],
            "cwd": "${workspaceFolder}",
            "externalConsole": false,
            "MIMode": "gdb",
            "setupCommands": [
                {
                    "description": "Habilitar pretty-printing",
                    "text": "-enable-pretty-printing",
                    "ignoreFailures": true
                }
            ],
            "preLaunchTask": "CMake: Build (Debug)"
        },
{{- end}}
{{end}}
    ],
    "compounds": []
}
`

// ─────────────────────────────────────────────────────────────────────────────
// Template: .vscode/settings.json
// ─────────────────────────────────────────────────────────────────────────────

// tmplVSCodeSettings é o template para .vscode/settings.json.
//
// Configura o workspace do VSCode com as seguintes categorias:
//
//	Clangd (servidor LSP):
//	  - Desabilita IntelliSense nativo da extensão C/C++ da Microsoft para
//	    evitar conflito com o Clangd (não execute os dois ao mesmo tempo!)
//	  - Configura argumentos do clangd (fallback flags, log level)
//
//	CMake Tools:
//	  - Define o preset de configure padrão
//	  - Configura o diretório de build
//
//	Editor:
//	  - Formatação automática ao salvar via clang-format
//	  - Trim de espaços em branco e newline final
//	  - Associação de extensões de arquivos C++
//
//	Arquivos:
//	  - Exclui diretórios de build da árvore de arquivos (não do Git)
//
// Referência: https://code.visualstudio.com/docs/getstarted/settings
const tmplVSCodeSettings = `{
    // ==========================================================================
    // .vscode/settings.json — Configurações do workspace {{ .ProjectName }}
    // ==========================================================================
    // Estas configurações aplicam-se apenas a este workspace.
    // Configurações pessoais ficam em ~/{Code/User}/settings.json.
    // ==========================================================================

    // ── Clangd LSP ────────────────────────────────────────────────────────────

    // Desabilita o IntelliSense nativo da extensão Microsoft C/C++ para evitar
    // conflito com o Clangd. O Clangd é mais preciso e usa o compile_commands.json
    // gerado pelo CMake — use apenas um dos dois ao mesmo tempo.
    "C_Cpp.intelliSenseEngine": "disabled",

    // Argumentos passados ao servidor Clangd ao iniciar.
    // Documentação: https://clangd.llvm.org/installation.html#editor-plugins
    "clangd.arguments": [
        // Usa o compile_commands.json do diretório especificado.
        "--compile-commands-dir=${workspaceFolder}/build/{{if .UseVCPKG}}vcpkg-debug{{else}}debug{{end}}",

        // Número de workers para indexação em background.
        "--j=4",

        // Habilita todas as opções de diagnóstico avançado.
        "--all-scopes-completion",

        // Sugestões de header ao completar símbolos desconhecidos.
        "--header-insertion=iwyu",

        // Exibe diagnósticos do clang-tidy inline no editor.
        "--clang-tidy",

        // Nível de log (error | warning | info | verbose).
        "--log=error",

        // Usa o cache em disco para acelerar inicializações subsequentes.
        "--enable-config"
    ],

    // Caminho para o binário clangd (descomente se não estiver no PATH).
    // "clangd.path": "/usr/bin/clangd-18",

    // Reinicia o servidor Clangd automaticamente quando o compile_commands.json mudar.
    "clangd.restartAfterCrash": true,

    // ── CMake Tools ───────────────────────────────────────────────────────────

    // Preset de configure padrão ao abrir o projeto.
    "cmake.configurePreset": "{{if .UseVCPKG}}vcpkg-debug{{else}}debug{{end}}",

    // Preset de build padrão.
    "cmake.buildPreset": "{{if .UseVCPKG}}build-vcpkg-debug{{else}}build-debug{{end}}",

    // Não exibe a notificação de "CMake configurado com sucesso" a cada abertura.
    "cmake.configureOnOpen": false,

    // Oculta a barra de status do CMake Tools (descomente para exibir).
    // "cmake.statusBarVisibility": "hidden",

    // Salva automaticamente arquivos modificados antes do build.
    "cmake.saveBeforeBuild": true,

    // ── Editor — Formatação ───────────────────────────────────────────────────

    // Formata o arquivo automaticamente ao salvar.
    "[cpp]": {
        "editor.formatOnSave": true,
        "editor.defaultFormatter": "llvm-vs-code-extensions.vscode-clangd",
        "editor.tabSize": 4,
        "editor.insertSpaces": true
    },

    "[c]": {
        "editor.formatOnSave": true,
        "editor.defaultFormatter": "llvm-vs-code-extensions.vscode-clangd",
        "editor.tabSize": 4,
        "editor.insertSpaces": true
    },

    // ── Editor — Geral ────────────────────────────────────────────────────────

    // Remove espaços em branco no final das linhas ao salvar.
    "files.trimTrailingWhitespace": true,

    // Insere uma linha em branco no final do arquivo ao salvar.
    "files.insertFinalNewline": true,

    // Remove linhas em branco extras no final do arquivo.
    "files.trimFinalNewlines": true,

    // Codificação padrão para novos arquivos.
    "files.encoding": "utf8",

    // Fim de linha padrão (lf = Unix, crlf = Windows, auto = detecta).
    "files.eol": "\n",

    // ── Associações de arquivos ───────────────────────────────────────────────

    // Garante que o VSCode reconheça extensões incomuns como C++.
    "files.associations": {
        "*.hpp": "cpp",
        "*.cpp": "cpp",
        "*.cxx": "cpp",
        "*.cc":  "cpp",
        "*.hxx": "cpp",
        "*.inl": "cpp",
        "*.tpp": "cpp",
        "CMakeLists.txt": "cmake",
        "*.cmake": "cmake"
    },

    // ── Exclusão de arquivos da árvore ────────────────────────────────────────
    // Oculta diretórios de build e cache da árvore de arquivos do VSCode.
    // NOTA: isto não afeta o .gitignore — apenas a visualização no editor.
    "files.exclude": {
        "build/":              true,
        "install/":            true,
        "vcpkg_installed/":    true,
        "**/.cache/":          true,
        "**/*.o":              true,
        "**/*.a":              true
    },

    // ── Busca — Exclusão ──────────────────────────────────────────────────────
    // Exclui diretórios pesados das buscas (Ctrl+Shift+F) para melhorar o desempenho.
    "search.exclude": {
        "build/**":            true,
        "install/**":          true,
        "vcpkg_installed/**":  true,
        ".git/**":             true
    },

    // ── Inlay Hints ───────────────────────────────────────────────────────────
    // Exibe dicas de tipo e parâmetro inline no código (requer suporte do clangd).
    "editor.inlayHints.enabled": "on",

    // ── Terminal ──────────────────────────────────────────────────────────────
    // Define a pasta de trabalho padrão do terminal integrado.
    "terminal.integrated.cwd": "${workspaceFolder}"
}
`

// ─────────────────────────────────────────────────────────────────────────────
// Template: .vscode/extensions.json
// ─────────────────────────────────────────────────────────────────────────────

// tmplVSCodeExtensions é o template para .vscode/extensions.json.
//
// Lista as extensões recomendadas para o projeto. O VSCode exibe uma
// notificação automática ao abrir o workspace, sugerindo a instalação.
//
// Extensões incluídas:
//
//	Essenciais:
//	  - llvm-vs-code-extensions.vscode-clangd   LSP C++ via Clangd
//	  - ms-vscode.cmake-tools                   Integração com CMake
//	  - twxs.cmake                              Syntax highlight CMake
//
//	Debug:
//	  - vadimcn.vscode-lldb                     Debug LLDB (recomendado)
//	  - ms-vscode.cpptools                      Debug GDB/LLDB (alternativa)
//
//	Qualidade de código:
//	  - notskm.clang-tidy                       Clang-tidy no editor
//	  - xaver.clang-format                      Formatação clang-format
//
//	Produtividade:
//	  - eamodio.gitlens                         Histórico Git inline
//	  - streetsidesoftware.code-spell-checker   Verificador ortográfico
//	  - usernamehw.errorlens                    Erros inline no código
//
// Referência: https://code.visualstudio.com/docs/editor/extension-marketplace
const tmplVSCodeExtensions = `{
    // ==========================================================================
    // .vscode/extensions.json — Extensões recomendadas para {{ .ProjectName }}
    // ==========================================================================
    // O VSCode exibirá uma notificação para instalar estas extensões ao abrir
    // o workspace pela primeira vez.
    //
    // Para instalar manualmente:
    //   Ctrl+Shift+P → "Extensions: Show Recommended Extensions"
    // ==========================================================================
    "recommendations": [

        // ── C++ e CMake (essenciais) ──────────────────────────────────────────

        // Servidor LSP Clangd — autocompletar, diagnósticos, navegação de código.
        // Usa o compile_commands.json gerado pelo CMake para máxima precisão.
        // IMPORTANTE: Desabilite a extensão ms-vscode.cpptools IntelliSense
        // se usar o Clangd para evitar conflitos (já configurado no settings.json).
        "llvm-vs-code-extensions.vscode-clangd",

        // CMake Tools — integração completa com CMake: configure, build, test,
        // seleção de preset, status bar e muito mais.
        "ms-vscode.cmake-tools",

        // Syntax highlighting e snippets para CMakeLists.txt e arquivos .cmake.
        "twxs.cmake",

        // ── Debug ─────────────────────────────────────────────────────────────

        // CodeLLDB — debugger LLDB nativo para VSCode.
        // Recomendado para Linux e macOS com Clang. Excelente suporte a Rust também.
        // Tipos STL são exibidos de forma legível sem configuração adicional.
        "vadimcn.vscode-lldb",

        // Microsoft C/C++ — extensão oficial com suporte a GDB, LLDB e MSVC.
        // Use como alternativa ao CodeLLDB ou para projetos Windows.
        // NOTA: Desabilite o IntelliSense desta extensão se usar o Clangd.
        "ms-vscode.cpptools",

        // ── Qualidade de código ───────────────────────────────────────────────

        // Exibe diagnósticos do clang-tidy diretamente no editor.
        // Requer clang-tidy instalado e configurado no .clangd.
        "notskm.clang-tidy",

        // ── Git e colaboração ─────────────────────────────────────────────────

        // GitLens — histórico de commits inline, blame, comparação de branches.
        "eamodio.gitlens",

        // ── Produtividade ─────────────────────────────────────────────────────

        // Error Lens — exibe mensagens de erro e warning inline na linha do código,
        // eliminando a necessidade de passar o mouse sobre os sublinhados vermelhos.
        "usernamehw.errorlens",

        // Code Spell Checker — verificação ortográfica em código e comentários.
        // Suporta camelCase, snake_case e PascalCase nativamente.
        "streetsidesoftware.code-spell-checker",

        // Highlight de pares de chaves/colchetes com cores diferentes por nível.
        "oderwat.indent-rainbow",

        // Better Comments — coloriza comentários por tipo (TODO, FIXME, !, ?).
        "aaron-bond.better-comments",

        // Hex Editor — útil para visualizar arquivos binários e buffers de memória.
        "ms-vscode.hexeditor"
    ],

    // Extensões NÃO recomendadas — listadas para evitar conflitos conhecidos.
    "unwantedRecommendations": [
        // O IntelliSense da extensão Microsoft C/C++ conflita com o Clangd.
        // A extensão em si é útil para debug (cppdbg), mas o IntelliSense
        // deve ser desabilitado via "C_Cpp.intelliSenseEngine": "disabled".
        // Não está na unwanted list para permitir o uso do debugger cppdbg.
    ]
}
`

// ─────────────────────────────────────────────────────────────────────────────
// Template: .vscode/c_cpp_properties.json
// ─────────────────────────────────────────────────────────────────────────────

// tmplVSCodeCppProperties é o template para .vscode/c_cpp_properties.json.
//
// Este arquivo configura o provedor de IntelliSense da extensão Microsoft C/C++.
// Quando o Clangd está ativo, este arquivo é secundário (o Clangd ignora ele),
// mas é mantido como fallback caso o Clangd não esteja disponível ou para
// usuários que preferem o IntelliSense nativo da Microsoft.
//
// Nota: O Clangd usa o compile_commands.json diretamente, que é muito mais
// preciso que as configurações manuais deste arquivo.
//
// Referência: https://code.visualstudio.com/docs/cpp/c-cpp-properties-schema-reference
const tmplVSCodeCppProperties = `{
    // ==========================================================================
    // .vscode/c_cpp_properties.json — IntelliSense fallback para {{ .ProjectName }}
    // ==========================================================================
    // NOTA: Se você usa o Clangd (recomendado), este arquivo é ignorado pelo LSP.
    // É mantido como fallback para o IntelliSense da extensão ms-vscode.cpptools
    // e para IDEs que leem este formato.
    //
    // Para uma configuração mais precisa, garanta que o compile_commands.json
    // esteja sendo gerado pelo CMake (CMAKE_EXPORT_COMPILE_COMMANDS = ON).
    // ==========================================================================
    "version": 4,
    "configurations": [
        {
            "name": "Linux",
            "includePath": [
                "${workspaceFolder}/include",
                "${workspaceFolder}/src",
                "${workspaceFolder}/build/{{if .UseVCPKG}}vcpkg-debug{{else}}debug{{end}}/_deps/**/include",
                "/usr/include",
                "/usr/local/include"
            ],
            "defines": [
                "{{.NameUpper}}_DEBUG=1"
            ],
            // Caminho para o compile_commands.json gerado pelo CMake.
            // Quando disponível, o IntelliSense usa este arquivo em vez das
            // configurações manuais acima.
            "compileCommands": "${workspaceFolder}/build/{{if .UseVCPKG}}vcpkg-debug{{else}}debug{{end}}/compile_commands.json",
            "compilerPath": "/usr/bin/clang++",
            "cStandard": "c17",
            "cppStandard": "c++{{.ProjectName}}",
            "intelliSenseMode": "linux-clang-x64"
        },
        {
            "name": "macOS",
            "includePath": [
                "${workspaceFolder}/include",
                "${workspaceFolder}/src",
                "${workspaceFolder}/build/{{if .UseVCPKG}}vcpkg-debug{{else}}debug{{end}}/_deps/**/include"
            ],
            "defines": [
                "{{.NameUpper}}_DEBUG=1"
            ],
            "compileCommands": "${workspaceFolder}/build/{{if .UseVCPKG}}vcpkg-debug{{else}}debug{{end}}/compile_commands.json",
            "compilerPath": "/usr/bin/clang++",
            "cStandard": "c17",
            "cppStandard": "c++20",
            "intelliSenseMode": "macos-clang-arm64",
            "macFrameworkPath": [
                "/Library/Developer/CommandLineTools/SDKs/MacOSX.sdk/System/Library/Frameworks"
            ]
        },
        {
            "name": "Windows",
            "includePath": [
                "${workspaceFolder}/include",
                "${workspaceFolder}/src",
                "${workspaceFolder}/build/{{if .UseVCPKG}}vcpkg-debug{{else}}debug{{end}}/_deps/**/include"
            ],
            "defines": [
                "{{.NameUpper}}_DEBUG=1",
                "_WIN32",
                "UNICODE",
                "_UNICODE"
            ],
            "compileCommands": "${workspaceFolder}/build/{{if .UseVCPKG}}vcpkg-debug{{else}}debug{{end}}/compile_commands.json",
            "compilerPath": "cl.exe",
            "cStandard": "c17",
            "cppStandard": "c++20",
            "intelliSenseMode": "windows-msvc-x64",
            "windowsSdkVersion": "10.0.19041.0"
        }
    ]
}
`
