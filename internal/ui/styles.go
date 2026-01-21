package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	Primary   = lipgloss.Color("#7C3AED") // Purple
	Secondary = lipgloss.Color("#10B981") // Green
	Warning   = lipgloss.Color("#F59E0B") // Amber
	Error     = lipgloss.Color("#EF4444") // Red
	Muted     = lipgloss.Color("#6B7280") // Gray
	White     = lipgloss.Color("#FFFFFF")
	Black     = lipgloss.Color("#000000")

	// Title style
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary).
			MarginBottom(1)

	// Subtitle style
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(Muted).
			MarginBottom(1)

	// Menu item styles
	MenuItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	SelectedMenuItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(Primary).
				Bold(true)

	// Status styles
	RunningStyle = lipgloss.NewStyle().
			Foreground(Secondary).
			Bold(true)

	StoppedStyle = lipgloss.NewStyle().
			Foreground(Muted)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Error).
			Bold(true)

	// Table styles
	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(Primary).
				BorderBottom(true).
				BorderStyle(lipgloss.NormalBorder())

	TableCellStyle = lipgloss.NewStyle().
			PaddingRight(2)

	// Box styles
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(1, 2)

	// Help style
	HelpStyle = lipgloss.NewStyle().
			Foreground(Muted).
			MarginTop(1)

	// Progress bar styles
	ProgressStyle = lipgloss.NewStyle().
			Foreground(Secondary)

	ProgressEmptyStyle = lipgloss.NewStyle().
				Foreground(Muted)

	// Info box
	InfoBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Secondary).
			Padding(1, 2)

	// Warning box
	WarningBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Warning).
			Padding(1, 2)

	// Logo style
	LogoStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)
)

// Logo returns the budgie ASCII art logo
func Logo() string {
	return LogoStyle.Render(`
    ____            __      _
   / __ )__  ______/ /___ _(_)___
  / __  / / / / __  / __ '/ / _ \
 / /_/ / /_/ / /_/ / /_/ / /  __/
/_____/\__,_/\__,_/\__, /_/\___/
                  /____/
`)
}

// StatusIcon returns a styled status icon
func StatusIcon(status string) string {
	switch status {
	case "running":
		return RunningStyle.Render("●")
	case "stopped":
		return StoppedStyle.Render("○")
	case "failed", "error":
		return ErrorStyle.Render("✗")
	default:
		return StoppedStyle.Render("?")
	}
}

// FormatBytes formats bytes into human-readable format
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return lipgloss.NewStyle().Render(fmt.Sprintf("%d B", bytes))
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

