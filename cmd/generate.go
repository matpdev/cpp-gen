package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/matpdev/cpp-gen/internal/generator/filegen"
	"github.com/matpdev/cpp-gen/internal/localconfig"
	"github.com/spf13/cobra"
)

// ─────────────────────────────────────────────────────────────────────────────
// Flags
// ─────────────────────────────────────────────────────────────────────────────

var (
	genFlagType          string
	genFlagOutput        string
	genFlagNoTest        bool
	genFlagNoInteractive bool
	genFlagBrief         string
	genFlagLayer         string
	genFlagList          bool
)

// ─────────────────────────────────────────────────────────────────────────────
// Command
// ─────────────────────────────────────────────────────────────────────────────

var generateCmd = &cobra.Command{
	Use:     "generate [type|schematic] [name]",
	Aliases: []string{"g"},
	Short:   "Gera um arquivo C++ baseado na configuração do projeto (.cppgenrc.json)",
	Long: `Gera arquivos C++ (.hpp / .cpp) no projeto atual com base nas configurações
definidas em .cppgenrc.json.

Tipos primitivos:
  class      — par .hpp/.cpp com classe C++
  struct     — header-only com struct C++
  free       — módulo de funções livres (.hpp + .cpp)
  interface  — classe abstrata pura (header-only)

Schematics (definidos pela arquitetura do projeto):
  service    — serviço de domínio com interface
  repository — interface + implementação de acesso a dados
  command    — padrão Command
  (use --list para ver todos os schematics disponíveis)

O comando lê .cppgenrc.json a partir do diretório atual (percorrendo a árvore
de diretórios para cima até encontrar o arquivo). Se não encontrar, execute
'cpp-gen init' primeiro para inicializar o projeto.`,
	Args: cobra.MaximumNArgs(2),
	RunE: runGenerate,
	Example: `  cpp-gen generate class Foo
  cpp-gen generate service UserService --layer application
  cpp-gen generate repository UserRepo
  cpp-gen g class Renderer --no-test
  cpp-gen g class Foo --output src/graphics
  cpp-gen generate --list`,
}

func init() {
	generateCmd.Flags().StringVarP(&genFlagType, "type", "t", "",
		"Tipo de arquivo a gerar: class | struct | free | interface")
	generateCmd.Flags().StringVarP(&genFlagOutput, "output", "o", "",
		"Diretório de saída (sobrescreve os paths do .cppgenrc.json)")
	generateCmd.Flags().BoolVar(&genFlagNoTest, "no-test", false,
		"Não gera arquivo de teste")
	generateCmd.Flags().BoolVarP(&genFlagNoInteractive, "no-interactive", "n", false,
		"Modo não-interativo (falha se faltar nome ou tipo)")
	generateCmd.Flags().StringVarP(&genFlagBrief, "brief", "b", "",
		"Descrição curta para o campo @brief do cabeçalho")
	generateCmd.Flags().StringVarP(&genFlagLayer, "layer", "l", "",
		"Camada arquitetural de destino (ex: application, infrastructure)")
	generateCmd.Flags().BoolVar(&genFlagList, "list", false,
		"Lista todos os schematics disponíveis no projeto")
}

// ─────────────────────────────────────────────────────────────────────────────
// Runner
// ─────────────────────────────────────────────────────────────────────────────

