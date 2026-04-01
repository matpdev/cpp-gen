# cpp-gen

> Gerador moderno de projetos C++ com CMake, gerenciadores de pacotes, configurações de IDE e ferramentas de desenvolvimento.

![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-MIT-green)
![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)

---

## Sumário

- [Sobre](#sobre)
- [Instalação](#instalação)
- [Uso](#uso)
- [O que é gerado](#o-que-é-gerado)
- [Opções e flags](#opções-e-flags)
- [Estrutura do projeto gerado](#estrutura-do-projeto-gerado)
- [Desenvolvimento](#desenvolvimento)
- [Arquitetura interna](#arquitetura-interna)
- [Licença](#licença)

---

## Sobre

`cpp-gen` é uma ferramenta CLI escrita em Go que automatiza a criação de projetos C++ modernos,
eliminando o tempo gasto na configuração inicial de:

- **CMake** hierárquico com boas práticas (CMake 3.20+, CMakePresets.json)
- **Gerenciadores de pacotes**: VCPKG (manifest mode) ou FetchContent nativo
- **IDEs**: Visual Studio Code, CLion ou Neovim com configurações completas
- **Ferramentas de qualidade**: Clangd LSP, Clang-Format, flags de warning
- **Git**: repositório inicializado, `.gitignore` abrangente e README

---

## Instalação

### Pré-requisitos

- [Go 1.22+](https://go.dev/dl/)

### A partir do código-fonte

```bash
git clone https://github.com/cpp-gen/cpp-gen.git
cd cpp-gen
go mod tidy
go build -o cpp-gen .
```

### Instalar globalmente

```bash
go install cpp-gen@latest
```

---

## Uso

### Modo interativo (recomendado)

```bash
# Abre o formulário TUI passo a passo
cpp-gen new

# Com nome pré-preenchido
cpp-gen new meu-projeto
```

O formulário guia você por todas as opções:

```
⚡ cpp-gen  v0.1.0
─────────────────────────────────────────────────────
  ◆  Estrutura de projeto CMake moderna (3.20+)
  ◆  Gerenciadores de pacotes: VCPKG ou FetchContent
  ◆  Configurações para VSCode, CLion e Neovim
  ◆  Git, .gitignore e README prontos
  ◆  Clangd e Clang-Format pré-configurados
─────────────────────────────────────────────────────

Uso:  cpp-gen new [nome-do-projeto]
```

### Modo não-interativo (CI/CD e scripts)

```bash
cpp-gen new meu-projeto \
  --no-interactive \
  --description "Minha aplicação C++" \
  --author "Fulano" \
  --std 20 \
  --type executable \
  --pkg vcpkg \
  --ide vscode
```

### Outros comandos

```bash
# Exibe versão
cpp-gen version

# Ajuda geral
cpp-gen --help

# Ajuda do subcomando new
cpp-gen new --help
```

---

## O que é gerado

### Exemplo: projeto executável com VSCode e VCPKG

```
meu-projeto/
├── CMakeLists.txt              ← Configuração CMake principal
├── CMakePresets.json           ← Presets debug/release/sanitize/vcpkg
├── vcpkg.json                  ← Dependências VCPKG (manifest mode)
├── vcpkg-configuration.json   ← Baseline de versões (builds reprodutíveis)
├── README.md                  ← README do projeto gerado
├── .gitignore                 ← Padrões C++/CMake/IDE abrangentes
├── .clangd                    ← Configuração LSP (compile_commands.json)
├── .clang-format              ← Regras de formatação (baseado em LLVM)
│
├── cmake/
│   ├── CompilerWarnings.cmake  ← Flags de warning (GCC/Clang/MSVC)
│   ├── Vcpkg.cmake            ← Módulo auxiliar de integração VCPKG
│   └── Dependencies.cmake     ← (se FetchContent) dependências declaradas
│
├── src/
│   ├── CMakeLists.txt         ← Target add_executable() ou add_library()
│   └── main.cpp               ← Código fonte inicial
│
├── include/
│   └── meu-projeto/           ← Headers públicos (namespace de include)
│
├── tests/
│   ├── CMakeLists.txt         ← Target de testes com CTest
│   └── test_main.cpp          ← Testes iniciais com macro CHECK()
│
├── docs/                      ← Documentação (vazio, pronto para Doxygen)
│
└── .vscode/
    ├── tasks.json             ← Configure, Build, Clean, Test, Format
    ├── launch.json            ← Debug com CodeLLDB e cppdbg/GDB
    ├── settings.json          ← Clangd, CMake Tools, formatação automática
    ├── extensions.json        ← Extensões recomendadas
    └── c_cpp_properties.json  ← IntelliSense fallback
```

### CMakePresets.json gerado

| Preset de Configure   | Descrição                                        |
|-----------------------|--------------------------------------------------|
| `debug`               | Debug com símbolos completos                     |
| `release`             | Release com otimizações, sem testes              |
| `release-with-debug`  | RelWithDebInfo (profiling)                       |
| `sanitize`            | Debug + AddressSanitizer + UBSanitizer           |
| `vcpkg-debug`         | Debug com toolchain VCPKG *(se VCPKG selecionado)* |
| `vcpkg-release`       | Release com VCPKG *(se VCPKG selecionado)*       |

```bash
# Listar todos os presets
cmake --list-presets

# Build rápido
cmake --preset debug
cmake --build --preset build-debug
ctest --preset test-debug --output-on-failure
```

---

## Opções e flags

### `cpp-gen new`

| Flag                  | Padrão       | Descrição                                              |
|-----------------------|:------------:|--------------------------------------------------------|
| `--output`, `-o`      | `.`          | Diretório onde a pasta do projeto será criada          |
| `--no-interactive`, `-n` | `false`   | Desativa o TUI; usa apenas as flags abaixo             |
| `--name`              | —            | Nome do projeto (alternativa ao argumento posicional)  |
| `--description`       | —            | Descrição breve do projeto                             |
| `--author`            | —            | Nome do autor ou organização                           |
| `--version`           | `1.0.0`      | Versão inicial (SemVer)                                |
| `--std`               | `20`         | Padrão C++: `17` \| `20` \| `23`                      |
| `--type`              | `executable` | `executable` \| `static-lib` \| `header-only`          |
| `--pkg`               | `none`       | `none` \| `vcpkg` \| `fetchcontent`                    |
| `--ide`               | `none`       | `none` \| `vscode` \| `clion` \| `nvim`                |
| `--no-git`            | `false`      | Não inicializar repositório Git                        |
| `--no-clangd`         | `false`      | Não gerar `.clangd`                                    |
| `--no-clang-format`   | `false`      | Não gerar `.clang-format`                              |

### Flags globais

| Flag             | Descrição                                    |
|------------------|----------------------------------------------|
| `--verbose`, `-v` | Exibe cada arquivo gerado durante o processo |
| `--help`, `-h`   | Exibe ajuda do comando                       |

---

## Estrutura do projeto

```
cpp-gen/
├── main.go                         ← Ponto de entrada
├── go.mod                          ← Módulo Go e dependências
│
├── cmd/
│   ├── root.go                     ← Comando raiz (banner, versão)
│   └── new.go                      ← Subcomando `new` (flags, handler, TUI)
│
└── internal/
    ├── config/
    │   └── config.go               ← Tipos enumerados e ProjectConfig
    │
    ├── tui/
    │   ├── form.go                 ← Formulário interativo (charmbracelet/huh)
    │   └── styles.go               ← Estilos lipgloss (cores, layout)
    │
    └── generator/
        ├── generator.go            ← Orquestrador, TemplateData, utilitários
        ├── structure.go            ← Estrutura de pastas e arquivos C++ iniciais
        ├── cmake.go                ← CMakeLists.txt, CMakePresets.json, helpers
        ├── git.go                  ← Git init, .gitignore, README.md
        ├── clang.go                ← .clangd, .clang-format
        │
        ├── ide/
        │   ├── ide.go              ← Interface Data, funções públicas, utilitários
        │   ├── vscode.go           ← tasks.json, launch.json, settings, extensions
        │   └── clion.go            ← .idea/, cmake.xml, run configs, .nvim.lua
        │
        └── packages/
            ├── vcpkg.go            ← vcpkg.json, vcpkg-configuration.json, Vcpkg.cmake
            └── fetchcontent.go     ← cmake/Dependencies.cmake com exemplos comentados
```

### Fluxo de execução

```
main()
  └── cmd.Execute()
        └── newCmd.RunE  (cmd/new.go)
              ├── tui.RunForm()          ← formulário interativo
              ├── cfg.Validate()
              ├── printProjectSummary()
              └── generator.New(cfg).Generate()
                    ├── generateStructure()   → src/, include/, tests/, docs/
                    ├── generateCMake()       → CMakeLists.txt, presets, helpers
                    ├── runPackages()         → vcpkg.json  | Dependencies.cmake
                    ├── runIDE()              → .vscode/    | .idea/  | .nvim.lua
                    ├── generateClang()       → .clangd     | .clang-format
                    └── generateGit()         → .gitignore  | README.md | git init
```

---

## Desenvolvimento

### Configurar o ambiente

```bash
git clone https://github.com/cpp-gen/cpp-gen.git
cd cpp-gen
go mod tidy
```

### Executar sem instalar

```bash
go run . new meu-projeto
```

### Build

```bash
go build -o cpp-gen .
./cpp-gen new --help
```

### Testes

```bash
go test ./...
go test ./... -v          # verbose
go test ./... -count=1    # desabilita cache de testes
```

### Verificar erros e lint

```bash
go vet ./...
# Com golangci-lint instalado:
golangci-lint run
```

### Dependências diretas

| Pacote                            | Versão   | Uso                                     |
|-----------------------------------|----------|-----------------------------------------|
| `github.com/spf13/cobra`          | v1.8.1   | Framework de CLI (comandos e flags)     |
| `github.com/charmbracelet/huh`    | v0.6.0   | Formulários TUI interativos             |
| `github.com/charmbracelet/lipgloss` | v1.0.0 | Estilos e cores no terminal             |

---

## Arquitetura interna

### Separação de responsabilidades

| Pacote                        | Responsabilidade                                              |
|-------------------------------|---------------------------------------------------------------|
| `cmd`                         | Interface CLI: parsing de flags, validação, orquestração      |
| `internal/config`             | Tipos de dados puros, sem lógica de I/O                       |
| `internal/tui`                | Interface de usuário interativa (sem lógica de geração)       |
| `internal/generator`          | Toda a lógica de geração de arquivos                          |
| `internal/generator/ide`      | Configurações específicas de IDE (isoladas por IDE)           |
| `internal/generator/packages` | Configurações de gerenciadores de pacotes (isoladas por pkg)  |

### Adicionando suporte a uma nova IDE

1. Crie `internal/generator/ide/minhaide.go` com a função `generateMinhaIDE()`
2. Adicione a constante `IDEMinhaIDE` em `internal/config/config.go`
3. Adicione a opção no formulário TUI em `internal/tui/form.go`
4. Adicione o case em `generator.runIDE()` em `internal/generator/generator.go`
5. Adicione o parser em `cmd/new.go` em `parseIDE()`

### Adicionando suporte a um novo gerenciador de pacotes

1. Crie `internal/generator/packages/meupkg.go` com `GenerateMeuPkg()`
2. Adicione a constante `PkgMeuPkg` em `internal/config/config.go`
3. Adicione o case no formulário TUI e no `generator.runPackages()`

---

## Contribuindo

1. Fork o repositório
2. Crie uma branch: `git checkout -b feature/minha-feature`
3. Commit: `git commit -m 'feat: adiciona suporte a XYZ'`
4. Push: `git push origin feature/minha-feature`
5. Abra um Pull Request

### Convenção de commits (Conventional Commits)

- `feat:` — nova funcionalidade
- `fix:` — correção de bug
- `docs:` — documentação
- `refactor:` — refatoração sem mudança de comportamento
- `test:` — adição ou correção de testes
- `chore:` — tarefas de manutenção

---

## Licença

MIT © 2025 — Veja [LICENSE](LICENSE) para detalhes.

---

*Feito com ❤️ e Go.*
