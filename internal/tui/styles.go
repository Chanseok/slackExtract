package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Color palette
	ColorGreen  = lipgloss.Color("42")  // Public channels
	ColorRed    = lipgloss.Color("196") // Private channels
	ColorYellow = lipgloss.Color("226") // DMs
	ColorGray   = lipgloss.Color("246") // Archived (lighter for dark theme visibility)
	ColorCyan   = lipgloss.Color("51")  // Selected highlight
	ColorWhite  = lipgloss.Color("231")
	ColorBlack  = lipgloss.Color("232") // Dark background

	// Base styles
	NormalStyle = lipgloss.NewStyle().
			Foreground(ColorWhite)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(ColorCyan).
			Bold(true)

	// Channel type styles
	PublicChannelStyle = lipgloss.NewStyle().
				Foreground(ColorGreen)

	PrivateChannelStyle = lipgloss.NewStyle().
				Foreground(ColorRed)

	DMStyle = lipgloss.NewStyle().
		Foreground(ColorYellow)

	ArchivedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("246")). // Lighter gray for better visibility
			Faint(true)

	// UI components
	HeaderStyle = lipgloss.NewStyle().
			Foreground(ColorCyan).
			Bold(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(ColorCyan)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorGray).
			Background(lipgloss.Color("235"))

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorGray).
			Italic(true)

	// Filter Menu Styles
	FilterMenuStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorCyan).
			Padding(1, 2).
			Background(ColorBlack)

	FilterTitleStyle = lipgloss.NewStyle().
				Foreground(ColorCyan).
				Bold(true).
				Underline(true).
				MarginBottom(1)

	FilterItemStyle = lipgloss.NewStyle().
			Foreground(ColorWhite)

	FilterSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorCyan).
				Bold(true)
)

// GetChannelIcon returns emoji icon based on channel type
func GetChannelIcon(isPrivate, isIM, isMpIM, isArchived bool) string {
	if isArchived {
		return "🗄️ "
	}
	if isIM {
		return "💬 "
	}
	if isMpIM {
		return "👥 "
	}
	if isPrivate {
		return "🔒 "
	}
	return "🔓 "
}

// GetChannelStyle returns appropriate style based on channel type
func GetChannelStyle(isPrivate, isIM, isMpIM, isArchived bool, isSelected bool) lipgloss.Style {
	var style lipgloss.Style

	if isArchived {
		style = ArchivedStyle
	} else if isIM || isMpIM {
		style = DMStyle
	} else if isPrivate {
		style = PrivateChannelStyle
	} else {
		style = PublicChannelStyle
	}

	if isSelected {
		style = style.Copy().Bold(true).Underline(true)
	}

	return style
}
