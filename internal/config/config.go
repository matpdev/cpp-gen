// Package config define todos os tipos, constantes e estruturas de configuração
// utilizados pelo cpp-gen para descrever um projeto C++ a ser gerado.
package config

import "path/filepath"

// ─────────────────────────────────────────────────────────────────────────────
// FolderLayout
// ─────────────────────────────────────────────────────────────────────────────

// FolderLayout representa o padrão de organização de pastas do projeto C++.
// Cada layout define onde ficam os headers públicos, arquivos de implementação
// e como os targets CMake referenciam os diretórios de include.
type FolderLayout string

const (
	// LayoutSeparate é o padrão clássico com separação include/<nome>/ e src/.
	// Headers públicos ficam em include/<nome>/*.hpp e implementações em src/*.cpp.
	// É o padrão mais comum em projetos CMake modernos e facilita distinguir
	// a API pública da implementação privada.
	//
	//  include/<nome>/  ← headers públicos (API)
	//  src/             ← implementações (.cpp) e headers privados
	//  tests/           ← testes
	LayoutSeparate FolderLayout = "separate"

	// LayoutMerged é o padrão Pitchfork (P1204R0 / vector-of-bool).
	// Headers e implementações ficam juntos no diretório <nome>/, eliminando
	// a navegação entre include/ e src/. Unit tests ficam como arquivos irmãos.
	// Include canônico: #include <<nome>/file.hpp>
	//
	//  <nome>/          ← headers (.hpp) e fontes (.cpp) juntos
	//  <nome>/*.test.cpp ← unit tests como arquivos irmãos (opcional)
	//  tests/           ← testes de integração/funcionais
	LayoutMerged FolderLayout = "merged"

	// LayoutFlat coloca tudo dentro de src/ sem separação entre headers e fontes.
	// É o padrão mais simples, indicado para executáveis e projetos pequenos onde
	// a distinção entre API pública e privada não é relevante.
	//
	//  src/             ← headers (.hpp) e fontes (.cpp) no mesmo lugar
	//  tests/           ← testes
	LayoutFlat FolderLayout = "flat"

	// LayoutModular segue o padrão Pitchfork para projetos multi-módulo usando libs/.
	// Cada módulo/biblioteca vive em libs/<nome>/{include,src}. Executáveis ficam
	// em apps/. Facilita a extração de módulos em projetos separados futuramente.
	//
	//  libs/<nome>/include/<nome>/  ← headers públicos
	//  libs/<nome>/src/             ← implementações
	//  apps/                        ← executáveis (se aplicável)
	//  tests/                       ← testes
	LayoutModular FolderLayout = "modular"

	// LayoutTwoRoot é um split simples include/ + src/ sem subdiretório de namespace.
	// Difere do Separate por não criar include/<nome>/: headers ficam diretamente
	// em include/*.hpp. Comum em projetos menores que não precisam do prefixo de
	// namespace no #include.
	//
	//  include/   ← headers públicos (sem subdir de namespace)
	//  src/       ← implementações
	//  tests/     ← testes
	LayoutTwoRoot FolderLayout = "two-root"
)

// FolderLayoutOptions retorna todos os layouts disponíveis em ordem sugerida.
func FolderLayoutOptions() []FolderLayout {
	return []FolderLayout{
		LayoutSeparate,
		LayoutMerged,
		LayoutFlat,
		LayoutModular,
		LayoutTwoRoot,
	}
}

// Label retorna o nome amigável do layout para exibição na UI.
func (l FolderLayout) Label() string {
	switch l {
	case LayoutSeparate:
		return "Separate   — include/<nome>/ + src/"
	case LayoutMerged:
		return "Merged     — <nome>/ (Pitchfork / P1204R0)"
	case LayoutFlat:
		return "Flat       — src/ com headers e fontes juntos"
	case LayoutModular:
		return "Modular    — libs/<nome>/ (Pitchfork multi-módulo)"
	case LayoutTwoRoot:
		return "Two-Root   — include/ + src/ (sem namespace subdir)"
	default:
		return string(l)
	}
}

// Description retorna uma explicação detalhada do layout, exibida como hint
// abaixo do campo de seleção na TUI.
func (l FolderLayout) Description() string {
	switch l {
	case LayoutSeparate:
		return "Clássico CMake. API em include/<nome>/, impl. em src/. Mais usado em libs."
	case LayoutMerged:
		return "Pitchfork (P1204R0). Headers e .cpp juntos em <nome>/. Navegação mais fácil."
	case LayoutFlat:
		return "Tudo em src/. Ideal para executáveis e projetos simples."
	case LayoutModular:
		return "Multi-módulo. Cada lib em libs/<nome>/. Pronto para escalar."
	case LayoutTwoRoot:
		return "Include/ raiz sem namespace. Simples para bibliotecas pequenas."
	default:
		return ""
	}
}

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

	// Layout define o padrão de organização de pastas e arquivos do projeto.
	// Determina onde ficam headers, implementações e como o CMake os referencia.
	Layout FolderLayout

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
		Layout:         LayoutSeparate,
		PackageManager: PkgNone,
		IDE:            IDENone,
		UseGit:         true,
		UseClangd:      true,
		UseClangFormat: true,
		OutputDir:      ".",
	}
}
