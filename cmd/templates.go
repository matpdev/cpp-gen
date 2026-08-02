package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/matpdev/cpp-gen/internal/generator/templatescaffold"
	"github.com/matpdev/cpp-gen/internal/registry"
	"github.com/matpdev/cpp-gen/internal/tui"
)

// ─────────────────────────────────────────────────────────────────────────────
// Subcommand: templates
// ─────────────────────────────────────────────────────────────────────────────

// templatesCmd lists or searches templates discoverable via internal/registry:
// the embedded default catalog, the GitHub community registry, and the
// user's local catalog (~/.config/cpp-gen/templates.json).
var templatesCmd = &cobra.Command{
	Use:   "templates [busca]",
	Short: "Lista ou busca templates disponíveis para `cpp-gen new --template`",
	Long: "Busca templates cadastrados localmente (~/.config/cpp-gen/templates.json) e no\n" +
		"registro remoto do GitHub. Sem argumentos, lista todos os templates conhecidos.\n\n" +
		"Use o nome retornado diretamente em:\n" +
		"  cpp-gen new meu-projeto --template <nome>",
	Args: cobra.MaximumNArgs(1),
	RunE: runTemplates,
}

func init() {
	rootCmd.AddCommand(templatesCmd)
	templatesCmd.AddCommand(templatesCreateCmd)

	templatesCreateCmd.Flags().String(
		"register", "",
		"Registra o template criado no catálogo local sob este nome (equivale a editar templates.json)",
	)
	templatesCreateCmd.Flags().String(
		"description", "",
		"Descrição salva junto com --register (mostrada em `cpp-gen templates`)",
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// Subcommand: templates create
// ─────────────────────────────────────────────────────────────────────────────

// templatesCreateCmd scaffolds a starter custom template — a folder using the
// placeholder/conditional-file conventions internal/generator/customtemplate
// expects — so the user has a real, working example to edit instead of
// starting from a blank folder.
var templatesCreateCmd = &cobra.Command{
	Use:   "create <diretório>",
	Short: "Cria um template inicial pronto pra ser customizado",
	Long: "Gera uma pasta com um template C++ mínimo (CMakeLists.txt, código fonte,\n" +
		"README) já usando as convenções de template do cpp-gen: variáveis como\n" +
		"{{.Name}} e {{.NameSnake}}, e um exemplo de arquivo condicional\n" +
		"({{if .UseVCPKG}}vcpkg.json{{end}}). Inclui um TEMPLATE_GUIDE.md com a\n" +
		"referência completa — apague-o quando terminar de customizar.",
	Args: cobra.ExactArgs(1),
	RunE: runTemplatesCreate,
}

func runTemplatesCreate(cmd *cobra.Command, args []string) error {
	dest := args[0]

	if err := templatescaffold.Generate(dest); err != nil {
		return fmt.Errorf("criar template: %w", err)
	}

	absDest, err := filepath.Abs(dest)
	if err != nil {
		absDest = dest
	}

	fmt.Println(tui.SuccessStyle.Render("✓ Template criado em " + absDest))
	fmt.Println()
	fmt.Println(tui.MutedStyle.Render("Leia " + filepath.Join(dest, "TEMPLATE_GUIDE.md") + " pra referência completa das variáveis."))

	registerAs, _ := cmd.Flags().GetString("register")
	if registerAs != "" {
		description, _ := cmd.Flags().GetString("description")
		entry := registry.Entry{
			Name:        registerAs,
			Description: description,
			Source:      absDest,
		}
		if err := registry.AddLocal(entry); err != nil {
			return fmt.Errorf("registrar template localmente: %w", err)
		}
		fmt.Println(tui.MutedStyle.Render(fmt.Sprintf("Registrado localmente como %q — rode `cpp-gen new <projeto> --template %s`.", registerAs, registerAs)))
	} else {
		fmt.Println(tui.MutedStyle.Render(fmt.Sprintf(
			"Teste com: cpp-gen new demo --template %s --no-interactive --output /tmp/scratch",
			dest,
		)))
	}

	return nil
}

func runTemplates(cmd *cobra.Command, args []string) error {
	query := ""
	if len(args) > 0 {
		query = args[0]
	}

	catalog := registry.Load(isVerbose(cmd))
	entries := catalog.Search(query)

	if len(entries) == 0 {
		if query == "" {
			fmt.Println(tui.MutedStyle.Render(
				"Nenhum template cadastrado ainda.\n" +
					"Adicione o seu em ~/.config/cpp-gen/templates.json, ou aguarde o registro público ficar disponível.",
			))
		} else {
			fmt.Println(tui.MutedStyle.Render(fmt.Sprintf("Nenhum template encontrado para %q.", query)))
		}
		return nil
	}

	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7C3AED"))
	sourceStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#06B6D4"))
	tagStyle := tui.MutedStyle

	fmt.Println(tui.FormatSection(fmt.Sprintf("Templates disponíveis (%d)", len(entries))))
	fmt.Println()

	for _, e := range entries {
		fmt.Printf("  %s\n", nameStyle.Render(e.Name))
		if e.Description != "" {
			fmt.Printf("    %s\n", e.Description)
		}
		fmt.Printf("    %s %s\n", tui.MutedStyle.Render("fonte:"), sourceStyle.Render(e.Source))
		if len(e.Tags) > 0 {
			fmt.Printf("    %s\n", tagStyle.Render(strings.Join(e.Tags, ", ")))
		}
		fmt.Println()
	}

	fmt.Println(tui.MutedStyle.Render("Use: cpp-gen new <nome-do-projeto> --template <nome-do-template>"))

	return nil
}
