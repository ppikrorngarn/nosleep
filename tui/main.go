package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	STATUS_ENABLED  = "Sleep is ENABLED (nosleep is OFF)."
	STATUS_DISABLED = "Sleep is DISABLED (nosleep is ON)."
)

// SleepState represents the current sleep state
type SleepState string

const (
	StateNormal  SleepState = "normal"
	StateAwake   SleepState = "awake"
	StateUnknown SleepState = "unknown"
)

// Phase represents the current UI phase
type Phase string

const (
	PhaseIdle    Phase = "idle"
	PhaseWorking Phase = "working"
	PhaseHelp    Phase = "help"
)

type model struct {
	sleepState   SleepState
	phase        Phase
	showHelp     bool
	errorMessage string
	spinner      spinner.Model
	helpContent  string
	width        int
	height       int
}

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

func initialModel() model {
	return model{
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

func (m model) Init() tea.Cmd {
	return tea.Batch(
		checkStatus(),
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
					return m, toggleSleep("on")
				case StateAwake:
					return m, toggleSleep("off")
				default:
					// For unknown state or other cases, we can't toggle
					return m, nil
				}
			case "s":
				return m, runSetup()
			case "h":
				m.showHelp = !m.showHelp
				if m.showHelp {
					return m, getHelp()
				}
				return m, nil
			case "r":
				return m, checkStatus()
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
		return m, checkStatus()

	case setupDoneMsg:
		m.phase = PhaseIdle
		if msg.err != nil {
			m.errorMessage = fmt.Sprintf("Setup failed: %v", msg.err)
			return m, tea.Tick(5*time.Second, func(_ time.Time) tea.Msg {
				return clearErrorMsg{}
			})
		}
		m.errorMessage = ""
		return m, checkStatus()

	case errorMsg:
		m.phase = PhaseIdle
		m.errorMessage = msg.message
		return m, tea.Tick(5*time.Second, func(_ time.Time) tea.Msg {
			return clearErrorMsg{}
		})

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

func (m model) View() string {
	var s strings.Builder

	// Header
	s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("  NoSleep · macOS"))
	s.WriteString("\n")

	if m.showHelp {
		// Help view
		s.WriteString("\n")
		s.WriteString(m.helpContent)
		s.WriteString("\n\n")
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("  Press H to return to dashboard"))
	} else {
		// Main dashboard view
		s.WriteString("\n")

		// Hero status card
		card := m.createStatusCard()
		s.WriteString(card)
		s.WriteString("\n")

		// Controls
		s.WriteString(m.createControls())

		// Error message if present
		if m.errorMessage != "" {
			s.WriteString("\n")
			s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#ff6666")).Render("  " + m.errorMessage))
		}
	}

	return s.String()
}

func (m model) createStatusCard() string {
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
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#888888")).
		Width(40).
		Align(lipgloss.Center).
		Padding(1, 2).
		Background(bgColor).
		Foreground(textColor)

	cardContent := fmt.Sprintf("%s  %s\n%s", icon, title, description)
	return cardStyle.Render(cardContent)
}

func (m model) createControls() string {
	var controls strings.Builder

	// Base controls
	baseControls := []string{
		"Space Toggle sleep",
		"s     Setup passwordless mode",
		"h     Help",
		"r     Refresh",
		"q     Quit",
	}

	// Add battery warning if asleep
	if m.sleepState == StateAwake {
		controls.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#d78700")).Render("  ⚠ Battery drain risk while disabled"))
		controls.WriteString("\n")
	}

	// Add controls
	for _, ctrl := range baseControls {
		controls.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("  " + ctrl))
		controls.WriteString("\n")
	}

	return controls.String()
}

func checkStatus() tea.Cmd {
	return func() tea.Msg {
		result, _ := runNosleepScript("status")
		var state SleepState

		if strings.Contains(result, "DISABLED") {
			state = StateAwake
		} else if strings.Contains(result, "ENABLED") {
			state = StateNormal
		} else {
			state = StateUnknown
		}

		return statusMsg{state: state}
	}
}

func getHelp() tea.Cmd {
	return func() tea.Msg {
		_, _ = runNosleepScript("help")
		// We don't need to return the result here, just update the state
		return statusMsg{state: StateUnknown}
	}
}

func toggleSleep(action string) tea.Cmd {
	return func() tea.Msg {
		_, err := runNosleepScript(action)
		if err != nil {
			return errorMsg{message: fmt.Sprintf("Failed to %s sleep: %v", action, err)}
		}

		// For on/off, we want to refresh status after completion
		return workDoneMsg{}
	}
}

// runSetup suspends the TUI and runs the script's setup command on the real
// terminal so that sudo can prompt for the password interactively.
func runSetup() tea.Cmd {
	scriptPath := filepath.Join(filepath.Dir(getBinaryPath()), "..", "nosleep.sh")
	c := exec.Command(scriptPath, "setup")
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return setupDoneMsg{err: err}
	})
}

func runNosleepScript(args ...string) (string, error) {
	scriptPath := filepath.Join(filepath.Dir(getBinaryPath()), "..", "nosleep.sh")

	cmd := exec.Command(scriptPath, args...)
	output, err := cmd.CombinedOutput()

	return string(output), err
}

func getBinaryPath() string {
	binaryPath, _ := os.Executable()
	return binaryPath
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
