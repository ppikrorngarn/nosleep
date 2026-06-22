package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Phase represents the current UI phase
type Phase string

const (
	PhaseIdle    Phase = "idle"
	PhaseWorking Phase = "working"
	PhaseHelp    Phase = "help"
)

type model struct {
	client       *Client
	sleepState   SleepState
	phase        Phase
	showHelp     bool
	errorMessage string
	spinner      spinner.Model
	helpContent  string
	width        int
	height       int
}

// Message types
type statusMsg struct {
	state SleepState
}
type workDoneMsg struct{}
type setupDoneMsg struct {
	err error
}
type errorMsg struct {
	message string
}
type clearErrorMsg struct{}

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

// Helper functions for creating UI elements
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

// Command functions
func checkStatus(client *Client) tea.Cmd {
	return func() tea.Msg {
		state, err := client.Status()
		if err != nil {
			return errorMsg{message: fmt.Sprintf("Status check failed: %v", err)}
		}
		return statusMsg{state: state}
	}
}

func getHelp() tea.Cmd {
	return func() tea.Msg {
		// Help doesn't need script output — just toggle the UI state
		return statusMsg{state: StateUnknown}
	}
}

func toggleSleep(client *Client, action string) tea.Cmd {
	return func() tea.Msg {
		var err error
		switch action {
		case "on":
			err = client.On()
		case "off":
			err = client.Off()
		}
		if err != nil {
			return errorMsg{message: fmt.Sprintf("Failed to %s sleep: %v", action, err)}
		}

		// For on/off, we want to refresh status after completion
		return workDoneMsg{}
	}
}

// runSetup suspends the TUI and runs the script's setup command on the real
// terminal so that sudo can prompt for the password interactively.
func runSetup(client *Client) tea.Cmd {
	c := client.SetupCommand()
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return setupDoneMsg{err: err}
	})
}

// Model initialization
func initialModel(client *Client) model {
	return model{
		client:       client,
		sleepState:   StateUnknown,
		phase:        PhaseIdle,
		showHelp:     false,
		errorMessage: "",
		spinner:      spinner.New(spinner.WithSpinner(spinner.Line)),
		helpContent:  "",
		width:        0,
		height:       0,
	}
}

// Model methods
func (m model) Init() tea.Cmd {
	return tea.Batch(
		checkStatus(m.client),
		m.spinner.Tick,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle key presses when not in working phase
		if m.phase != PhaseWorking {
			switch msg.String() {
			case " ":
				// Toggle sleep state based on current state
				switch m.sleepState {
				case StateNormal:
					return m, toggleSleep(m.client, "on")
				case StateAwake:
					return m, toggleSleep(m.client, "off")
				default:
					// For unknown state or other cases, we can't toggle
					return m, nil
				}
			case "s":
				return m, runSetup(m.client)
			case "h":
				m.showHelp = !m.showHelp
				if m.showHelp {
					return m, getHelp()
				}
				return m, nil
			case "r":
				return m, checkStatus(m.client)
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		} else {
			// During working phase, only allow quitting
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		}

	case statusMsg:
		m.sleepState = msg.state
		m.phase = PhaseIdle
		m.errorMessage = ""
		return m, nil

	case workDoneMsg:
		m.phase = PhaseIdle
		m.errorMessage = ""
		return m, checkStatus(m.client)

	case setupDoneMsg:
		m.phase = PhaseIdle
		if msg.err != nil {
			return m.handleError(fmt.Sprintf("Setup failed: %v", msg.err))
		}
		m.errorMessage = ""
		return m, checkStatus(m.client)

	case errorMsg:
		m.phase = PhaseIdle
		return m.handleError(msg.message)

	case clearErrorMsg:
		m.errorMessage = ""
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	return m, nil
}

// Helper method for error handling
func (m model) handleError(message string) (tea.Model, tea.Cmd) {
	m.errorMessage = message
	return m, tea.Tick(5*time.Second, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

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

func main() {
	client, err := NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing client: %v\n", err)
		os.Exit(1)
	}
	defer client.Cleanup()

	p := tea.NewProgram(initialModel(client), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
