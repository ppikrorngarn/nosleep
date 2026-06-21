package main

import (
	"fmt"
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
