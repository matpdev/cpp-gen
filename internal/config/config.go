// Package config define todos os tipos, constantes e estruturas de configuração
// utilizados pelo cpp-gen para descrever um projeto C++ a ser gerado.
package config

import "path/filepath"

// ─────────────────────────────────────────────────────────────────────────────
// Tipos enumerados
// ─────────────────────────────────────────────────────────────────────────────

// IDE representa a IDE alvo para geração de configurações específicas
// (tasks, launch, settings, etc.).
type IDE string

const (
	IDENone   IDE = "none"   // Sem configuração de IDE
	IDEVSCode IDE = "vscode" // Visual Studio Code (.vscode/)
	IDECLion  IDE = "clion"  // CLion / CMakePresets.json
	IDENvim   IDE = "nvim"   // Neovim (.nvim.lua + clangd)
	IDEZed    IDE = "zed"    // Zed (.zed/settings.json + .zed/tasks.json)
)

// IDEOptions retorna todas as IDEs disponíveis como slice, útil para UIs.
func IDEOptions() []IDE {
	return []IDE{IDENone, IDEVSCode, IDECLion, IDENvim, IDEZed}
}

// Label retorna o nome amigável da IDE para exibição.
func (i IDE) Label() string {
	switch i {
	case IDEVSCode:
		return "Visual Studio Code"
	case IDECLion:
		return "CLion"
	case IDENvim:
		return "Neovim"
	case IDEZed:
		return "Zed"
	default:
		return "Nenhuma"
	}
}

// ─────────────────────────────────────────────────────────────────────────────

// PackageManager representa o gerenciador de pacotes C++ a ser configurado.
type PackageManager string

const (
	PkgNone         PackageManager = "none"         // Sem gerenciador de pacotes
	PkgVCPKG        PackageManager = "vcpkg"        // VCPKG (vcpkg.json manifest)
	PkgFetchContent PackageManager = "fetchcontent" // CMake FetchContent
)

// PackageManagerOptions retorna todos os gerenciadores disponíveis.
func PackageManagerOptions() []PackageManager {
	return []PackageManager{PkgNone, PkgVCPKG, PkgFetchContent}
}

// Label retorna o nome amigável do gerenciador para exibição.
func (p PackageManager) Label() string {
	switch p {
	case PkgVCPKG:
		return "VCPKG (vcpkg.json)"
	case PkgFetchContent:
		return "FetchContent (CMake nativo)"
	default:
		return "Nenhum"
	}
}

// ─────────────────────────────────────────────────────────────────────────────

// ProjectType representa o tipo de artefato que o projeto C++ irá gerar.
type ProjectType string

const (
	TypeExecutable ProjectType = "executable"  // Binário executável
	TypeStaticLib  ProjectType = "static-lib"  // Biblioteca estática (.a / .lib)
	TypeHeaderOnly ProjectType = "header-only" // Biblioteca header-only (interface)
)

// ProjectTypeOptions retorna todos os tipos de projeto disponíveis.
func ProjectTypeOptions() []ProjectType {
	return []ProjectType{TypeExecutable, TypeStaticLib, TypeHeaderOnly}
}

// Label retorna o nome amigável do tipo de projeto.
func (t ProjectType) Label() string {
	switch t {
	case TypeExecutable:
		return "Executável"
	case TypeStaticLib:
		return "Biblioteca Estática"
	case TypeHeaderOnly:
		return "Biblioteca Header-Only"
	default:
		return string(t)
	}
}

// IsLibrary retorna true se o projeto é qualquer tipo de biblioteca.
func (t ProjectType) IsLibrary() bool {
	return t == TypeStaticLib || t == TypeHeaderOnly
}

// ─────────────────────────────────────────────────────────────────────────────

// CppStandard representa a versão do padrão ISO C++ a ser utilizada.
type CppStandard string