func runGenerate(cmd *cobra.Command, args []string) error {
	// ── Styles ────────────────────────────────────────────────────────────────
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	yellow := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))

	// ── 1. Detect working directory ───────────────────────────────────────────
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("não foi possível determinar o diretório atual: %w", err)
	}

	// ── 2. Load config (local → global → default) ───────────────────────────
	cfg, src, err := localconfig.Resolve(dir)
	if err != nil {
		return fmt.Errorf("erro ao carregar configuração: %w", err)
	}

	// Inform the user which config source is active (skip for default to avoid
	// noise when the user just wants a quick generation).
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	switch src {
	case localconfig.SourceGlobal:
		fmt.Fprintln(os.Stderr, muted.Render("ℹ usando configuração global (~/.config/cpp-gen/config.json)"))
	case localconfig.SourceDefault:
		fmt.Fprintln(os.Stderr, muted.Render("ℹ nenhuma configuração encontrada; usando padrões embutidos"))
	}

	// ── --list: mostrar schematics disponíveis e sair ─────────────────────────
	if genFlagList {
		return runListSchematics(cfg)
	}

	// ── 3. Resolve fileType/schematicName and fileName ─────────────────────────
	primitiveTypes := map[string]bool{
		"class": true, "struct": true, "free": true, "interface": true,
	}

	schematicName := ""
	fileType := ""
	fileName := ""

	// args[0] = type or schematic name; args[1] = entity name
	if len(args) >= 1 {
		arg0 := strings.ToLower(args[0])
		if primitiveTypes[arg0] {
			fileType = arg0
		} else {
			schematicName = arg0
		}
	}
	if len(args) >= 2 {
		fileName = args[1]
	}

	// --type flag forces primitive mode
	if genFlagType != "" {
		t := strings.ToLower(genFlagType)
		if primitiveTypes[t] {
			fileType = t
			schematicName = ""
		}
	}

	// If neither was resolved, fall back to config default type
	if fileType == "" && schematicName == "" {
		fileType = cfg.Generate.DefaultType
	}

	// ── 4. Prompt for name if missing ─────────────────────────────────────────
	if fileName == "" && !genFlagNoInteractive {
		if err := huh.NewInput().
			Title("Nome do arquivo").
			Description("Nome da classe, struct ou módulo a gerar (ex: FooBar, utils/math)").
			Value(&fileName).
			Run(); err != nil {
			return fmt.Errorf("entrada cancelada: %w", err)
		}
		fileName = strings.TrimSpace(fileName)
	}

	// ── 5. Prompt for type/schematic if still missing ─────────────────────────
	if fileType == "" && schematicName == "" && !genFlagNoInteractive {
		if err := huh.NewSelect[string]().
			Title("Tipo de arquivo").
			Description("Selecione o tipo de construct C++ a gerar").
			Options(
				huh.NewOption("class     — par .hpp/.cpp com classe C++", "class"),
				huh.NewOption("struct    — header-only com struct C++", "struct"),
				huh.NewOption("free      — módulo de funções livres (.hpp + .cpp)", "free"),
				huh.NewOption("interface — classe abstrata pura (header-only)", "interface"),
			).
			Value(&fileType).
			Run(); err != nil {
			return fmt.Errorf("seleção cancelada: %w", err)
		}
	}

	// ── 6. Validate that name is present ─────────────────────────────────────
	if fileName == "" {
		return fmt.Errorf("nome do arquivo não fornecido; use 'cpp-gen generate [type] [name]' ou omita --no-interactive")
	}

	// ── Gerar via schematic ou tipo primitivo ─────────────────────────────────
	var generated []filegen.GeneratedFile

	if schematicName != "" {
		req := filegen.SchematicRequest{
			Name:          fileName,
			SchematicName: schematicName,
			Layer:         genFlagLayer,
			Brief:         genFlagBrief,
			NoTest:        genFlagNoTest,
		}
		generated, err = filegen.GenerateSchematic(req, cfg)
	} else {
		req := filegen.FileRequest{
			Name:      fileName,
			Type:      filegen.FileType(fileType),
			Brief:     genFlagBrief,
			OutputDir: genFlagOutput,
			NoTest:    genFlagNoTest,
		}
		generated, err = filegen.Generate(req, cfg)
	}
	if err != nil {
		return fmt.Errorf("erro durante a geração: %w", err)
	}

	// ── Report results ────────────────────────────────────────────────────────
	for _, gf := range generated {
		if gf.Skipped {
			fmt.Println(yellow.Render("⚠ já existe: " + gf.Path))
		} else {
			fmt.Println(green.Render("✓ criado: " + gf.Path))
		}
	}

	return nil
}

// runListSchematics prints all schematics available in the current project.
func runListSchematics(cfg *localconfig.LocalConfig) error {
	entries := filegen.ListSchematics(cfg)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7C3AED"))
	builtinStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	customStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#06B6D4"))
	nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Width(16)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	countStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	archLabel := string(cfg.Architecture.Style)
	placementLabel := string(cfg.Architecture.Placement)
	fmt.Printf("\n%s\n\n",
		titleStyle.Render(fmt.Sprintf("Schematics disponíveis (%s · %s)", archLabel, placementLabel)),
	)

	lastWasCustom := false
	printedBuiltin := false
	for _, e := range entries {
		if e.IsCustom && !lastWasCustom {
			fmt.Printf("  %s\n", customStyle.Render("custom (.cppgenrc.json)"))
			lastWasCustom = true
		} else if !e.IsCustom && !printedBuiltin {
			fmt.Printf("  %s\n", builtinStyle.Render("built-in"))
			printedBuiltin = true
		}
		fmt.Printf("    %s %s %s\n",
			nameStyle.Render(e.Name),
			descStyle.Render(e.Description),
			countStyle.Render(fmt.Sprintf("(%d arquivo(s))", e.FileCount)),
		)
	}
	fmt.Println()
	return nil
}
