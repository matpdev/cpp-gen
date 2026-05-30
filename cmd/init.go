// Package cmd contains all CLI commands for cpp-gen.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/matpdev/cpp-gen/internal/localconfig"
	"github.com/spf13/cobra"
)

// ─────────────────────────────────────────────────────────────────────────────
// Command definition
// ─────────────────────────────────────────────────────────────────────────────

// initCmd initializes a .cppgenrc.json file in an existing C++ project that
// was not created by cpp-gen, or that needs to regenerate its config.
//
// With --global it writes to ~/.config/cpp-gen/config.json instead, setting
// device-wide defaults used whenever no local config is present.
//
// Usage:
//
//	cpp-gen init                         # fully interactive, local
//	cpp-gen init --global                # fully interactive, global
//	cpp-gen init --no-interactive [flags] # non-interactive mode (scripting)
//	cpp-gen init --force                  # overwrite existing config
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Inicializa um arquivo .cppgenrc.json no projeto atual (ou config global com --global)",
	Long: `Inicializa um arquivo .cppgenrc.json no diretório atual.

Use este comando em projetos C++ existentes que não foram criados pelo
cpp-gen, ou quando precisar regenerar o arquivo de configuração local.

Com a flag --global, salva em ~/.config/cpp-gen/config.json.
Essa configuração será usada automaticamente em qualquer projeto que
não possua um .cppgenrc.json local.

Prioridade de configuração:
  1. Local   — .cppgenrc.json na raiz do projeto
  2. Global  — ~/.config/cpp-gen/config.json
  3. Padrão — valores embutidos no cpp-gen

Exemplos:
  # Modo interativo local (recomendado)
  cpp-gen init

  # Configuração global do dispositivo
  cpp-gen init --global

  # Modo não-interativo para scripts e CI
  cpp-gen init --no-interactive --author "Fulano" --license MIT

  # Sobrescrever configuração existente
  cpp-gen init --force`,

	RunE: runInit,

	Example: `  # Modo interativo local (recomendado)
  cpp-gen init

  # Configuração global do dispositivo
  cpp-gen init --global

  # Modo não-interativo para scripts e CI
  cpp-gen init \
    --no-interactive \
    --author "Fulano" \
    --license MIT \
    --header-style doxygen \
    --test-framework catch2

  # Sobrescrever configuração global existente
  cpp-gen init --global --force`,
}

// ─────────────────────────────────────────────────────────────────────────────
// Flag registration
// ─────────────────────────────────────────────────────────────────────────────

