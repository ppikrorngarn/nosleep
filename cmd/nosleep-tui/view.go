package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	textSubtle = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	textError  = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff6666"))
	textWarn   = lipgloss.NewStyle().Foreground(lipgloss.Color("#d78700"))

	cardBaseStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#888888")).
			Width(40).
			Align(lipgloss.Center).
			Padding(1, 2)
)

// View renders the TUI interface based on the current model state.
func (m model) View() string {
	var s strings.Builder

	// Header
	s.WriteString(textSubtle.Render("  NoSleep · macOS"))
	s.WriteString("\n")

	if m.showHelp {
		// Help view
		s.WriteString("\n")
		s.WriteString(m.helpContent)
		s.WriteString("\n\n")
		s.WriteString(textSubtle.Render("  Press H to return to dashboard"))
	} else {
		// Main dashboard view
		s.WriteString("\n")

		// Hero status card
		card := createStatusCard(m)
		s.WriteString(card)
		s.WriteString("\n")

		// Controls
		s.WriteString(createControls(m))

		// Error message if present
		if m.errorMessage != "" {
			s.WriteString("\n")
			s.WriteString(textError.Render("  " + m.errorMessage))
		}
	}

	return s.String()
}

// createStatusCard generates the main hero visual showing whether the
// Mac is currently awake or sleeping.
func createStatusCard(m model) string {
	var title, description, icon string
	var bgColor, textColor lipgloss.Color

	// If we're working, show spinner instead of state info
	if m.phase == PhaseWorking {
		title = "Working..."
		description = "Please wait..."
		icon = m.spinner.View()
		bgColor = lipgloss.Color("#585858")
		textColor = lipgloss.Color("#ffffff")
	} else {
		// Handle different sleep states
		switch m.sleepState {
		case StateAwake:
			title = "AWAKE"
			description = "Your Mac will not sleep"
			icon = "☕"
			bgColor = lipgloss.Color("#d78700")
			textColor = lipgloss.Color("#ffffff")
		case StateNormal:
			title = "SLEEPING"
			description = "Your Mac can sleep normally"
			icon = "😴"
			bgColor = lipgloss.Color("#585858")
			textColor = lipgloss.Color("#ffffff")
		default:
			title = "UNKNOWN"
			description = "Checking status..."
			icon = "❓"
			bgColor = lipgloss.Color("#585858")
			textColor = lipgloss.Color("#ffffff")
		}
	}

	// Create the card with appropriate styling
	cardStyle := cardBaseStyle.Copy().
		Background(bgColor).
		Foreground(textColor)

	cardContent := fmt.Sprintf("%s  %s\n%s", icon, title, description)
	return cardStyle.Render(cardContent)
}

// createControls returns the list of available keyboard shortcuts
// and optionally renders a battery warning if sleep is disabled.
func createControls(m model) string {
	var controls strings.Builder

	// Base controls
	baseControls := []string{
		"Space Toggle sleep",
		"s     Setup passwordless mode",
		"h     Help",
		"r     Refresh",
		"q     Quit",
	}

	// Always reserve a line for the battery warning so the layout height
	// stays constant when toggling between ON/OFF states.
	if m.sleepState == StateAwake {
		controls.WriteString(textWarn.Render("  ⚠ Battery drain risk while disabled"))
	} else {
		controls.WriteString(" ")
	}
	controls.WriteString("\n")

	// Add controls
	for _, ctrl := range baseControls {
		controls.WriteString(textSubtle.Render("  " + ctrl))
		controls.WriteString("\n")
	}

	return controls.String()
}
