package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Command functions

// checkStatus queries the bash script for the current system sleep state
func checkStatus(client *Client) tea.Cmd {
	return func() tea.Msg {
		state, err := client.Status()
		if err != nil {
			return errorMsg{message: fmt.Sprintf("Status check failed: %v", err)}
		}

		return statusMsg{state: state}
	}
}

// getHelp emits a dummy status message to trigger the UI to show the help screen
func getHelp() tea.Cmd {
	return func() tea.Msg {
		// Help doesn't need script output — just toggle the UI state
		return statusMsg{state: StateUnknown}
	}
}

// toggleSleep executes either 'on' or 'off' command via the client
func toggleSleep(client *Client, action string) tea.Cmd {
	return func() tea.Msg {
		if client.NeedsSetup() {
			return errorMsg{message: "Setup required — press S to enable passwordless mode"}
		}

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