func init() {
	// ── Control flags ────────────────────────────────────────────────────

	initCmd.Flags().BoolP(
		"no-interactive", "n", false,
		"Desativa o formulário TUI; usa apenas as flags fornecidas",
	)
	initCmd.Flags().BoolP(
		"force", "f", false,
		"Sobrescreve o arquivo de configuração se já existir",
	)
	initCmd.Flags().BoolP(
		"global", "g", false,
		"Salva em ~/.config/cpp-gen/config.json (configuração global do dispositivo)",
	)

	// ── Metadata flags ────────────────────────────────────────────────────────

	initCmd.Flags().String(
		"author", "",
		"Nome do autor ou organização",
	)
	initCmd.Flags().String(
		"license", "MIT",
		"Licença do projeto: MIT | Apache-2.0 | GPL-3.0 | none",
	)

	// ── Path flags ────────────────────────────────────────────────────────────

	initCmd.Flags().String(
		"src", "src",
		"Diretório dos arquivos de implementação (.cpp)",
	)
	initCmd.Flags().String(
		"include", "include",
		"Diretório dos arquivos de cabeçalho (.hpp)",
	)
	initCmd.Flags().String(
		"tests", "tests",
		"Diretório dos arquivos de teste",
	)

	// ── Header flags ──────────────────────────────────────────────────────────

	initCmd.Flags().String(
		"header-style", "doxygen",
		"Estilo do cabeçalho de arquivo: doxygen | block | line | none",
	)

	// ── Namespace flag ────────────────────────────────────────────────────────

	initCmd.Flags().Bool(
		"namespace", true,
		"Envolve o código gerado em um namespace C++",
	)

	// ── Test framework flag ───────────────────────────────────────────────────

	initCmd.Flags().String(
		"test-framework", "catch2",
		"Framework de testes: catch2 | gtest | none",
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// Main handler
// ─────────────────────────────────────────────────────────────────────────────

// runInit is the handler for the `init` command. It decides between interactive
// and non-interactive mode, builds the LocalConfig and writes .cppgenrc.json.
func runInit(cmd *cobra.Command, args []string) error {
	// ── Resolve scope (local vs global) ──────────────────────────────────

	globalMode, _ := cmd.Flags().GetBool("global")
	force, _ := cmd.Flags().GetBool("force")

	// ── Resolve target directory and project name ────────────────────────

	var targetDir string
	var projectName string

	if globalMode {
		dir, err := localconfig.GlobalConfigDir()
		if err != nil {
			return fmt.Errorf("não foi possível determinar o diretório de config global: %w", err)
		}
		targetDir = dir
		projectName = "" // global config has no project name
	} else {
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("não foi possível determinar o diretório atual: %w", err)
		}
		targetDir = dir
		projectName = filepath.Base(dir)
	}

	// ── Guard: existing config ────────────────────────────────────────────

	errStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#EF4444"))
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))

	alreadyExists := func() bool {
		if globalMode {
			return localconfig.GlobalExists()
		}
		return localconfig.Exists(targetDir)
	}()

	if alreadyExists && !force {
		targetLabel := localconfig.FileName
		if globalMode {
			targetLabel = "~/.config/cpp-gen/config.json"
		}
		fmt.Fprintf(os.Stderr, "%s %s já existe.\n",
			errStyle.Render("Erro:"), targetLabel)
		fmt.Fprintf(os.Stderr, "  %s\n",
			hintStyle.Render("Use --force / -f para sobrescrever."),
		)
		return nil
	}

	// ── Collect configuration ──────────────────────────────────────────────

	noInteractive, _ := cmd.Flags().GetBool("no-interactive")

	var (
		author        string
		license       string
		srcDir        string
		includeDir    string
		testsDir      string
		headerStyle   string
		headerFields  []string
		nsEnabled     bool
		testFramework string
		createTest    bool
	)

	if noInteractive {
		author, _ = cmd.Flags().GetString("author")
		license, _ = cmd.Flags().GetString("license")
		srcDir, _ = cmd.Flags().GetString("src")
		includeDir, _ = cmd.Flags().GetString("include")
		testsDir, _ = cmd.Flags().GetString("tests")
		headerStyle, _ = cmd.Flags().GetString("header-style")
		nsEnabled, _ = cmd.Flags().GetBool("namespace")
		testFramework, _ = cmd.Flags().GetString("test-framework")
		createTest = true
		headerFields = []string{"file", "brief", "author", "date"}
	} else {
		// Seed with flag defaults so interactive mode shows sensible values.
		author, _ = cmd.Flags().GetString("author")
		license, _ = cmd.Flags().GetString("license")
		srcDir, _ = cmd.Flags().GetString("src")
		includeDir, _ = cmd.Flags().GetString("include")
		testsDir, _ = cmd.Flags().GetString("tests")
		headerStyle, _ = cmd.Flags().GetString("header-style")
		nsEnabled, _ = cmd.Flags().GetBool("namespace")
		testFramework, _ = cmd.Flags().GetString("test-framework")
		headerFields = []string{"file", "brief", "author", "date"}
		createTest = true

		form := huh.NewForm(
			// ── Group 1: Metadata ────────────────────────────────────────
			huh.NewGroup(
				huh.NewInput().
					Title("Autor").
					Description("Nome do autor ou organização").
					Value(&author),

				huh.NewSelect[string]().
					Title("Licença").
					Options(
						huh.NewOption("MIT", "MIT"),
						huh.NewOption("Apache-2.0", "Apache-2.0"),
						huh.NewOption("GPL-3.0", "GPL-3.0"),
						huh.NewOption("Nenhuma", "none"),
					).
					Value(&license),
			),

			// ── Group 2: Header ──────────────────────────────────────────
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Estilo de cabeçalho").
					Description("Formato do comentário gerado no topo de cada arquivo").
					Options(
						huh.NewOption("Doxygen (/** @file ... */)", "doxygen"),
						huh.NewOption("Block (/* ... */)", "block"),
						huh.NewOption("Line (// ...)", "line"),
						huh.NewOption("Nenhum", "none"),
					).
					Value(&headerStyle),

				huh.NewMultiSelect[string]().
					Title("Campos do cabeçalho").
					Description("Selecione os campos incluídos no cabeçalho").
					Options(
						huh.NewOption("file", "file"),
						huh.NewOption("brief", "brief"),
						huh.NewOption("author", "author"),
						huh.NewOption("date", "date"),
						huh.NewOption("version", "version"),
						huh.NewOption("copyright", "copyright"),
						huh.NewOption("license", "license"),
					).
					Value(&headerFields),
			),

			// ── Group 3: Code generation ─────────────────────────────────
			huh.NewGroup(
				huh.NewConfirm().
					Title("Habilitar namespace?").
					Description("O código gerado será envolvido em um namespace C++").
					Value(&nsEnabled),

				huh.NewSelect[string]().
					Title("Framework de testes").
					Options(
						huh.NewOption("Catch2", "catch2"),
						huh.NewOption("Google Test", "gtest"),
						huh.NewOption("Nenhum", "none"),
					).
					Value(&testFramework),

				huh.NewConfirm().
					Title("Criar arquivo de teste ao gerar?").
					Description("cpp-gen generate criará automaticamente um arquivo de teste").
					Value(&createTest),
			),
		)

		if err := form.Run(); err != nil {
			// User pressed Ctrl+C or Esc — abort gracefully.
			mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
			fmt.Println(mutedStyle.Render("Operação cancelada."))
			return nil
		}
	}

	// ── Build config ───────────────────────────────────────────────────────

	cfg := localconfig.DefaultLocalConfig(projectName, author)

	cfg.Author = author
	cfg.License = license

	cfg.Paths.Src = srcDir
	cfg.Paths.Include = includeDir
	cfg.Paths.Tests = testsDir

	headerEnabled := headerStyle != "none"
	cfg.Header.Enabled = headerEnabled
	cfg.Header.Style = headerStyle
	if len(headerFields) > 0 {
		cfg.Header.Fields = headerFields
	}

	cfg.Namespace.Enabled = nsEnabled

	cfg.Generate.TestFramework = testFramework
	cfg.Generate.CreateTest = createTest

	// ── Write file ──────────────────────────────────────────────────────

	var outputPath string
	if globalMode {
		if err := localconfig.WriteGlobal(cfg); err != nil {
			return fmt.Errorf("erro ao escrever config global: %w", err)
		}
		globalDir, _ := localconfig.GlobalConfigDir()
		outputPath = filepath.Join(globalDir, localconfig.GlobalFileName)
	} else {
		if err := localconfig.Write(targetDir, cfg); err != nil {
			return fmt.Errorf("erro ao escrever %s: %w", localconfig.FileName, err)
		}
		outputPath = filepath.Join(targetDir, localconfig.FileName)
	}

	// ── Success message ──────────────────────────────────────────────

	successStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#22C55E"))

	pathStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A78BFA"))

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8"))

	scopeLabel := "local"
	if globalMode {
		scopeLabel = "global"
	}

	fmt.Printf("\n%s  %s %s\n\n",
		successStyle.Render("✓"),
		labelStyle.Render("["+scopeLabel+"]"),
		pathStyle.Render(outputPath),
	)

	return nil
}
