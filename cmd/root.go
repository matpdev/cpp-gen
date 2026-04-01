// Package cmd contém todos os comandos CLI do cpp-gen.
//
// Os comandos são construídos com a biblioteca Cobra e seguem a estrutura:
//
//	cpp-gen <comando> [flags]
//
// Comandos disponíveis:
//   - new   : Cria um novo projeto C++ (interativo ou via flags)
//   - version: Exibe a versão atual do cpp-gen
package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// ─────────────────────────────────────────────────────────────────────────────
// Metadados da aplicação
// ─────────────────────────────────────────────────────────────────────────────

const (
	// AppName é o nome canônico do binário CLI.
	AppName = "cpp-gen"

	// AppVersion é a versão atual do cpp-gen no formato SemVer.
	AppVersion = "1.0.0"

	// AppDescription é a descrição curta exibida no help raiz.
	AppDescription = "Gerador moderno de projetos C++ com CMake, VCPKG, FetchContent e suporte a IDEs."
)

// ─────────────────────────────────────────────────────────────────────────────
// Comando raiz
// ─────────────────────────────────────────────────────────────────────────────

// rootCmd é o comando base do CLI. Todos os subcomandos são registrados nele.
// Executá-lo sem subcomando exibe o banner e o help padrão do Cobra.
var rootCmd = &cobra.Command{
	Use:   AppName,
	Short: AppDescription,
	Long:  renderBanner(),

	// SilenceUsage evita que o Cobra imprima a mensagem de uso completa em
	// erros de runtime (apenas erros de parsing de flags mostram o uso).
	SilenceUsage: true,

	// SilenceErrors delega o tratamento de erros para o main(), que os formata
	// com o estilo de erro do cpp-gen antes de exibir ao usuário.
	SilenceErrors: true,
}

// ─────────────────────────────────────────────────────────────────────────────
// Inicialização
// ─────────────────────────────────────────────────────────────────────────────

// init registra os subcomandos e as flags persistentes do rootCmd.
// É chamado automaticamente pelo Go antes de main().
func init() {
	// Flags persistentes (disponíveis em todos os subcomandos)
	rootCmd.PersistentFlags().BoolP(
		"verbose", "v", false,
		"Ativa saída detalhada durante a geração do projeto",
	)

	// Subcomandos registrados
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(versionCmd)
}

// ─────────────────────────────────────────────────────────────────────────────
// Execute
// ─────────────────────────────────────────────────────────────────────────────

// Execute é o ponto de entrada público do pacote cmd.
// Deve ser chamado pelo main() para iniciar o processamento do CLI.
// Retorna qualquer erro encontrado durante a execução do comando.
func Execute() error {
	return rootCmd.Execute()
}

// ─────────────────────────────────────────────────────────────────────────────
// Subcomando: version
// ─────────────────────────────────────────────────────────────────────────────

// versionCmd exibe a versão atual do cpp-gen.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Exibe a versão atual do cpp-gen",
	Run: func(cmd *cobra.Command, args []string) {
		verStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED"))

		mutedStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

		fmt.Printf("%s %s\n", verStyle.Render(AppName), mutedStyle.Render("v"+AppVersion))
	},
}

// ─────────────────────────────────────────────────────────────────────────────
// Banner
// ─────────────────────────────────────────────────────────────────────────────

// renderBanner gera o texto de apresentação completo do CLI, exibido quando
// o usuário executa `cpp-gen` sem subcomandos ou com `--help`.
//
// O banner inclui:
//   - Título estilizado com a cor primária
//   - Versão atual
//   - Descrição resumida da ferramenta
//   - Lista de capacidades principais
func renderBanner() string {
	// ── Estilos ───────────────────────────────────────────────────────────────
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED"))

	versionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280"))

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A78BFA"))

	accentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#06B6D4"))

	mutedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280"))

	// ── Composição do banner ──────────────────────────────────────────────────
	title := titleStyle.Render("⚡ cpp-gen") +
		"  " +
		versionStyle.Render("v"+AppVersion)

	subtitle := subtitleStyle.Render(AppDescription)

	separator := mutedStyle.Render("─────────────────────────────────────────────────────")

	features := []string{
		accentStyle.Render("◆") + "  Estrutura de projeto CMake moderna (3.20+)",
		accentStyle.Render("◆") + "  Gerenciadores de pacotes: VCPKG ou FetchContent",
		accentStyle.Render("◆") + "  Configurações para VSCode, CLion e Neovim",
		accentStyle.Render("◆") + "  Git, .gitignore e README prontos",
		accentStyle.Render("◆") + "  Clangd e Clang-Format pré-configurados",
	}

	usage := mutedStyle.Render("Uso:") + "  cpp-gen new [nome-do-projeto]"

	banner := title + "\n" +
		subtitle + "\n\n" +
		separator + "\n"

	for _, f := range features {
		banner += "  " + f + "\n"
	}

	banner += separator + "\n\n" + usage

	return banner
}

// ─────────────────────────────────────────────────────────────────────────────
// Utilitários internos
// ─────────────────────────────────────────────────────────────────────────────

// isVerbose retorna true se a flag --verbose foi passada no comando raiz.
// Deve ser chamada dentro de subcomandos para condicionar saída detalhada.
func isVerbose(cmd *cobra.Command) bool {
	v, err := cmd.Root().PersistentFlags().GetBool("verbose")
	if err != nil {
		return false
	}
	return v
}

// printError imprime uma mensagem de erro formatada no stderr e encerra o processo.
// Usado pelo main() mas disponível para subcomandos que precisam abortar com estilo.
func printError(err error) {
	errStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#EF4444"))

	fmt.Fprintf(os.Stderr, "%s %s\n", errStyle.Render("Erro:"), err.Error())
}
