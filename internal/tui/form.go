// Package tui contém todos os componentes de interface de usuário do terminal
// utilizados pelo cpp-gen, incluindo formulários interativos e estilos visuais.
package tui

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"cpp-gen/internal/config"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// ─────────────────────────────────────────────────────────────────────────────
// RunForm
// ─────────────────────────────────────────────────────────────────────────────

// RunForm exibe o formulário interativo TUI e retorna a ProjectConfig
// preenchida pelo usuário. Se o usuário cancelar (Esc / Ctrl+C),
// retorna um erro com a mensagem "user aborted".
//
// initialName é pré-preenchido no campo de nome quando fornecido via argumento
// posicional na linha de comando (ex: cpp-gen new meu-projeto).
func RunForm(initialName string) (*config.ProjectConfig, error) {
	cfg := config.Default()
	cfg.Name = initialName

	// Variáveis intermediárias de string para os campos de seleção,
	// pois huh.NewSelect requer *string enquanto config usa tipos customizados.
	var (
		standard    = string(cfg.Standard)
		projectType = string(cfg.ProjectType)
		pkgManager  = string(cfg.PackageManager)
		ide         = string(cfg.IDE)
	)

	// ── Grupo 1: Identidade do Projeto ────────────────────────────────────────
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

	// ── Grupo 2: Configurações Técnicas C++ ───────────────────────────────────
	groupTechnical := huh.NewGroup(
		huh.NewNote().
			Title("Configurações C++").
			Description("Defina o padrão da linguagem e o tipo de artefato."),

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

	// ── Grupo 3: Gerenciador de Pacotes ───────────────────────────────────────
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

	// ── Grupo 4: Ambiente de Desenvolvimento ──────────────────────────────────
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
			Description("Gera .clang-format com estilo Google/LLVM customizado.").
			Affirmative("Sim").
			Negative("Não").
			Value(&cfg.UseClangFormat),
	)

	// ── Construção e execução do formulário ───────────────────────────────────
	form := huh.NewForm(
		groupIdentity,
		groupTechnical,
		groupPackages,
		groupIDE,
	).
		WithTheme(buildTheme()).
		WithWidth(72)

	if err := form.Run(); err != nil {
		// huh retorna este erro quando o usuário pressiona Esc ou Ctrl+C
		if errors.Is(err, huh.ErrUserAborted) {
			return nil, errors.New("user aborted")
		}
		return nil, fmt.Errorf("erro no formulário: %w", err)
	}

	// Converte as variáveis de string de volta para os tipos customizados
	cfg.Standard = config.CppStandard(standard)
	cfg.ProjectType = config.ProjectType(projectType)
	cfg.PackageManager = config.PackageManager(pkgManager)
	cfg.IDE = config.IDE(ide)

	return cfg, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Validadores de campo
// ─────────────────────────────────────────────────────────────────────────────

// reProjectName define os caracteres válidos para nomes de projeto:
// letras minúsculas, dígitos e hífens, com início e fim alfanuméricos.
var reProjectName = regexp.MustCompile(`^[a-z0-9][a-z0-9\-]*[a-z0-9]$|^[a-z0-9]$`)

// validateProjectName valida o nome do projeto inserido pelo usuário.
// Regras:
//   - Não pode ser vazio
//   - Deve ter pelo menos 2 caracteres
//   - Apenas letras minúsculas, dígitos e hífens
//   - Não pode começar ou terminar com hífen
//   - Não pode conter espaços ou caracteres especiais
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

// reVersion valida o formato SemVer básico: MAJOR.MINOR.PATCH (todos numéricos).
var reVersion = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// validateVersion valida o campo de versão do projeto.
// Aceita o formato MAJOR.MINOR.PATCH (ex: "1.0.0", "0.3.12").
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
// Tema customizado
// ─────────────────────────────────────────────────────────────────────────────

// buildTheme cria um tema huh customizado baseado na paleta de cores do cpp-gen.
// Mantém o visual consistente com os estilos lipgloss definidos em styles.go.
func buildTheme() *huh.Theme {
	theme := huh.ThemeCharm()

	// Cabeçalho de grupo / nota
	theme.Focused.Title = TitleStyle.Copy()
	theme.Focused.Description = MutedStyle.Copy()

	// Campo selecionado / ativo
	theme.Focused.SelectedOption = lipglossColor(colorAccent)
	theme.Focused.UnselectedOption = MutedStyle.Copy()

	// Cursor de seleção
	theme.Focused.SelectSelector = InfoStyle.Copy()

	// Botões de confirmação
	theme.Focused.FocusedButton = lipglossColor(colorSuccess)
	theme.Focused.BlurredButton = MutedStyle.Copy()

	return theme
}

// lipglossColor cria um estilo lipgloss simples apenas com a cor de foreground,
// compatível com o tipo esperado pelos campos do tema huh.
func lipglossColor(c lipglossColorType) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(c)
}

// lipglossColorType é um alias interno para lipgloss.Color, tornando a assinatura
// de lipglossColor mais explícita e evitando imports circulares em testes.
type lipglossColorType = lipgloss.Color
