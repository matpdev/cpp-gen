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
	"github.com/matpdev/cpp-gen/internal/generator/customtemplate"
	"github.com/matpdev/cpp-gen/internal/localconfig"
	"github.com/matpdev/cpp-gen/internal/registry"

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
// templateSearchValue is a sentinel templateType value (not a real
// config.ProjectTemplate) meaning "the user is picking from the template
// registry" — resolved into TemplateCustom + a concrete source once a
// catalog entry is chosen in groupTemplateSearch.
const templateSearchValue = "search"

// templateOptions builds the options for the template selection field.
// The "Procurar templates..." option only appears when the registry
// (internal/registry) actually has something to show.
func templateOptions(hasSearchable bool) []huh.Option[string] {
	opts := []huh.Option[string]{
		huh.NewOption(
			"Em branco  — projeto C++ padrão configurável",
			string(config.TemplateBlank),
		),
		huh.NewOption(
			"Vulkan     — aplicação Vulkan com vklib, GLFW, GLM, ImGui",
			string(config.TemplateVulkan),
		),
		huh.NewOption(
			"Custom     — repositório Git ou pasta local",
			string(config.TemplateCustom),
		),
	}
	if hasSearchable {
		opts = append(opts, huh.NewOption(
			"Procurar   — buscar no catálogo de templates (local + GitHub)",
			templateSearchValue,
		))
	}
	return opts
}

