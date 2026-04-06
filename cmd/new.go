// Package cmd contains all CLI commands for cpp-gen.
package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/matpdev/cpp-gen/internal/config"
	"github.com/matpdev/cpp-gen/internal/generator"
	"github.com/matpdev/cpp-gen/internal/tui"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// ─────────────────────────────────────────────────────────────────────────────
// Command definition
// ─────────────────────────────────────────────────────────────────────────────

// newCmd is the main subcommand of cpp-gen. It guides the user through the
// process of creating a new C++ project via an interactive TUI form or,
// optionally, via flags for use in scripts and automations.
//
// Usage:
//
//	cpp-gen new                        # fully interactive
//	cpp-gen new meu-projeto            # name pre-filled, rest interactive
//	cpp-gen new --no-interactive [flags] # non-interactive mode (scripting)
var newCmd = &cobra.Command{
	Use:   "new [nome-do-projeto]",
	Short: "Cria um novo projeto C++ com CMake e ferramentas configuradas",
	Long:  renderNewLong(),

	// Accepts 0 or 1 positional argument (project name).
	Args: cobra.MaximumNArgs(1),

	// RunE is preferred over Run as it allows returning errors for centralized
	// handling in main(), avoiding manual calls to os.Exit().
	RunE: runNew,

	// Examples shown in the subcommand --help.
	Example: `  # Modo interativo (recomendado)
  cpp-gen new
  cpp-gen new meu-projeto

  # Modo não-interativo para scripts e CI
  cpp-gen new meu-projeto \
    --no-interactive \
    --description "Meu projeto C++" \
    --author "Fulano" \
    --std 20 \
    --type executable \
    --layout merged \
    --pkg vcpkg \
    --ide vscode`,
}

// ─────────────────────────────────────────────────────────────────────────────
// Flag registration
// ─────────────────────────────────────────────────────────────────────────────

