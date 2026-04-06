// Package tui contains all terminal user interface components
// used by cpp-gen, including interactive forms and visual styles.
package tui

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/matpdev/cpp-gen/internal/config"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// ─────────────────────────────────────────────────────────────────────────────
// RunForm
// ─────────────────────────────────────────────────────────────────────────────

// RunForm displays the interactive TUI form and returns the ProjectConfig
// filled in by the user. If the user cancels (Esc / Ctrl+C),
// it returns an error with the message "user aborted".
//
// initialName is pre-filled in the name field when provided via positional
// argument on the command line (e.g. cpp-gen new my-project).
func RunForm(initialName string) (*config.ProjectConfig, error) {
	cfg := config.Default()
	cfg.Name = initialName

	// Intermediate string variables for selection fields,
	// since huh.NewSelect requires *string while config uses custom types.
	var (
		standard         = string(cfg.Standard)
		projectType      = string(cfg.ProjectType)
		layout           = string(cfg.Layout)
		pkgManager       = string(cfg.PackageManager)
		ide              = string(cfg.IDE)
		clangFormatStyle = string(cfg.ClangFormatStyle)
	)

	// ── Group 1: Project Identity ─────────────────────────────────────────────
	groupIdentity := huh.NewGroup(
		huh.NewNote().
			Title("⚡ cpp-gen").
			Description("Gerador moderno de projetos C++\nPreencha as informações abaixo para criar seu projeto."),

		huh.NewInput().
			Title("Nome do projeto").
			Description("Usado para nomear diretórios, targets CMake e variáveis.").
			Placeholder("meu-projeto").
			Value(&cfg.Name).
			Validate(validateProjectName),

		huh.NewInput().
			Title("Descrição").
			Description("Uma breve descrição do que o projeto faz.").
			Placeholder("Um projeto C++ moderno").
			Value(&cfg.Description),

		huh.NewInput().
			Title("Autor").
			Description("Seu nome ou o nome da organização.").
			Placeholder("Seu Nome").
			Value(&cfg.Author),

		huh.NewInput().
			Title("Versão inicial").
			Description("Versão no formato SemVer (MAJOR.MINOR.PATCH).").
			Placeholder("1.0.0").
			Value(&cfg.Version).
			Validate(validateVersion),
	)

	// ── Group 2: C++ Technical Settings ──────────────────────────────────────
	groupTechnical := huh.NewGroup(
		huh.NewNote().
			Title("Configurações C++").
			Description("Defina o padrão da linguagem e o tipo de artefato gerado pelo CMake."),

		huh.NewSelect[string]().
			Title("Padrão C++").
			Description("ISO C++ standard a ser configurado no CMake.").
			Options(
				huh.NewOption("C++17  — Padrão amplamente suportado", string(config.Cpp17)),
				huh.NewOption("C++20  — Conceitos, corrotinas, ranges (recomendado)", string(config.Cpp20)),
				huh.NewOption("C++23  — Padrão mais recente (suporte variável)", string(config.Cpp23)),
			).
			Value(&standard),

		huh.NewSelect[string]().
			Title("Tipo de projeto").
			Description("Define o artefato final gerado pelo CMake.").
			Options(
				huh.NewOption("Executável       — add_executable()", string(config.TypeExecutable)),
				huh.NewOption("Biblioteca Est.  — add_library(STATIC)", string(config.TypeStaticLib)),
				huh.NewOption("Header-Only      — add_library(INTERFACE)", string(config.TypeHeaderOnly)),
			).
			Value(&projectType),
	)

	// ── Group 3: Folder Layout ────────────────────────────────────────────────
	groupLayout := huh.NewGroup(
		huh.NewNote().
			Title("Layout de Pastas").
			Description("Escolha como os arquivos do projeto serão organizados.\nSelecione uma opção para ver a estrutura de diretórios correspondente."),

		huh.NewSelect[string]().
			Title("Estrutura de diretórios").
			DescriptionFunc(func() string {
				return config.FolderLayout(layout).TreePreview()
			}, &layout).
			Options(
				huh.NewOption(
					"Separate  — include/<nome>/ + src/  (clássico CMake)",
					string(config.LayoutSeparate),
				),
				huh.NewOption(
					"Merged    — <nome>/  headers e .cpp juntos  (Pitchfork)",
					string(config.LayoutMerged),
				),
				huh.NewOption(
					"Flat      — src/  tudo junto, sem separação",
					string(config.LayoutFlat),
				),
				huh.NewOption(
					"Modular   — libs/<nome>/  multi-módulo  (Pitchfork libs/)",
					string(config.LayoutModular),
				),
				huh.NewOption(
					"Two-Root  — include/ + src/  sem namespace subdir",
					string(config.LayoutTwoRoot),
				),
			).
			Value(&layout),
	)

	// ── Group 4: Package Manager ──────────────────────────────────────────────
	groupPackages := huh.NewGroup(
		huh.NewNote().
			Title("Gerenciador de Pacotes").
			Description("Escolha como as dependências C++ serão gerenciadas."),

		huh.NewSelect[string]().
			Title("Gerenciador de pacotes").
			Description("Configura a integração no CMakeLists.txt e arquivos auxiliares.").
			Options(
				huh.NewOption("Nenhum           — Gerenciar manualmente", string(config.PkgNone)),
				huh.NewOption("VCPKG            — vcpkg.json manifest mode", string(config.PkgVCPKG)),
				huh.NewOption("FetchContent     — CMake FetchContent nativo", string(config.PkgFetchContent)),
			).
			Value(&pkgManager),
	)

	// ── Group 5: Development Environment ─────────────────────────────────────
	groupIDE := huh.NewGroup(
		huh.NewNote().
			Title("IDE e Ferramentas").
			Description("Configure o ambiente de desenvolvimento e as ferramentas de análise."),

		huh.NewSelect[string]().
			Title("IDE alvo").
			Description("Gera tasks, launch configs e settings para a IDE escolhida.").
			Options(
				huh.NewOption("Nenhuma          — Apenas CMake", string(config.IDENone)),
				huh.NewOption("VSCode           — tasks.json, launch.json, settings.json", string(config.IDEVSCode)),
				huh.NewOption("CLion            — CMakePresets.json otimizado", string(config.IDECLion)),
				huh.NewOption("Neovim           — .nvim.lua + configuração LSP", string(config.IDENvim)),
				huh.NewOption("Zed              — .zed/settings.json + .zed/tasks.json", string(config.IDEZed)),
			).
			Value(&ide),

		huh.NewConfirm().
			Title("Inicializar repositório Git?").
			Description("Cria .git/, .gitignore e commit inicial.").
			Affirmative("Sim").
			Negative("Não").
			Value(&cfg.UseGit),

		huh.NewConfirm().
			Title("Adicionar configuração Clangd?").
			Description("Gera .clangd apontando para compile_commands.json.").
			Affirmative("Sim").
			Negative("Não").
			Value(&cfg.UseClangd),

		huh.NewConfirm().
			Title("Adicionar Clang-Format?").
			Description("Gera .clang-format para formatação automática do código.").
			Affirmative("Sim").
			Negative("Não").
			Value(&cfg.UseClangFormat),

		huh.NewSelect[string]().
			Title("Estilo do Clang-Format").
			DescriptionFunc(func() string {
				return config.ClangFormatStyle(clangFormatStyle).Description()
			}, &clangFormatStyle).
			Options(
				huh.NewOption("LLVM        — personalizado  (4 espaços, Allman, 100 cols)", string(config.ClangFormatLLVM)),
				huh.NewOption("Google      — Google C++ Style Guide  (2 espaços, 80 cols)", string(config.ClangFormatGoogle)),
				huh.NewOption("Chromium    — baseado em Google  (Chromium project)", string(config.ClangFormatChromium)),
				huh.NewOption("Mozilla     — Mozilla Coding Style", string(config.ClangFormatMozilla)),
				huh.NewOption("WebKit      — WebKit Coding Style  (4 espaços)", string(config.ClangFormatWebKit)),
				huh.NewOption("Microsoft   — Microsoft C++ Style", string(config.ClangFormatMicrosoft)),
				huh.NewOption("GNU         — GNU Coding Standards", string(config.ClangFormatGNU)),
			).
			Value(&clangFormatStyle),
	)

	// ── Form construction and execution ──────────────────────────────────────
	form := huh.NewForm(
		groupIdentity,
		groupTechnical,
		groupLayout,
		groupPackages,
		groupIDE,
	).
		WithTheme(buildTheme()).
		WithWidth(72)

	if err := form.Run(); err != nil {
		// huh returns this error when the user presses Esc or Ctrl+C
		if errors.Is(err, huh.ErrUserAborted) {
			return nil, errors.New("user aborted")
		}
		return nil, fmt.Errorf("erro no formulário: %w", err)
	}

	// Converts string variables back to custom types
	cfg.Standard = config.CppStandard(standard)
	cfg.ProjectType = config.ProjectType(projectType)
	cfg.Layout = config.FolderLayout(layout)
	cfg.PackageManager = config.PackageManager(pkgManager)
	cfg.IDE = config.IDE(ide)
	cfg.ClangFormatStyle = config.ClangFormatStyle(clangFormatStyle)

	return cfg, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Field validators
// ─────────────────────────────────────────────────────────────────────────────

// reProjectName defines the valid characters for project names:
// lowercase letters, digits and hyphens, with alphanumeric start and end.
var reProjectName = regexp.MustCompile(`^[a-z0-9][a-z0-9\-]*[a-z0-9]$|^[a-z0-9]$`)

// validateProjectName validates the project name entered by the user.
// Rules:
//   - Cannot be empty
//   - Must have at least 2 characters
//   - Only lowercase letters, digits and hyphens
//   - Cannot start or end with a hyphen
//   - Cannot contain spaces or special characters
func validateProjectName(name string) error {
	name = strings.TrimSpace(name)

	if name == "" {
		return errors.New("o nome do projeto não pode ser vazio")
	}
	if len(name) < 2 {
		return errors.New("o nome deve ter pelo menos 2 caracteres")
	}
	if len(name) > 64 {
		return errors.New("o nome deve ter no máximo 64 caracteres")
	}

	// Verifica se há letras maiúsculas (fornece dica útil ao usuário)
	for _, r := range name {
		if unicode.IsUpper(r) {
			return fmt.Errorf("use letras minúsculas (sugestão: %q)", strings.ToLower(name))
		}
	}

	if !reProjectName.MatchString(name) {
		return errors.New("use apenas letras minúsculas, números e hífens (ex: meu-projeto)")
	}

	return nil
}

// reVersion validates the basic SemVer format: MAJOR.MINOR.PATCH (all numeric).
var reVersion = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// validateVersion validates the project version field.
// Accepts the MAJOR.MINOR.PATCH format (e.g. "1.0.0", "0.3.12").
func validateVersion(v string) error {
	v = strings.TrimSpace(v)
	if v == "" {
		return errors.New("a versão não pode ser vazia")
	}
	if !reVersion.MatchString(v) {
		return errors.New("use o formato SemVer: MAJOR.MINOR.PATCH (ex: 1.0.0)")
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Custom theme
// ─────────────────────────────────────────────────────────────────────────────

// buildTheme creates a custom huh theme based on the cpp-gen color palette.
// Keeps the visual consistent with the lipgloss styles defined in styles.go.
func buildTheme() *huh.Theme {
	theme := huh.ThemeCharm()

	// Group header / note
	theme.Focused.Title = TitleStyle.Copy()
	theme.Focused.Description = MutedStyle.Copy()

	// Selected / active field
	theme.Focused.SelectedOption = lipglossColor(colorAccent)
	theme.Focused.UnselectedOption = MutedStyle.Copy()

	// Selection cursor
	theme.Focused.SelectSelector = InfoStyle.Copy()

	// Confirmation buttons
	theme.Focused.FocusedButton = lipglossColor(colorSuccess)
	theme.Focused.BlurredButton = MutedStyle.Copy()

	return theme
}

// lipglossColor creates a simple lipgloss style with only the foreground color,
// compatible with the type expected by huh theme fields.
func lipglossColor(c lipglossColorType) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(c)
}

// lipglossColorType is an internal alias for lipgloss.Color, making the
// lipglossColor signature more explicit and avoiding circular imports in tests.
type lipglossColorType = lipgloss.Color