const (
	Cpp17 CppStandard = "17" // ISO C++17
	Cpp20 CppStandard = "20" // ISO C++20
	Cpp23 CppStandard = "23" // ISO C++23
)

// CppStandardOptions retorna todos os padrões C++ suportados.
func CppStandardOptions() []CppStandard {
	return []CppStandard{Cpp17, Cpp20, Cpp23}
}

// Label retorna o nome amigável do padrão C++ para exibição.
func (s CppStandard) Label() string {
	return "C++" + string(s)
}

// ─────────────────────────────────────────────────────────────────────────────
// Estrutura principal de configuração
// ─────────────────────────────────────────────────────────────────────────────

// ProjectConfig contém todas as informações necessárias para gerar
// um projeto C++ completo. É preenchida via TUI interativa ou flags CLI.
type ProjectConfig struct {
	// ── Metadados do projeto ──────────────────────────────────────────────────

	// Name é o nome do projeto (ex: "meu-projeto").
	// Usado para nomear pastas, targets CMake e variáveis.
	Name string

	// Description é uma breve descrição do projeto.
	Description string

	// Author é o nome do autor ou organização.
	Author string

	// Version é a versão inicial do projeto no formato SemVer (ex: "1.0.0").
	Version string

	// ── Configurações técnicas ────────────────────────────────────────────────

	// Standard define o padrão C++ a ser utilizado (17, 20, 23).
	Standard CppStandard

	// ProjectType define o tipo de artefato: executável, biblioteca, etc.
	ProjectType ProjectType

	// ── Ferramentas e integrações ─────────────────────────────────────────────

	// PackageManager define o sistema de gerenciamento de dependências C++.
	PackageManager PackageManager

	// IDE define a IDE alvo para geração de configurações específicas.
	IDE IDE

	// ── Flags de funcionalidades opcionais ────────────────────────────────────

	// UseGit indica se um repositório Git deve ser inicializado no projeto.
	UseGit bool

	// UseClangd indica se o arquivo de configuração .clangd deve ser gerado.
	UseClangd bool

	// UseClangFormat indica se o arquivo .clang-format deve ser gerado.
	UseClangFormat bool

	// ── Configuração de saída ─────────────────────────────────────────────────

	// OutputDir é o diretório base onde o projeto será criado.
	// O projeto ficará em OutputDir/Name.
	OutputDir string
}

// ─────────────────────────────────────────────────────────────────────────────
// Métodos auxiliares
// ─────────────────────────────────────────────────────────────────────────────

// ProjectPath retorna o caminho completo do diretório raiz do projeto gerado,
// combinando OutputDir e Name de forma segura para o sistema operacional.
func (c *ProjectConfig) ProjectPath() string {
	if c.OutputDir == "" || c.OutputDir == "." {
		return c.Name
	}
	return filepath.Join(c.OutputDir, c.Name)
}

// Validate verifica se a configuração contém os campos obrigatórios preenchidos
// e retorna uma lista de erros encontrados.
func (c *ProjectConfig) Validate() []string {
	var errs []string

	if c.Name == "" {
		errs = append(errs, "nome do projeto é obrigatório")
	}
	if c.Version == "" {
		errs = append(errs, "versão do projeto é obrigatória")
	}
	if c.Standard == "" {
		errs = append(errs, "padrão C++ é obrigatório")
	}
	if c.ProjectType == "" {
		errs = append(errs, "tipo de projeto é obrigatório")
	}

	return errs
}

// Default retorna uma ProjectConfig com valores padrão sensatos,
// útil como ponto de partida antes de aplicar as escolhas do usuário.
func Default() *ProjectConfig {
	return &ProjectConfig{
		Version:        "1.0.0",
		Standard:       Cpp20,
		ProjectType:    TypeExecutable,
		PackageManager: PkgNone,
		IDE:            IDENone,
		UseGit:         true,
		UseClangd:      true,
		UseClangFormat: true,
		OutputDir:      ".",
	}
}
