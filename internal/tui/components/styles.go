package components

import "github.com/charmbracelet/lipgloss"

// Colors for the TUI
var (
	Primary    = lipgloss.Color("#7C3AED") // Purple
	Secondary  = lipgloss.Color("#10B981") // Green
	Accent     = lipgloss.Color("#F59E0B") // Amber
	Muted      = lipgloss.Color("#6B7280") // Gray
	Error      = lipgloss.Color("#EF4444") // Red
	Success    = lipgloss.Color("#10B981") // Green
	Background = lipgloss.Color("#1F2937") // Dark gray
	Foreground = lipgloss.Color("#F9FAFB") // Light gray
)

// Styles holds all styled components for the wizard
type Styles struct {
	// Title styles
	Title       lipgloss.Style
	Subtitle    lipgloss.Style
	Description lipgloss.Style

	// Option styles
	Option         lipgloss.Style
	SelectedOption lipgloss.Style
	DisabledOption lipgloss.Style

	// Input styles
	Input      lipgloss.Style
	InputFocus lipgloss.Style
	Cursor     lipgloss.Style

	// Status styles
	Success lipgloss.Style
	Error   lipgloss.Style
	Warning lipgloss.Style
	Info    lipgloss.Style

	// Layout styles
	Container   lipgloss.Style
	StepCounter lipgloss.Style
	Help        lipgloss.Style

	// Progress styles
	Progress        lipgloss.Style
	ProgressSpinner lipgloss.Style

	// Summary styles
	SummaryLabel lipgloss.Style
	SummaryValue lipgloss.Style
}

// DefaultStyles returns the default style configuration
func DefaultStyles() Styles {
	return Styles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary).
			MarginBottom(1),

		Subtitle: lipgloss.NewStyle().
			Foreground(Foreground).
			MarginBottom(1),

		Description: lipgloss.NewStyle().
			Foreground(Muted).
			Italic(true).
			MarginBottom(2),

		Option: lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(Foreground),

		SelectedOption: lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(Primary).
			Bold(true),

		DisabledOption: lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(Muted).
			Strikethrough(true),

		Input: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Muted).
			Padding(0, 1),

		InputFocus: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(0, 1),

		Cursor: lipgloss.NewStyle().
			Foreground(Primary),

		Success: lipgloss.NewStyle().
			Foreground(Success).
			Bold(true),

		Error: lipgloss.NewStyle().
			Foreground(Error).
			Bold(true),

		Warning: lipgloss.NewStyle().
			Foreground(Accent),

		Info: lipgloss.NewStyle().
			Foreground(Secondary),

		Container: lipgloss.NewStyle().
			Padding(1, 2),

		StepCounter: lipgloss.NewStyle().
			Foreground(Muted).
			MarginBottom(1),

		Help: lipgloss.NewStyle().
			Foreground(Muted).
			MarginTop(2),

		Progress: lipgloss.NewStyle().
			Foreground(Primary),

		ProgressSpinner: lipgloss.NewStyle().
			Foreground(Primary),

		SummaryLabel: lipgloss.NewStyle().
			Foreground(Muted).
			Width(20),

		SummaryValue: lipgloss.NewStyle().
			Foreground(Foreground).
			Bold(true),
	}
}

// Indicator returns the selection indicator
func Indicator() string {
	return lipgloss.NewStyle().
		Foreground(Primary).
		SetString("→ ").
		String()
}

// Checkbox returns a checkbox string
func Checkbox(checked bool) string {
	if checked {
		return lipgloss.NewStyle().
			Foreground(Success).
			SetString("[✓] ").
			String()
	}
	return lipgloss.NewStyle().
		Foreground(Muted).
		SetString("[ ] ").
		String()
}

// Bullet returns a bullet point string
func Bullet() string {
	return lipgloss.NewStyle().
		Foreground(Primary).
		SetString("• ").
		String()
}

// Divider returns a horizontal divider
func Divider(width int) string {
	return lipgloss.NewStyle().
		Foreground(Muted).
		SetString(repeat("─", width)).
		String()
}

func repeat(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
