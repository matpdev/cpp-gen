// Package cmd contém todos os comandos CLI do cpp-gen.
package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"cpp-gen/internal/config"
	"cpp-gen/internal/generator"
	"cpp-gen/internal/tui"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// ─────────────────────────────────────────────────────────────────────────────
// Definição do comando
// ─────────────────────────────────────────────────────────────────────────────

// newCmd é o subcomando principal do cpp-gen. Ele guia o usuário pelo processo
// de criação de um novo projeto C++ através de um formulário TUI interativo ou,
// opcionalmente, através de flags para uso em scripts e automações.
//
// Uso:
//
//	cpp-gen new                        # totalmente interativo
//	cpp-gen new meu-projeto            # nome pré-preenchido, restante interativo
//	cpp-gen new --no-interactive [flags] # modo não-interativo (scripting)
var newCmd = &cobra.Command{
	Use:   "new [nome-do-projeto]",
	Short: "Cria um novo projeto C++ com CMake e ferramentas configuradas",
	Long:  renderNewLong(),

	// Aceita 0 ou 1 argumento posicional (nome do projeto).
	Args: cobra.MaximumNArgs(1),

	// RunE é preferível a Run pois permite retornar erros para tratamento
	// centralizado no main(), evitando chamadas manuais a os.Exit().
	RunE: runNew,

	// Exemplos exibidos no --help do subcomando.
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
// Registro de flags
// ─────────────────────────────────────────────────────────────────────────────

func init() {
	// ── Flags de controle ─────────────────────────────────────────────────────

	newCmd.Flags().StringP(
		"output", "o", ".",
		"Diretório de saída onde a pasta do projeto será criada",
	)

	newCmd.Flags().BoolP(
		"no-interactive", "n", false,
		"Desativa o formulário TUI; usa apenas as flags fornecidas",
	)

	// ── Flags de metadados (modo não-interativo) ───────────────────────────────

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

	// ── Flags de configuração técnica ─────────────────────────────────────────

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

	// ── Flags de funcionalidades opcionais ────────────────────────────────────

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
}

// ─────────────────────────────────────────────────────────────────────────────
// Handler principal
// ─────────────────────────────────────────────────────────────────────────────

// runNew é o handler do comando `new`. Decide entre modo interativo e
// não-interativo, coleta a configuração, executa o gerador e exibe o resumo.
func runNew(cmd *cobra.Command, args []string) error {
	outputDir, _ := cmd.Flags().GetString("output")
	noInteractive, _ := cmd.Flags().GetBool("no-interactive")

	// Nome do projeto: pode vir como argumento posicional ou pela flag --name.
	// O argumento posicional tem precedência sobre a flag.
	initialName := ""
	if len(args) > 0 {
		initialName = args[0]
	} else if flagName, _ := cmd.Flags().GetString("name"); flagName != "" {
		initialName = flagName
	}

	// ── Coleta da configuração ────────────────────────────────────────────────

	var cfg *config.ProjectConfig
	var err error

	if noInteractive {
		cfg, err = buildConfigFromFlags(cmd, initialName)
		if err != nil {
			return fmt.Errorf("configuração inválida: %w", err)
		}
	} else {
		// Modo interativo: abre o formulário TUI com o nome pré-preenchido.
		cfg, err = tui.RunForm(initialName)
		if err != nil {
			// Cancelamento pelo usuário não é um erro — apenas encerra silenciosamente.
			if err.Error() == "user aborted" {
				fmt.Println(tui.MutedStyle.Render("\nOperação cancelada."))
				return nil
			}
			return fmt.Errorf("erro no formulário: %w", err)
		}
	}

	// Aplica o diretório de saída definido via flag --output.
	cfg.OutputDir = outputDir

	// ── Validação final ───────────────────────────────────────────────────────

	if errs := cfg.Validate(); len(errs) > 0 {
		return fmt.Errorf("configuração incompleta:\n  • %s", strings.Join(errs, "\n  • "))
	}

	// ── Exibe resumo antes de gerar ───────────────────────────────────────────

	printProjectSummary(cfg)

	// ── Execução do gerador ───────────────────────────────────────────────────

	gen := generator.New(cfg, isVerbose(cmd))

	if err := gen.Generate(); err != nil {
		return fmt.Errorf("falha ao gerar o projeto: %w", err)
	}

	// ── Mensagem de conclusão ─────────────────────────────────────────────────

	printSuccessMessage(cfg)

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Modo não-interativo
// ─────────────────────────────────────────────────────────────────────────────

// buildConfigFromFlags constrói uma ProjectConfig exclusivamente a partir das
// flags da linha de comando, sem abrir nenhum formulário TUI.
//
// Usado quando --no-interactive é passado, ideal para pipelines de CI/CD
// e scripts de automação.
func buildConfigFromFlags(cmd *cobra.Command, initialName string) (*config.ProjectConfig, error) {
	cfg := config.Default()

	// ── Nome ──────────────────────────────────────────────────────────────────
	if initialName != "" {
		cfg.Name = initialName
	}

	// ── Metadados ─────────────────────────────────────────────────────────────
	if v, _ := cmd.Flags().GetString("description"); v != "" {
		cfg.Description = v
	}
	if v, _ := cmd.Flags().GetString("author"); v != "" {
		cfg.Author = v
	}
	if v, _ := cmd.Flags().GetString("version"); v != "" {
		cfg.Version = v
	}

	// ── Layout de pastas ──────────────────────────────────────────────────────
	if layoutStr, _ := cmd.Flags().GetString("layout"); layoutStr != "" {
		l, err := parseLayout(layoutStr)
		if err != nil {
			return nil, err
		}
		cfg.Layout = l
	}

	// ── Padrão C++ ────────────────────────────────────────────────────────────
	if stdStr, _ := cmd.Flags().GetString("std"); stdStr != "" {
		std, err := parseCppStandard(stdStr)
		if err != nil {
			return nil, err
		}
		cfg.Standard = std
	}

	// ── Tipo de projeto ───────────────────────────────────────────────────────
	if typeStr, _ := cmd.Flags().GetString("type"); typeStr != "" {
		pt, err := parseProjectType(typeStr)
		if err != nil {
			return nil, err
		}
		cfg.ProjectType = pt
	}

	// ── Gerenciador de pacotes ────────────────────────────────────────────────
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

	// ── Flags de opt-out ──────────────────────────────────────────────────────
	noGit, _ := cmd.Flags().GetBool("no-git")
	noClangd, _ := cmd.Flags().GetBool("no-clangd")
	noClangFmt, _ := cmd.Flags().GetBool("no-clang-format")

	cfg.UseGit = !noGit
	cfg.UseClangd = !noClangd
	cfg.UseClangFormat = !noClangFmt

	// ── Validação obrigatória em modo não-interativo ───────────────────────────
	if cfg.Name == "" {
		return nil, errors.New(
			"nome do projeto é obrigatório; use o argumento posicional ou --name",
		)
	}

	return cfg, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Parsers de enum (flags → tipos config)
// ─────────────────────────────────────────────────────────────────────────────

// parseCppStandard converte uma string (ex: "20") para config.CppStandard.
// Retorna erro se o valor não for reconhecido.
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

// parseProjectType converte uma string (ex: "executable") para config.ProjectType.
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

// parsePackageManager converte uma string (ex: "vcpkg") para config.PackageManager.
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

// parseIDE converte uma string (ex: "vscode") para config.IDE.
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

// parseLayout converte uma string (ex: "merged") para config.FolderLayout.
// Aceita variações comuns com hífen, underscore e sem separador.
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

// ─────────────────────────────────────────────────────────────────────────────
// Saída formatada
// ─────────────────────────────────────────────────────────────────────────────

// printProjectSummary exibe um resumo formatado da configuração do projeto
// antes de iniciar a geração, permitindo ao usuário revisar as escolhas.
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
		tui.FormatKeyValue("Clang-Format", boolLabel(cfg.UseClangFormat)),
		tui.FormatKeyValue("Destino", cfg.ProjectPath()),
	}

	fmt.Println(boxStyle.Render(strings.Join(rows, "\n")))
}

// printSuccessMessage exibe a mensagem final de conclusão com as instruções
// de primeiros passos (next steps) para o usuário.
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

// buildNextSteps retorna uma lista de comandos sugeridos para o usuário
// executar após a geração do projeto.
func buildNextSteps(cfg *config.ProjectConfig) []string {
	projectPath := cfg.ProjectPath()
	steps := []string{
		fmt.Sprintf("cd %s", projectPath),
	}

	// Instrução de configure CMake (varia conforme gerenciador de pacotes)
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
// Helpers de formatação
// ─────────────────────────────────────────────────────────────────────────────

// orDash retorna o valor se não for vazio, ou "—" caso contrário.
func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

// boolLabel retorna uma string amigável para valores booleanos.
func boolLabel(b bool) string {
	if b {
		return "Sim"
	}
	return "Não"
}

// ─────────────────────────────────────────────────────────────────────────────
// Texto de ajuda longa
// ─────────────────────────────────────────────────────────────────────────────

// printSuccessMessage exibe a mensagem final de conclusão com as instruções
// de primeiros passos (next steps) para o usuário.
// renderNewLong gera o texto de ajuda detalhado do subcomando `new`,
// exibido com `cpp-gen new --help`.
func renderNewLong() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED"))

	accentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#06B6D4"))

	mutedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280"))

	_ = os.Stdout // garante o import de "os"

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
