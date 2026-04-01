// Package cmd contains all CLI commands for cpp-gen.
//
// Commands are built with the Cobra library and follow the structure:
//
//	cpp-gen <command> [flags]
//
// Available commands:
//   - new    : Creates a new C++ project (interactive or via flags)
//   - version: Displays the current version of cpp-gen
package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// ─────────────────────────────────────────────────────────────────────────────
// Application metadata
// ─────────────────────────────────────────────────────────────────────────────

const (
	// AppName is the canonical name of the CLI binary.
	AppName = "cpp-gen"

	// AppDescription is the short description shown in root help.
	AppDescription = "Gerador moderno de projetos C++ com CMake, VCPKG, FetchContent e suporte a IDEs."
)

// Variables injected at compile time via ldflags by goreleaser and the Makefile.
// Must be `var` (never `const`) for the injection to work.
//
//	-X github.com/matpdev/cpp-gen/cmd.AppVersion=1.2.3
//	-X github.com/matpdev/cpp-gen/cmd.BuildDate=2024-01-01T00:00:00Z
//	-X github.com/matpdev/cpp-gen/cmd.GitCommit=abc1234
var (
	// AppVersion is the current version in SemVer format. Default value "dev"
	// indicates a local build without ldflags injection.
	AppVersion = "dev"

	// BuildDate is the UTC build timestamp in RFC 3339 format.
	BuildDate = "unknown"

	// GitCommit is the short hash of the HEAD commit at build time.
	GitCommit = "unknown"
)

// ─────────────────────────────────────────────────────────────────────────────
// Root command
// ─────────────────────────────────────────────────────────────────────────────

// rootCmd is the base CLI command. All subcommands are registered on it.
// Running it without a subcommand displays the banner and Cobra's default help.
var rootCmd = &cobra.Command{
	Use:   AppName,
	Short: AppDescription,
	Long:  renderBanner(),

	// SilenceUsage prevents Cobra from printing the full usage message on
	// runtime errors (only flag parsing errors show the usage).
	SilenceUsage: true,

	// SilenceErrors delegates error handling to main(), which formats them
	// with cpp-gen's error style before displaying to the user.
	SilenceErrors: true,
}

// ─────────────────────────────────────────────────────────────────────────────
// Initialization
// ─────────────────────────────────────────────────────────────────────────────

// init registers subcommands and persistent flags on rootCmd.
// It is called automatically by Go before main().
func init() {
	// Persistent flags (available in all subcommands)
	rootCmd.PersistentFlags().BoolP(
		"verbose", "v", false,
		"Ativa saída detalhada durante a geração do projeto",
	)

	// Registered subcommands
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(versionCmd)
}

// ─────────────────────────────────────────────────────────────────────────────
// Execute
// ─────────────────────────────────────────────────────────────────────────────

// Execute is the public entry point of the cmd package.
// It must be called by main() to start CLI processing.
// Returns any error encountered during command execution.
func Execute() error {
	return rootCmd.Execute()
}

// ─────────────────────────────────────────────────────────────────────────────
// Subcommand: version
// ─────────────────────────────────────────────────────────────────────────────

// versionCmd displays the current version of cpp-gen.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Exibe a versão atual do cpp-gen",
	Run: func(cmd *cobra.Command, args []string) {
		verStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED"))

		labelStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A78BFA"))

		mutedStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

		fmt.Printf("%s %s\n\n",
			verStyle.Render(AppName),
			mutedStyle.Render("v"+AppVersion),
		)
		fmt.Printf("  %s  %s\n", labelStyle.Render("commit  "), mutedStyle.Render(GitCommit))
		fmt.Printf("  %s  %s\n", labelStyle.Render("built   "), mutedStyle.Render(BuildDate))
		fmt.Println()
	},
}

// ─────────────────────────────────────────────────────────────────────────────
// Banner
// ─────────────────────────────────────────────────────────────────────────────

// renderBanner generates the full presentation text of the CLI, displayed when
// the user runs `cpp-gen` without subcommands or with `--help`.
//
// The banner includes:
//   - Stylized title with primary color
//   - Current version
//   - Brief tool description
//   - List of main capabilities
func renderBanner() string {
	// ── Styles ────────────────────────────────────────────────────────────────
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

	// ── Banner composition ────────────────────────────────────────────────────
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
// Internal utilities
// ─────────────────────────────────────────────────────────────────────────────

// isVerbose returns true if the --verbose flag was passed on the root command.
// Should be called inside subcommands to conditionally produce verbose output.
func isVerbose(cmd *cobra.Command) bool {
	v, err := cmd.Root().PersistentFlags().GetBool("verbose")
	if err != nil {
		return false
	}
	return v
}

// printError prints a formatted error message to stderr and terminates the process.
// Used by main() but available to subcommands that need to abort with style.
func printError(err error) {
	errStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#EF4444"))

	fmt.Fprintf(os.Stderr, "%s %s\n", errStyle.Render("Erro:"), err.Error())
}