func init() {
	// ── Control flags ─────────────────────────────────────────────────────────

	newCmd.Flags().StringP(
		"output", "o", ".",
		"Diretório de saída onde a pasta do projeto será criada",
	)

	newCmd.Flags().BoolP(
		"no-interactive", "n", false,
		"Desativa o formulário TUI; usa apenas as flags fornecidas",
	)

	// ── Metadata flags (non-interactive mode) ─────────────────────────────────

	newCmd.Flags().String(
		"name", "",
		"Nome do projeto (alternativa ao argumento posicional)",
	)
	newCmd.Flags().String(
		"description", "",
		"Descrição breve do projeto",
	)
	newCmd.Flags().String(
		"author", "",
		"Nome do autor ou organização",
	)
	newCmd.Flags().String(
		"version", "1.0.0",
		"Versão inicial do projeto no formato SemVer (ex: 1.0.0)",
	)

	// ── Technical configuration flags ─────────────────────────────────────────

	newCmd.Flags().String(
		"std", "20",
		"Padrão C++ a usar: 17 | 20 | 23",
	)
	newCmd.Flags().String(
		"type", "executable",
		"Tipo do projeto: executable | static-lib | header-only",
	)
	newCmd.Flags().String(
		"pkg", "none",
		"Gerenciador de pacotes: none | vcpkg | fetchcontent",
	)
	newCmd.Flags().String(
		"ide", "none",
		"IDE alvo: none | vscode | clion | nvim | zed",
	)
	newCmd.Flags().String(
		"layout", "separate",
		"Layout de pastas: separate | merged | flat | modular | two-root",
	)

	// ── Optional feature flags ────────────────────────────────────────────────

	newCmd.Flags().Bool(
		"no-git", false,
		"Não inicializar repositório Git",
	)
	newCmd.Flags().Bool(
		"no-clangd", false,
		"Não gerar arquivo .clangd",
	)
	newCmd.Flags().Bool(
		"no-clang-format", false,
		"Não gerar arquivo .clang-format",
	)
	newCmd.Flags().String(
		"clang-format-style", "llvm",
		"Estilo base do .clang-format: llvm, google, chromium, mozilla, webkit, microsoft, gnu",
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// Main handler
// ─────────────────────────────────────────────────────────────────────────────

// runNew is the handler for the `new` command. It decides between interactive
// and non-interactive mode, collects the configuration, runs the generator
// and displays the summary.
func runNew(cmd *cobra.Command, args []string) error {
	outputDir, _ := cmd.Flags().GetString("output")
	noInteractive, _ := cmd.Flags().GetBool("no-interactive")

	// Project name: can come as a positional argument or via the --name flag.
	// The positional argument takes precedence over the flag.
	initialName := ""
	if len(args) > 0 {
		initialName = args[0]
	} else if flagName, _ := cmd.Flags().GetString("name"); flagName != "" {
		initialName = flagName
	}

	// ── Configuration collection ──────────────────────────────────────────────

	var cfg *config.ProjectConfig
	var err error

	if noInteractive {
		cfg, err = buildConfigFromFlags(cmd, initialName)
		if err != nil {
			return fmt.Errorf("configuração inválida: %w", err)
		}
	} else {
		// Interactive mode: opens the TUI form with the name pre-filled.
		cfg, err = tui.RunForm(initialName)
		if err != nil {
			// User cancellation is not an error — just exits silently.
			if err.Error() == "user aborted" {
				fmt.Println(tui.MutedStyle.Render("\nOperação cancelada."))
				return nil
			}
			return fmt.Errorf("erro no formulário: %w", err)
		}
	}

	// Applies the output directory defined via the --output flag.
	cfg.OutputDir = outputDir

	// ── Final validation ──────────────────────────────────────────────────────

	if errs := cfg.Validate(); len(errs) > 0 {
		return fmt.Errorf("configuração incompleta:\n  • %s", strings.Join(errs, "\n  • "))
	}

	// ── Display summary before generating ────────────────────────────────────

	printProjectSummary(cfg)

	// ── Generator execution ───────────────────────────────────────────────────

	gen := generator.New(cfg, isVerbose(cmd))

	if err := gen.Generate(); err != nil {
		return fmt.Errorf("falha ao gerar o projeto: %w", err)
	}

	// ── Completion message ────────────────────────────────────────────────────

	printSuccessMessage(cfg)

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Non-interactive mode
// ─────────────────────────────────────────────────────────────────────────────

// buildConfigFromFlags builds a ProjectConfig exclusively from command-line
// flags, without opening any TUI form.
//
// Used when --no-interactive is passed, ideal for CI/CD pipelines
// and automation scripts.
func buildConfigFromFlags(cmd *cobra.Command, initialName string) (*config.ProjectConfig, error) {
	cfg := config.Default()

	// ── Name ──────────────────────────────────────────────────────────────────
	if initialName != "" {
		cfg.Name = initialName
	}

	// ── Metadata ──────────────────────────────────────────────────────────────
	if v, _ := cmd.Flags().GetString("description"); v != "" {
		cfg.Description = v
	}
	if v, _ := cmd.Flags().GetString("author"); v != "" {
		cfg.Author = v
	}
	if v, _ := cmd.Flags().GetString("version"); v != "" {
		cfg.Version = v
	}

	// ── Folder layout ─────────────────────────────────────────────────────────
	if layoutStr, _ := cmd.Flags().GetString("layout"); layoutStr != "" {
		l, err := parseLayout(layoutStr)
		if err != nil {
			return nil, err
		}
		cfg.Layout = l
	}

	// ── C++ standard ──────────────────────────────────────────────────────────
	if stdStr, _ := cmd.Flags().GetString("std"); stdStr != "" {
		std, err := parseCppStandard(stdStr)
		if err != nil {
			return nil, err
		}
		cfg.Standard = std
	}

	// ── Project type ──────────────────────────────────────────────────────────
	if typeStr, _ := cmd.Flags().GetString("type"); typeStr != "" {
		pt, err := parseProjectType(typeStr)
		if err != nil {
			return nil, err
		}
		cfg.ProjectType = pt
	}

	// ── Package manager ───────────────────────────────────────────────────────
	if pkgStr, _ := cmd.Flags().GetString("pkg"); pkgStr != "" {
		pm, err := parsePackageManager(pkgStr)
		if err != nil {
			return nil, err
		}
		cfg.PackageManager = pm
	}

	// ── IDE ───────────────────────────────────────────────────────────────────
	if ideStr, _ := cmd.Flags().GetString("ide"); ideStr != "" {
		ide, err := parseIDE(ideStr)
		if err != nil {
			return nil, err
		}
		cfg.IDE = ide
	}

	// ── Opt-out flags ─────────────────────────────────────────────────────────
	noGit, _ := cmd.Flags().GetBool("no-git")
	noClangd, _ := cmd.Flags().GetBool("no-clangd")
	noClangFmt, _ := cmd.Flags().GetBool("no-clang-format")

	cfg.UseGit = !noGit
	cfg.UseClangd = !noClangd
	cfg.UseClangFormat = !noClangFmt

	// ── Clang-Format style ────────────────────────────────────────────────────
	if styleStr, _ := cmd.Flags().GetString("clang-format-style"); styleStr != "" {
		style, err := parseClangFormatStyle(styleStr)
		if err != nil {
			return nil, err
		}
		cfg.ClangFormatStyle = style
	}

	// ── Required validation in non-interactive mode ───────────────────────────
	if cfg.Name == "" {
		return nil, errors.New(
			"nome do projeto é obrigatório; use o argumento posicional ou --name",
		)
	}

	return cfg, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Enum parsers (flags → config types)
// ─────────────────────────────────────────────────────────────────────────────

// parseCppStandard converts a string (e.g. "20") to config.CppStandard.
// Returns an error if the value is not recognized.
func parseCppStandard(s string) (config.CppStandard, error) {
	switch s {
	case "17":
		return config.Cpp17, nil
	case "20":
		return config.Cpp20, nil
	case "23":
		return config.Cpp23, nil
	default:
		return "", fmt.Errorf("padrão C++ inválido %q; valores aceitos: 17, 20, 23", s)
	}
}

// parseProjectType converts a string (e.g. "executable") to config.ProjectType.
func parseProjectType(s string) (config.ProjectType, error) {
	switch strings.ToLower(s) {
	case "executable":
		return config.TypeExecutable, nil
	case "static-lib", "staticlib", "static_lib":
		return config.TypeStaticLib, nil
	case "header-only", "headeronly", "header_only":
		return config.TypeHeaderOnly, nil
	default:
		return "", fmt.Errorf(
			"tipo de projeto inválido %q; valores aceitos: executable, static-lib, header-only", s,
		)
	}
}

// parsePackageManager converts a string (e.g. "vcpkg") to config.PackageManager.
func parsePackageManager(s string) (config.PackageManager, error) {
	switch strings.ToLower(s) {
	case "none", "":
		return config.PkgNone, nil
	case "vcpkg":
		return config.PkgVCPKG, nil
	case "fetchcontent", "fetch-content", "fetch_content":
		return config.PkgFetchContent, nil
	default:
		return "", fmt.Errorf(
			"gerenciador de pacotes inválido %q; valores aceitos: none, vcpkg, fetchcontent", s,
		)
	}
}

// parseIDE converts a string (e.g. "vscode") to config.IDE.
func parseIDE(s string) (config.IDE, error) {
	switch strings.ToLower(s) {
	case "none", "":
		return config.IDENone, nil
	case "vscode", "vs-code", "vs_code":
		return config.IDEVSCode, nil
	case "clion":
		return config.IDECLion, nil
	case "nvim", "neovim":
		return config.IDENvim, nil
	case "zed":
		return config.IDEZed, nil
	default:
		return "", fmt.Errorf(
			"IDE inválida %q; valores aceitos: none, vscode, clion, nvim, zed", s,
		)
	}
}

// parseLayout converts a string (e.g. "merged") to config.FolderLayout.
// Accepts common variations with hyphens, underscores and no separator.
func parseLayout(s string) (config.FolderLayout, error) {
	switch strings.ToLower(s) {
	case "separate", "sep":
		return config.LayoutSeparate, nil
	case "merged", "merge", "pitchfork":
		return config.LayoutMerged, nil
	case "flat":
		return config.LayoutFlat, nil
	case "modular", "mod", "libs":
		return config.LayoutModular, nil
	case "two-root", "tworoot", "two_root", "split":
		return config.LayoutTwoRoot, nil
	default:
		return "", fmt.Errorf(
			"layout inválido %q; valores aceitos: separate, merged, flat, modular, two-root", s,
		)
	}
}

// parseClangFormatStyle converts a string (e.g. "google") to config.ClangFormatStyle.
func parseClangFormatStyle(s string) (config.ClangFormatStyle, error) {
	switch strings.ToLower(s) {
	case "llvm":
		return config.ClangFormatLLVM, nil
	case "google":
		return config.ClangFormatGoogle, nil
	case "chromium":
		return config.ClangFormatChromium, nil
	case "mozilla":
		return config.ClangFormatMozilla, nil
	case "webkit":
		return config.ClangFormatWebKit, nil
	case "microsoft":
		return config.ClangFormatMicrosoft, nil
	case "gnu":
		return config.ClangFormatGNU, nil
	default:
		return "", fmt.Errorf(
			"estilo clang-format inválido %q; valores aceitos: llvm, google, chromium, mozilla, webkit, microsoft, gnu",
			s,
		)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Formatted output
// ─────────────────────────────────────────────────────────────────────────────

// printProjectSummary displays a formatted summary of the project configuration
// before generation starts, allowing the user to review their choices.
func printProjectSummary(cfg *config.ProjectConfig) {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED")).
		MarginTop(1)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#374151")).
		Padding(0, 2).
		MarginTop(1).
		MarginBottom(1)

	fmt.Println(headerStyle.Render("Resumo do Projeto"))

	rows := []string{
		tui.FormatKeyValue("Nome", cfg.Name),
		tui.FormatKeyValue("Descrição", orDash(cfg.Description)),
		tui.FormatKeyValue("Autor", orDash(cfg.Author)),
		tui.FormatKeyValue("Versão", cfg.Version),
		tui.FormatKeyValue("Padrão C++", cfg.Standard.Label()),
		tui.FormatKeyValue("Tipo", cfg.ProjectType.Label()),
		tui.FormatKeyValue("Layout", cfg.Layout.Label()),
		tui.FormatKeyValue("Pacotes", cfg.PackageManager.Label()),
		tui.FormatKeyValue("IDE", cfg.IDE.Label()),
		tui.FormatKeyValue("Git", boolLabel(cfg.UseGit)),
		tui.FormatKeyValue("Clangd", boolLabel(cfg.UseClangd)),
		tui.FormatKeyValue("Clang-Format", func() string {
			if !cfg.UseClangFormat {
				return "Não"
			}
			return "Sim  (" + string(cfg.ClangFormatStyle) + ")"
		}()),
		tui.FormatKeyValue("Destino", cfg.ProjectPath()),
	}

	fmt.Println(boxStyle.Render(strings.Join(rows, "\n")))
}

// printSuccessMessage displays the final completion message with next-steps
// instructions for the user.
func printSuccessMessage(cfg *config.ProjectConfig) {
	successStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#10B981")).
		MarginTop(1)

	codeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#06B6D4"))

	mutedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280"))

	fmt.Println(successStyle.Render("✓ Projeto criado com sucesso!"))
	fmt.Println()
	fmt.Println(mutedStyle.Render("Próximos passos:"))

	steps := buildNextSteps(cfg)
	for i, step := range steps {
		fmt.Printf("  %s  %s\n",
			mutedStyle.Render(fmt.Sprintf("%d.", i+1)),
			codeStyle.Render(step),
		)
	}
	fmt.Println()
}

// buildNextSteps returns a list of suggested commands for the user
// to run after project generation.
func buildNextSteps(cfg *config.ProjectConfig) []string {
	projectPath := cfg.ProjectPath()
	steps := []string{
		fmt.Sprintf("cd %s", projectPath),
	}

	// CMake configure instruction (varies by package manager)
	switch cfg.PackageManager {
	case config.PkgVCPKG:
		steps = append(steps,
			"# Certifique-se que VCPKG_ROOT está definido no ambiente",
			"cmake --preset debug",
		)
	default:
		steps = append(steps, "cmake -B build -DCMAKE_BUILD_TYPE=Debug")
	}

	steps = append(steps, "cmake --build build")

	if cfg.ProjectType == config.TypeExecutable {
		steps = append(steps, fmt.Sprintf("./build/%s", cfg.Name))
	}

	if cfg.UseGit {
		steps = append(steps, "git log --oneline  # veja o commit inicial")
	}

	return steps
}

// ─────────────────────────────────────────────────────────────────────────────
// Formatting helpers
// ─────────────────────────────────────────────────────────────────────────────

// orDash returns the value if non-empty, or "—" otherwise.
func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

// boolLabel returns a user-friendly string for boolean values.
func boolLabel(b bool) string {
	if b {
		return "Sim"
	}
	return "Não"
}

// ─────────────────────────────────────────────────────────────────────────────
// Long help text
// ─────────────────────────────────────────────────────────────────────────────

// renderNewLong generates the detailed help text for the `new` subcommand,
// displayed with `cpp-gen new --help`.
func renderNewLong() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED"))

	accentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#06B6D4"))

	mutedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280"))

	_ = os.Stdout // ensures the "os" import is used

	return titleStyle.Render("cpp-gen new") + " — Cria um novo projeto C++ moderno\n\n" +
		"Por padrão executa em modo " + accentStyle.Render("interativo") +
		", guiando você por um formulário TUI passo a passo.\n" +
		"Para uso em scripts, passe " + accentStyle.Render("--no-interactive") +
		" junto com as flags desejadas.\n\n" +
		mutedStyle.Render("O projeto gerado inclui:") + "\n" +
		"  • CMakeLists.txt hierárquico (src/, tests/, cmake/)\n" +
		"  • Presets CMake (CMakePresets.json)\n" +
		"  • Configurações de IDE (tasks, launch, settings)\n" +
		"  • Integração com VCPKG ou FetchContent\n" +
		"  • .gitignore, README.md e commit inicial\n" +
		"  • .clangd e .clang-format pré-configurados"
}
