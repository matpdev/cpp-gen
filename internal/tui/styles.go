// Package tui contains all terminal user interface components
// used by cpp-gen, including interactive forms and visual styles.
package tui

import "github.com/charmbracelet/lipgloss"

// ─────────────────────────────────────────────────────────────────────────────
// Color palette
// ─────────────────────────────────────────────────────────────────────────────

const (
	colorPrimary   = lipgloss.Color("#7C3AED") // Primary purple
	colorSecondary = lipgloss.Color("#A78BFA") // Light purple
	colorAccent    = lipgloss.Color("#06B6D4") // Cyan
	colorSuccess   = lipgloss.Color("#10B981") // Green
	colorWarning   = lipgloss.Color("#F59E0B") // Yellow
	colorError     = lipgloss.Color("#EF4444") // Red
	colorMuted     = lipgloss.Color("#6B7280") // Gray
	colorText      = lipgloss.Color("#F9FAFB") // Soft white
	colorBorder    = lipgloss.Color("#374151") // Dark gray
)

// ─────────────────────────────────────────────────────────────────────────────
// Text styles
// ─────────────────────────────────────────────────────────────────────────────

// TitleStyle is the style used in the main title / application banner.
var TitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(colorPrimary)

// SubtitleStyle is the style used in subtitles and banner descriptions.
var SubtitleStyle = lipgloss.NewStyle().
	Foreground(colorSecondary)

// SectionStyle is the style for section headers within lists.
var SectionStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(colorAccent)

// MutedStyle is used for secondary text and low-priority information.
var MutedStyle = lipgloss.NewStyle().
	Foreground(colorMuted)

// BoldStyle applies bold to text without changing the color.
var BoldStyle = lipgloss.NewStyle().
	Bold(true)

// ─────────────────────────────────────────────────────────────────────────────
// Status / feedback styles
// ─────────────────────────────────────────────────────────────────────────────

// SuccessStyle is used for success messages (e.g. "✓ Project created").
var SuccessStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(colorSuccess)

// WarningStyle is used for warning messages (e.g. "⚠ Git not found").
var WarningStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(colorWarning)

// ErrorStyle is used for critical error messages.
var ErrorStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(colorError)

// InfoStyle is used for neutral informational messages.
var InfoStyle = lipgloss.NewStyle().
	Foreground(colorAccent)

// ─────────────────────────────────────────────────────────────────────────────
// Line prefix styles (status icons)
// ─────────────────────────────────────────────────────────────────────────────

// CheckMark returns the stylized success symbol.
func CheckMark() string {
	return SuccessStyle.Render("✓")
}

// CrossMark returns the stylized error symbol.
func CrossMark() string {
	return ErrorStyle.Render("✗")
}

// Arrow returns a stylized arrow used as a step prefix.
func Arrow() string {
	return InfoStyle.Render("→")
}

// Bullet returns a stylized bullet for lists.
func Bullet() string {
	return MutedStyle.Render("•")
}

// ─────────────────────────────────────────────────────────────────────────────
// Layout / container styles
// ─────────────────────────────────────────────────────────────────────────────

// BoxStyle is the style for bordered boxes with rounded corners,
// used for summaries and information panels.
var BoxStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(colorBorder).
	Padding(1, 2).
	MarginTop(1)

// SummaryHeaderStyle is the style for the final summary block header.
var SummaryHeaderStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(colorPrimary).
	MarginBottom(1)

// KeyStyle is the style for keys in key:value pairs in the summary.
var KeyStyle = lipgloss.NewStyle().
	Foreground(colorSecondary).
	Width(20)

// ValueStyle is the style for values in key:value pairs in the summary.
var ValueStyle = lipgloss.NewStyle().
	Foreground(colorText)

// ─────────────────────────────────────────────────────────────────────────────
// High-level formatting functions
// ─────────────────────────────────────────────────────────────────────────────

// FormatStep formats a generation step progress line,
// displaying a success/error icon and the step description.
//
// Example:
//
//	✓ Folder structure created
//	✗ Failed to initialize Git
func FormatStep(success bool, message string) string {
	if success {
		return CheckMark() + "  " + message
	}
	return CrossMark() + "  " + ErrorStyle.Render(message)
}

// FormatKeyValue formats an aligned key/value pair for the project summary.
//
// Example:
//
//	Name                 my-project
//	C++ Standard         C++20
func FormatKeyValue(key, value string) string {
	return KeyStyle.Render(key) + ValueStyle.Render(value)
}

// FormatSection formats a section header with a visual separator.
//
// Example:
//
//	── Settings ─────────────────────────
func FormatSection(title string) string {
	return SectionStyle.Render("── " + title + " " + repeatRune('─', 35-len(title)))
}

// repeatRune returns a string with the rune repeated n times (minimum 0).
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
