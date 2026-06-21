package main

import (
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
