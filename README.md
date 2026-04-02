# cpp-gen

> Modern C++ project generator with CMake, package managers, IDE configurations and development tools.

![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-MIT-green)
![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)

---

## Table of Contents

- [About](#about)
- [Installation](#installation)
- [Usage](#usage)
- [What is generated](#what-is-generated)
- [Options and flags](#options-and-flags)
- [Project structure](#project-structure)
- [Development](#development)
- [Internal architecture](#internal-architecture)
- [License](#license)

---

## About

`cpp-gen` is a CLI tool written in Go that automates the creation of modern C++ projects,
eliminating the time spent on the initial configuration of:

- **CMake** hierarchical with best practices (CMake 3.20+, CMakePresets.json)
- **Package managers**: VCPKG (manifest mode) or native FetchContent
- **IDEs**: Visual Studio Code, CLion or Neovim with complete configurations
- **Quality tools**: Clangd LSP, Clang-Format, warning flags
- **Git**: initialized repository, comprehensive `.gitignore` and README

---

## Installation

### Arch Linux (AUR) — recommended

```bash
# With yay
yay -S cpp-gen-bin

# With paru
paru -S cpp-gen-bin
```

### macOS / Linux — Homebrew

```bash
brew tap matpdev/tap
brew install cpp-gen
```

### Binary (Linux / macOS)

Download the latest pre-compiled binary from [GitHub Releases](https://github.com/matpdev/cpp-gen/releases/latest):

```bash
# Linux x86_64
curl -LO https://github.com/matpdev/cpp-gen/releases/latest/download/cpp-gen_linux_amd64.tar.gz
tar -xzf cpp-gen_linux_amd64.tar.gz
install -m755 cpp-gen ~/.local/bin/

# Verify checksum
sha256sum -c checksums.txt
```

### Go install

```bash
go install github.com/matpdev/cpp-gen@latest
```

### From source

```bash
git clone https://github.com/matpdev/cpp-gen.git
cd cpp-gen
go mod tidy
go build -o cpp-gen .
```

---

## Usage

### Interactive mode (recommended)

```bash
# Opens the step-by-step TUI form
cpp-gen new

# With pre-filled name
cpp-gen new meu-projeto
```

The form guides you through all the options:

```
⚡ cpp-gen  v0.1.0
─────────────────────────────────────────────────────
  ◆  Modern CMake project structure (3.20+)
  ◆  Package managers: VCPKG or FetchContent
  ◆  Configurations for VSCode, CLion and Neovim
  ◆  Git, .gitignore and README ready
  ◆  Clangd and Clang-Format pre-configured
─────────────────────────────────────────────────────

Usage:  cpp-gen new [project-name]
```

### Non-interactive mode (CI/CD and scripts)

```bash
cpp-gen new meu-projeto \
  --no-interactive \
  --description "My C++ application" \
  --author "John Doe" \
  --std 20 \
  --type executable \
  --pkg vcpkg \
  --ide vscode
```

### Other commands

```bash
# Display version
cpp-gen version

# General help
cpp-gen --help

# Help for the new subcommand
cpp-gen new --help
```

---

## What is generated

### Example: executable project with VSCode and VCPKG

```
meu-projeto/
├── CMakeLists.txt              ← Main CMake configuration
├── CMakePresets.json           ← debug/release/sanitize/vcpkg presets
├── vcpkg.json                  ← VCPKG dependencies (manifest mode)
├── vcpkg-configuration.json   ← Version baseline (reproducible builds)
├── README.md                  ← Generated project README
├── .gitignore                 ← Comprehensive C++/CMake/IDE patterns
├── .clangd                    ← LSP configuration (compile_commands.json)
├── .clang-format              ← Formatting rules (LLVM-based)
│
├── cmake/
│   ├── CompilerWarnings.cmake  ← Warning flags (GCC/Clang/MSVC)
│   ├── Vcpkg.cmake            ← VCPKG integration helper module
│   └── Dependencies.cmake     ← (if FetchContent) declared dependencies
│
├── src/
│   ├── CMakeLists.txt         ← add_executable() or add_library() target
│   └── main.cpp               ← Initial source code
│
├── include/
│   └── meu-projeto/           ← Public headers (include namespace)
│
├── tests/
│   ├── CMakeLists.txt         ← Test target with CTest
│   └── test_main.cpp          ← Initial tests with CHECK() macro
│
├── docs/                      ← Documentation (empty, ready for Doxygen)
│
└── .vscode/
    ├── tasks.json             ← Configure, Build, Clean, Test, Format
    ├── launch.json            ← Debug with CodeLLDB and cppdbg/GDB
    ├── settings.json          ← Clangd, CMake Tools, automatic formatting
    ├── extensions.json        ← Recommended extensions
    └── c_cpp_properties.json  ← IntelliSense fallback
```

### Generated CMakePresets.json

| Configure Preset      | Description                                       |
|-----------------------|---------------------------------------------------|
| `debug`               | Debug with full symbols                           |
| `release`             | Release with optimizations, no tests              |
| `release-with-debug`  | RelWithDebInfo (profiling)                        |
| `sanitize`            | Debug + AddressSanitizer + UBSanitizer            |
| `vcpkg-debug`         | Debug with VCPKG toolchain *(if VCPKG selected)*  |
| `vcpkg-release`       | Release with VCPKG *(if VCPKG selected)*          |

```bash
# List all presets
cmake --list-presets

# Quick build
cmake --preset debug
cmake --build --preset build-debug
ctest --preset test-debug --output-on-failure
```

---

## Options and flags

### `cpp-gen new`

| Flag                     | Default      | Description                                             |
|--------------------------|:------------:|---------------------------------------------------------|
| `--output`, `-o`         | `.`          | Directory where the project folder will be created      |
| `--no-interactive`, `-n` | `false`      | Disables the TUI; uses only the flags below             |
| `--name`                 | —            | Project name (alternative to the positional argument)   |
| `--description`          | —            | Brief project description                               |
| `--author`               | —            | Author or organization name                             |
| `--version`              | `1.0.0`      | Initial version (SemVer)                                |
| `--std`                  | `20`         | C++ standard: `17` \| `20` \| `23`                     |
| `--type`                 | `executable` | `executable` \| `static-lib` \| `header-only`           |
| `--pkg`                  | `none`       | `none` \| `vcpkg` \| `fetchcontent`                     |
| `--ide`                  | `none`       | `none` \| `vscode` \| `clion` \| `nvim`                 |
| `--no-git`               | `false`      | Do not initialize a Git repository                      |
| `--no-clangd`            | `false`      | Do not generate `.clangd`                               |
| `--no-clang-format`      | `false`      | Do not generate `.clang-format`                         |

### Global flags

| Flag              | Description                                     |
|-------------------|-------------------------------------------------|
| `--verbose`, `-v` | Displays each file generated during the process |
| `--help`, `-h`    | Displays command help                           |

---

## Project structure

```
cpp-gen/
├── main.go                         ← Entry point
├── go.mod                          ← Go module and dependencies
│
├── cmd/
│   ├── root.go                     ← Root command (banner, version)
│   └── new.go                      ← `new` subcommand (flags, handler, TUI)
│
└── internal/
    ├── config/
    │   └── config.go               ← Enumerated types and ProjectConfig
    │
    ├── tui/
    │   ├── form.go                 ← Interactive form (charmbracelet/huh)
    │   └── styles.go               ← lipgloss styles (colors, layout)
    │
    └── generator/
        ├── generator.go            ← Orchestrator, TemplateData, utilities
        ├── structure.go            ← Folder structure and initial C++ files
        ├── cmake.go                ← CMakeLists.txt, CMakePresets.json, helpers
        ├── git.go                  ← Git init, .gitignore, README.md
        ├── clang.go                ← .clangd, .clang-format
        │
        ├── ide/
        │   ├── ide.go              ← Data interface, public functions, utilities
        │   ├── vscode.go           ← tasks.json, launch.json, settings, extensions
        │   └── clion.go            ← .idea/, cmake.xml, run configs, .nvim.lua
        │
        └── packages/
            ├── vcpkg.go            ← vcpkg.json, vcpkg-configuration.json, Vcpkg.cmake
            └── fetchcontent.go     ← cmake/Dependencies.cmake with commented examples
```

### Execution flow

```
main()
  └── cmd.Execute()
        └── newCmd.RunE  (cmd/new.go)
              ├── tui.RunForm()          ← interactive form
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

## Development

### Set up the environment

```bash
git clone https://github.com/matpdev/cpp-gen.git
cd cpp-gen
go mod tidy
```

### Run without installing

```bash
go run . new meu-projeto
```

### Build

```bash
go build -o cpp-gen .
./cpp-gen new --help
```

### Tests

```bash
go test ./...
go test ./... -v          # verbose
go test ./... -count=1    # disable test cache
```

### Check errors and lint

```bash
go vet ./...
# With golangci-lint installed:
golangci-lint run
```

### Direct dependencies

| Package                             | Version  | Usage                              |
|-------------------------------------|----------|------------------------------------|
| `github.com/spf13/cobra`            | v1.8.1   | CLI framework (commands and flags) |
| `github.com/charmbracelet/huh`      | v0.6.0   | Interactive TUI forms              |
| `github.com/charmbracelet/lipgloss` | v1.0.0   | Terminal styles and colors         |

---

## Internal architecture

### Separation of responsibilities

| Package                       | Responsibility                                               |
|-------------------------------|--------------------------------------------------------------|
| `cmd`                         | CLI interface: flag parsing, validation, orchestration       |
| `internal/config`             | Pure data types, no I/O logic                                |
| `internal/tui`                | Interactive user interface (no generation logic)             |
| `internal/generator`          | All file generation logic                                    |
| `internal/generator/ide`      | IDE-specific configurations (isolated per IDE)               |
| `internal/generator/packages` | Package manager configurations (isolated per pkg)            |

### Adding support for a new IDE

1. Create `internal/generator/ide/myide.go` with the `generateMyIDE()` function
2. Add the `IDEMyIDE` constant in `internal/config/config.go`
3. Add the option in the TUI form in `internal/tui/form.go`
4. Add the case in `generator.runIDE()` in `internal/generator/generator.go`
5. Add the parser in `cmd/new.go` in `parseIDE()`

### Adding support for a new package manager

1. Create `internal/generator/packages/mypkg.go` with `GenerateMyPkg()`
2. Add the `PkgMyPkg` constant in `internal/config/config.go`
3. Add the case in the TUI form and in `generator.runPackages()`

---

## Contributing

1. Fork the repository
2. Create a branch: `git checkout -b feature/my-feature`
3. Commit: `git commit -m 'feat: add support for XYZ'`
4. Push: `git push origin feature/my-feature`
5. Open a Pull Request

### Commit convention (Conventional Commits)

- `feat:` — new feature
- `fix:` — bug fix
- `docs:` — documentation
- `refactor:` — refactoring without behavior change
- `test:` — adding or fixing tests
- `chore:` — maintenance tasks

---

## License

MIT © 2025 — See [LICENSE](LICENSE) for details.

---

*Made with ❤️ and Go.*