// RunForm's initialTemplate parameter, when non-empty, pre-selects the
// template ("blank", "vulkan", a registry alias, or an explicit source) and
// skips the template selection step entirely — matching what a user expects
// when they've already passed `cpp-gen new --template X` on the command line.
func RunForm(initialName, initialTemplate string) (*config.ProjectConfig, error) {
	cfg := config.Default()
	cfg.Name = initialName

	// Best-effort template registry lookup (internal/registry): embedded
	// defaults + local user file, plus the GitHub registry when reachable.
	// Non-verbose — an unreachable registry should never clutter the form,
	// it just means the "Procurar templates..." option won't appear.
	catalog := registry.Load(false)
	sourceToEntry := make(map[string]registry.Entry, len(catalog.Entries))
	searchOptions := make([]huh.Option[string], 0, len(catalog.Entries))
	for _, e := range catalog.Entries {
		sourceToEntry[e.Source] = e
		searchOptions = append(searchOptions, huh.NewOption(e.Name, e.Source))
	}

	// Intermediate string variables for selection fields,
	// since huh.NewSelect requires *string while config uses custom types.
	var (
		templateType         = string(cfg.Template)
		templateSource       string
		templateSearchChoice string
		vulkanUseVCPKG       = true // padrão: VCPKG habilitado para Vulkan
		standard             = string(cfg.Standard)
		projectType          = string(cfg.ProjectType)
		layout               = string(cfg.Layout)
		pkgManager           = string(cfg.PackageManager)
		ide                  = string(cfg.IDE)
		clangFormatStyle     = string(cfg.ClangFormatStyle)
		debugAdapter         = string(cfg.DebugAdapter)
		archStyle            = "flat"
		archPatterns         []string
		headerOnly           = false
	)

	// Resolve --template up front so it can pre-select and skip the template
	// group below. A bad alias/source fails fast here instead of silently
	// falling through to an interactive prompt the user didn't ask for.
	templatePreset := initialTemplate != ""
	if templatePreset {
		switch strings.ToLower(initialTemplate) {
		case "blank":
			templateType = string(config.TemplateBlank)
		case "vulkan":
			templateType = string(config.TemplateVulkan)
		default:
			source, err := registry.ResolveSource(initialTemplate, false)
			if err != nil {
				return nil, err
			}
			templateType = string(config.TemplateCustom)
			templateSource = source
		}
	}

	// ── Group 0: Template Selection ─────────────────────────────────────────────────────
	groupTemplate := huh.NewGroup(
		huh.NewNote().
			Title("⚡ cpp-gen").
			Description("Gerador moderno de projetos C++\nEscolha um template para começar."),

		huh.NewSelect[string]().
			Title("Template do projeto").
			DescriptionFunc(func() string {
				if templateType == templateSearchValue {
					return fmt.Sprintf("Escolha entre %d template(s) cadastrado(s) localmente ou no registro do GitHub.", len(catalog.Entries))
				}
				return config.ProjectTemplate(templateType).Description()
			}, &templateType).
			Options(templateOptions(len(searchOptions) > 0)...).
			Value(&templateType),
	)

	// ── Group 0b: Custom Template Source ─────────────────────────────────────
	groupCustomSource := huh.NewGroup(
		huh.NewNote().
			Title("Template Custom").
			Description("Informe de onde buscar o template. O cpp-gen substitui as variáveis\ndo projeto (nome, versão, etc.) nos arquivos encontrados."),

		huh.NewInput().
			Title("Fonte do template").
			Description("Caminho local, owner/repo[/subdir][#ref], ou URL git (https/ssh).").
			Placeholder("owner/repo ou ./meu-template ou https://github.com/owner/repo.git").
			Value(&templateSource).
			Validate(validateTemplateSource),
	).WithHideFunc(func() bool {
		return templateType != string(config.TemplateCustom)
	})

	// ── Group 0c: Template Registry Search ───────────────────────────────────
	// Only meaningful when there's at least one entry to pick from — built
	// unconditionally, but simply never added to the form otherwise (see
	// "Procurar..." option handling in templateOptions).
	groupTemplateSearch := huh.NewGroup(
		huh.NewNote().
			Title("Procurar Templates").
			Description("Templates cadastrados localmente (~/.config/cpp-gen/templates.json)\ne no registro do GitHub."),

		huh.NewSelect[string]().
			Title("Escolha um template").
			DescriptionFunc(func() string {
				e, ok := sourceToEntry[templateSearchChoice]
				if !ok {
					return ""
				}
				desc := e.Description
				if len(e.Tags) > 0 {
					desc += "\nTags: " + strings.Join(e.Tags, ", ")
				}
				desc += "\nFonte: " + e.Source
				return desc
			}, &templateSearchChoice).
			Options(searchOptions...).
			Value(&templateSearchChoice),
	).WithHideFunc(func() bool {
		return templateType != templateSearchValue
	})

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
	).WithHideFunc(func() bool {
		return templateType == string(config.TemplateVulkan) ||
			templateType == string(config.TemplateCustom) ||
			templateType == templateSearchValue
	})

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
	).WithHideFunc(func() bool {
		return templateType == string(config.TemplateVulkan) ||
			templateType == string(config.TemplateCustom) ||
			templateType == templateSearchValue
	})

	// ── Group 3b: Architecture ────────────────────────────────────────────────
	groupArchitecture := huh.NewGroup(
		huh.NewNote().
			Title("Arquitetura do Projeto").
			Description("Define como o código será organizado e quais padrões estarão disponíveis\npara geração com 'cpp-gen generate'."),

		huh.NewSelect[string]().
			Title("Estilo de arquitetura").
			DescriptionFunc(func() string {
				return archStyleDescription(archStyle)
			}, &archStyle).
			Options(
				huh.NewOption("Flat        — sem organização por camada (padrão)", "flat"),
				huh.NewOption("Layered     — Clean Architecture (domain/app/infra/presentation)", "layered"),
				huh.NewOption("Modular     — por feature/módulo auto-contido", "modular"),
				huh.NewOption("Hexagonal   — Ports & Adapters (core/adapters)", "hexagonal"),
				huh.NewOption("MVC         — Model-View-Controller", "mvc"),
				huh.NewOption("MVVM        — Model-View-ViewModel", "mvvm"),
				huh.NewOption("ECS         — Entity-Component-System (jogos)", "ecs"),
			).
			Value(&archStyle),

		huh.NewMultiSelect[string]().
			Title("Padrões adicionais").
			Description("Padrões que estarão disponíveis como schematics em 'cpp-gen generate'.").
			Options(
				huh.NewOption("repository  — interface + implementação de acesso a dados", "repository"),
				huh.NewOption("service     — serviço de domínio com interface", "service"),
				huh.NewOption("command     — padrão Command (undo/redo, filas)", "command"),
				huh.NewOption("observer    — padrão Observer/Event", "observer"),
				huh.NewOption("factory     — padrão Factory", "factory"),
				huh.NewOption("singleton   — instância única thread-safe", "singleton"),
				huh.NewOption("module      — módulo auto-contido", "module"),
			).
			Value(&archPatterns),

		huh.NewConfirm().
			Title("Header-only layout?").
			Description("Mantém .hpp e .cpp no mesmo diretório, sem pasta include/ separada.\nAtivado automaticamente para layouts Merged e Flat.").
			Value(&headerOnly),
	).WithHideFunc(func() bool {
		return templateType == string(config.TemplateVulkan) ||
			templateType == string(config.TemplateCustom) ||
			templateType == templateSearchValue
	})

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
	).WithHideFunc(func() bool {
		return templateType == string(config.TemplateVulkan) ||
			templateType == string(config.TemplateCustom) ||
			templateType == templateSearchValue
	})

	// ── Group 4b: Vulkan — Package Manager ───────────────────────────────────
	groupVulkanPackages := huh.NewGroup(
		huh.NewNote().
			Title("Dependências Vulkan").
			Description("O template Vulkan usa FetchContent para vk-bootstrap automaticamente.\nOpcionalmente, use VCPKG para gerenciar as demais dependências\n(glm, glfw3, imgui, VulkanMemoryAllocator, stb, tinyobjloader)."),

		huh.NewConfirm().
			Title("Usar VCPKG?").
			Description("Gera vcpkg.json com todas as dependências do projeto.\nSem VCPKG, o CMake buscará os pacotes no sistema (apt, brew, etc.).").
			Affirmative("Sim — gerar vcpkg.json").
			Negative("Não — usar pacotes do sistema").
			Value(&vulkanUseVCPKG),
	).WithHideFunc(func() bool {
		return templateType != string(config.TemplateVulkan)
	})

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

		huh.NewSelect[string]().
			Title("Debug Adapter").
			DescriptionFunc(func() string {
				return config.DebugAdapter(debugAdapter).Description()
			}, &debugAdapter).
			Options(
				huh.NewOption("LLDB   — macOS / Linux + Clang  (recomendado)", string(config.DebugAdapterLLDB)),
				huh.NewOption("GDB    — Linux + GCC", string(config.DebugAdapterGDB)),
				huh.NewOption("Ambos  — gera configurações para LLDB e GDB", string(config.DebugAdapterBoth)),
			).
			Value(&debugAdapter),

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
	var groups []*huh.Group
	if !templatePreset {
		groups = append(groups, groupTemplate, groupCustomSource)
		if len(searchOptions) > 0 {
			groups = append(groups, groupTemplateSearch)
		}
	}
	groups = append(groups,
		groupIdentity,
		groupTechnical,
		groupLayout,
		groupArchitecture,
		groupPackages,
		groupVulkanPackages,
		groupIDE,
	)

	form := huh.NewForm(groups...).
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
	if templateType == string(config.TemplateVulkan) {
		if vulkanUseVCPKG {
			cfg.PackageManager = config.PkgVCPKG
		} else {
			cfg.PackageManager = config.PkgNone
		}
	} else {
		cfg.PackageManager = config.PackageManager(pkgManager)
	}
	cfg.IDE = config.IDE(ide)
	cfg.ClangFormatStyle = config.ClangFormatStyle(clangFormatStyle)
	cfg.DebugAdapter = config.DebugAdapter(debugAdapter)
	switch templateType {
	case string(config.TemplateCustom):
		cfg.Template = config.TemplateCustom
		cfg.TemplateSource = strings.TrimSpace(templateSource)
	case templateSearchValue:
		cfg.Template = config.TemplateCustom
		cfg.TemplateSource = templateSearchChoice
	default:
		cfg.Template = config.ProjectTemplate(templateType)
	}
	cfg.ArchStyle = localconfig.ArchStyle(archStyle)
	cfg.ArchPatterns = archPatterns
	cfg.HeaderOnly = headerOnly

	return cfg, nil
}

