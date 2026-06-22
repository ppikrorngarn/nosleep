package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

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
