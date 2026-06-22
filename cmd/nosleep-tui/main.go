package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// main is the entry point for the NoSleep terminal UI.
// It extracts the embedded bash script to a temp file, sets up the
// Bubble Tea program, and cleans up the temp file upon exit.
func main() {
	client, err := NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing client: %v\n", err)
		os.Exit(1)
	}
	defer client.Cleanup()

	p := tea.NewProgram(initialModel(client), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
