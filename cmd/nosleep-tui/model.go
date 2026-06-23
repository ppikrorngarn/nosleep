package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// Phase represents the current UI phase
type Phase string

const (
	PhaseIdle    Phase = "idle"
	PhaseWorking Phase = "working"
	PhaseHelp    Phase = "help"
)

const helpText = `NoSleep prevents your Mac from sleeping by toggling the system
disablesleep setting via pmset. This works with lid closed and on battery.

Key Bindings
  Space   Toggle sleep on/off
  s       Setup passwordless mode (sudoers rule)
  h       Toggle this help screen
  r       Refresh current sleep state
  q       Quit the application

Safety
  Your Mac can remain awake with lid closed and while not on power.
  This can increase battery drain and cause the laptop to become hot,
  especially in a bag or poorly ventilated space. Use intentionally.

Note: First time users should press 's' to enable passwordless mode.`

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

// Message types used by Bubble Tea to update application state
type (
	statusMsg     struct{ state SleepState }
	workDoneMsg   struct{}
	setupDoneMsg  struct{ err error }
	errorMsg      struct{ message string }
	clearErrorMsg struct{}
)

// Model initialization
func initialModel(client *Client) model {
	return model{
		client:       client,
		sleepState:   StateUnknown,
		phase:        PhaseIdle,
		showHelp:     false,
		errorMessage: "",
		spinner:      spinner.New(spinner.WithSpinner(spinner.Line)),
		helpContent:  helpText,
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
		// Allow esc to close help at any time
		if msg.String() == "esc" && m.showHelp {
			m.showHelp = false
			return m, nil
		}

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
				// h opens help (only if not already open)
				if !m.showHelp {
					m.showHelp = true
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
