package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	STATUS_ENABLED  = "Sleep is ENABLED (nosleep is OFF)."
	STATUS_DISABLED = "Sleep is DISABLED (nosleep is ON)."
)

type model struct {
	choices   []string
	cursor    int
	selected  int
	status    string
	showResult bool
	resultMessage string
	lastAction string
}

type statusMsg string
type resultMsg struct {
	action string
	result string
}

func initialModel() model {
	return model{
		choices: []string{
			"Turn NoSleep ON",
			"Turn NoSleep OFF",
			"Check Status",
			"Setup Passwordless Mode",
			"Help",
			"Quit",
		},
		cursor:   0,
		selected: 0,
		status:   "Checking status...",
		showResult: false,
	}
}

func (m model) Init() tea.Cmd {
	return checkStatus()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.showResult {
			if msg.String() == "enter" || msg.String() == "q" {
				m.showResult = false
				return m, checkStatus()
			}
			return m, nil
		}

		switch msg.String() {
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			m.selected = m.cursor
			return m, handleSelection(m.selected)
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			return m, checkStatus()
		}

	case statusMsg:
		m.status = string(msg)
	case resultMsg:
		m.showResult = true
		m.lastAction = msg.action
		m.resultMessage = msg.result
	}

	return m, nil
}

func (m model) View() string {
	var s strings.Builder

	if m.showResult {
		s.WriteString(fmt.Sprintf("  Action: %s\n\n", m.lastAction))
		s.WriteString(m.resultMessage)
		s.WriteString("\n\n  Press Enter to return to main menu or Q to quit")
	} else {
		s.WriteString("  NoSleep for macOS\n\n")
		s.WriteString(fmt.Sprintf("  Status: %s\n\n", m.status))
		s.WriteString("  Select an action:\n\n")

		for i, choice := range m.choices {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			s.WriteString(fmt.Sprintf("  %s %s\n", cursor, choice))
		}

		s.WriteString("\n  Controls: ↑↓ Enter | R: refresh | Q: quit")
	}

	return s.String()
}

func checkStatus() tea.Cmd {
	return func() tea.Msg {
		result, _ := runNosleepScript("status")
		status := strings.TrimSpace(result)
		if status == "" {
			status = STATUS_ENABLED
		}
		return statusMsg(status)
	}
}

func handleSelection(choice int) tea.Cmd {
	switch choice {
	case 0: // Turn NoSleep ON
		return func() tea.Msg {
			result, err := runNosleepScript("on")
			if err != nil {
				return resultMsg{
					action: "Turn NoSleep ON",
					result: fmt.Sprintf("Error: %v\n\n%s", err, result),
				}
			}
			return resultMsg{
				action: "Turn NoSleep ON",
				result: fmt.Sprintf("Success!\n\n%s", result),
			}
		}
	case 1: // Turn NoSleep OFF
		return func() tea.Msg {
			result, err := runNosleepScript("off")
			if err != nil {
				return resultMsg{
					action: "Turn NoSleep OFF",
					result: fmt.Sprintf("Error: %v\n\n%s", err, result),
				}
			}
			return resultMsg{
				action: "Turn NoSleep OFF",
				result: fmt.Sprintf("Success!\n\n%s", result),
			}
		}
	case 2: // Check Status
		return func() tea.Msg {
			result, _ := runNosleepScript("status")
			return resultMsg{
				action: "Check Status",
				result: fmt.Sprintf("Current status:\n\n%s", result),
			}
		}
	case 3: // Setup Passwordless Mode
		return func() tea.Msg {
			result, err := runNosleepScript("setup")
			if err != nil {
				return resultMsg{
					action: "Setup Passwordless Mode",
					result: fmt.Sprintf("Error: %v\n\n%s", err, result),
				}
			}
			return resultMsg{
				action: "Setup Passwordless Mode",
				result: fmt.Sprintf("Success!\n\n%s", result),
			}
		}
	case 4: // Help
		return func() tea.Msg {
			result, _ := runNosleepScript("help")
			return resultMsg{
				action: "Help",
				result: result,
			}
		}
	case 5: // Quit
		return tea.Quit
	}
	return nil
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
