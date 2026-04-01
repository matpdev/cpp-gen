// Package tui contém todos os componentes de interface de usuário do terminal
// utilizados pelo cpp-gen, incluindo formulários interativos e estilos visuais.
package tui

import "github.com/charmbracelet/lipgloss"

// ─────────────────────────────────────────────────────────────────────────────
// Paleta de cores
// ─────────────────────────────────────────────────────────────────────────────

const (
	colorPrimary   = lipgloss.Color("#7C3AED") // Roxo principal
	colorSecondary = lipgloss.Color("#A78BFA") // Roxo claro
	colorAccent    = lipgloss.Color("#06B6D4") // Ciano
	colorSuccess   = lipgloss.Color("#10B981") // Verde
	colorWarning   = lipgloss.Color("#F59E0B") // Amarelo
	colorError     = lipgloss.Color("#EF4444") // Vermelho
	colorMuted     = lipgloss.Color("#6B7280") // Cinza
	colorText      = lipgloss.Color("#F9FAFB") // Branco suave
	colorBorder    = lipgloss.Color("#374151") // Cinza escuro
)

// ─────────────────────────────────────────────────────────────────────────────
// Estilos de texto
// ─────────────────────────────────────────────────────────────────────────────

// TitleStyle é o estilo usado no título principal / banner da aplicação.
var TitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(colorPrimary)

// SubtitleStyle é o estilo usado em subtítulos e descrições do banner.
var SubtitleStyle = lipgloss.NewStyle().
	Foreground(colorSecondary)

// SectionStyle é o estilo para cabeçalhos de seção dentro de listas.
var SectionStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(colorAccent)

// MutedStyle é usado para textos secundários e informações de baixa prioridade.
var MutedStyle = lipgloss.NewStyle().
	Foreground(colorMuted)

// BoldStyle aplica negrito ao texto sem alterar a cor.
var BoldStyle = lipgloss.NewStyle().
	Bold(true)

// ─────────────────────────────────────────────────────────────────────────────
// Estilos de status / feedback
// ─────────────────────────────────────────────────────────────────────────────

// SuccessStyle é usado para mensagens de sucesso (ex: "✓ Projeto criado").
var SuccessStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(colorSuccess)

// WarningStyle é usado para mensagens de aviso (ex: "⚠ Git não encontrado").
var WarningStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(colorWarning)

// ErrorStyle é usado para mensagens de erro crítico.
var ErrorStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(colorError)

// InfoStyle é usado para mensagens informativas neutras.
var InfoStyle = lipgloss.NewStyle().
	Foreground(colorAccent)

// ─────────────────────────────────────────────────────────────────────────────
// Estilos de prefixo de linha (ícones de status)
// ─────────────────────────────────────────────────────────────────────────────

// CheckMark retorna o símbolo de sucesso estilizado.
func CheckMark() string {
	return SuccessStyle.Render("✓")
}

// CrossMark retorna o símbolo de erro estilizado.
func CrossMark() string {
	return ErrorStyle.Render("✗")
}

// Arrow retorna uma seta estilizada usada como prefixo de etapas.
func Arrow() string {
	return InfoStyle.Render("→")
}

// Bullet retorna um marcador estilizado para listas.
func Bullet() string {
	return MutedStyle.Render("•")
}

// ─────────────────────────────────────────────────────────────────────────────
// Estilos de layout / contêiner
// ─────────────────────────────────────────────────────────────────────────────

// BoxStyle é o estilo para caixas delimitadas com borda arredondada,
// usado para resumos e painéis de informação.
var BoxStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(colorBorder).
	Padding(1, 2).
	MarginTop(1)

// SummaryHeaderStyle é o estilo para o cabeçalho do bloco de resumo final.
var SummaryHeaderStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(colorPrimary).
	MarginBottom(1)

// KeyStyle é o estilo para chaves em pares chave: valor no resumo.
var KeyStyle = lipgloss.NewStyle().
	Foreground(colorSecondary).
	Width(20)

// ValueStyle é o estilo para valores em pares chave: valor no resumo.
var ValueStyle = lipgloss.NewStyle().
	Foreground(colorText)

// ─────────────────────────────────────────────────────────────────────────────
// Funções de formatação de alto nível
// ─────────────────────────────────────────────────────────────────────────────

// FormatStep formata uma linha de progresso de etapa de geração,
// exibindo um ícone de sucesso/erro e a descrição da etapa.
//
// Exemplo:
//
//	✓ Estrutura de pastas criada
//	✗ Falha ao inicializar Git
func FormatStep(success bool, message string) string {
	if success {
		return CheckMark() + "  " + message
	}
	return CrossMark() + "  " + ErrorStyle.Render(message)
}

// FormatKeyValue formata um par chave/valor alinhado para o resumo do projeto.
//
// Exemplo:
//
//	Nome                 meu-projeto
//	Padrão C++           C++20
func FormatKeyValue(key, value string) string {
	return KeyStyle.Render(key) + ValueStyle.Render(value)
}

// FormatSection formata um cabeçalho de seção com separador visual.
//
// Exemplo:
//
//	── Configurações ─────────────────────────
func FormatSection(title string) string {
	return SectionStyle.Render("── " + title + " " + repeatRune('─', 35-len(title)))
}

// repeatRune retorna uma string com o rune repetido n vezes (mínimo 0).
func repeatRune(r rune, n int) string {
	if n <= 0 {
		return ""
	}
	result := make([]rune, n)
	for i := range result {
		result[i] = r
	}
	return string(result)
}