// archStyleDescription returns a human-readable description for the given
// architecture style, shown dynamically in the selection field.
func archStyleDescription(style string) string {
	switch style {
	case "layered":
		return "Camadas: domain/ → application/ → infrastructure/ → presentation/\nIdeal para aplicações com lógica de negócio complexa."
	case "modular":
		return "Cada feature é um módulo auto-contido em modules/<feature>/\nIdeal para apps com domínios bem separados."
	case "hexagonal":
		return "core/ports/ define interfaces; adapters/ provê implementações.\nIdeal para alta testabilidade e troca de dependências."
	case "mvc":
		return "models/ + views/ + controllers/\nIdeal para aplicações desktop com UI clara."
	case "mvvm":
		return "models/ + viewmodels/ + views/\nIdeal para apps desktop com data binding."
	case "ecs":
		return "systems/ + components/ + core/\nIdeal para jogos e simulações."
	default:
		return "Todos os arquivos em src/ e include/ sem organização por camada.\nIdeal para projetos pequenos ou bibliotecas."
	}
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

// validateTemplateSource validates the custom template source field,
// reusing customtemplate.ParseSource so syntax errors surface immediately
// in the form instead of only when the generator tries to fetch it.
func validateTemplateSource(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return errors.New("informe uma fonte de template")
	}
	_, err := customtemplate.ParseSource(raw)
	return err
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
	theme.Focused.Title = TitleStyle
	theme.Focused.Description = MutedStyle

	// Selected / active field
	theme.Focused.SelectedOption = lipglossColor(colorAccent)
	theme.Focused.UnselectedOption = MutedStyle

	// Selection cursor
	theme.Focused.SelectSelector = InfoStyle

	// Confirmation buttons
	theme.Focused.FocusedButton = lipglossColor(colorSuccess)
	theme.Focused.BlurredButton = MutedStyle

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
