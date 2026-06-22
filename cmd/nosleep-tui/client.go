package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

//go:generate cp ../../cli/nosleep.sh ./nosleep.sh
//go:embed nosleep.sh
var nosleepScript []byte

// SleepState represents the current sleep state.
type SleepState string

const (
	StateNormal  SleepState = "normal"
	StateAwake   SleepState = "awake"
	StateUnknown SleepState = "unknown"
)

// statusResponse is the JSON output from "nosleep.sh status --json".
type statusResponse struct {
	State        string `json:"state"`
	Disablesleep int    `json:"disablesleep"`
}

// actionResponse is the JSON output from "nosleep.sh on/off/setup --json".
type actionResponse struct {
	Ok     bool   `json:"ok"`
	Action string `json:"action"`
	User   string `json:"user,omitempty"`
}

// Client wraps the nosleep.sh script and provides Go-native methods.
type Client struct {
	scriptPath string
}

// NewClient creates a Client by extracting the embedded nosleep.sh to a temp file.
func NewClient() (*Client, error) {
	f, err := os.CreateTemp("", "nosleep-*.sh")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file for script: %w", err)
	}
	scriptPath := f.Name()

	// We must use f.Chmod() instead of relying on WriteFile permissions
	// because CreateTemp creates the file with 0600 permissions first.
	if err := f.Chmod(0755); err != nil {
		f.Close()
		os.Remove(scriptPath)
		return nil, fmt.Errorf("failed to make script executable: %w", err)
	}

	if err := os.WriteFile(scriptPath, nosleepScript, 0755); err != nil {
		os.Remove(scriptPath)
		return nil, fmt.Errorf("failed to write embedded script: %w", err)
	}
	f.Close()

	return &Client{scriptPath: scriptPath}, nil
}

// Cleanup removes the temporary script file.
func (c *Client) Cleanup() {
	if c.scriptPath != "" {
		os.Remove(c.scriptPath)
	}
}

// Status queries the current sleep state and returns a SleepState.
func (c *Client) Status() (SleepState, error) {
	out, err := c.run("status", "--json")
	if err != nil {
		return StateUnknown, fmt.Errorf("status check failed: %w", err)
	}

	var resp statusResponse
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		return StateUnknown, fmt.Errorf("failed to parse status JSON: %w", err)
	}

	switch resp.State {
	case "awake":
		return StateAwake, nil
	case "normal":
		return StateNormal, nil
	default:
		return StateUnknown, nil
	}
}

// On disables system sleep (runs "nosleep.sh on").
func (c *Client) On() error {
	out, err := c.run("on", "--json")
	if err != nil {
		return fmt.Errorf("failed to disable sleep: %w", err)
	}
	var resp actionResponse
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		return fmt.Errorf("failed to parse on JSON: %w", err)
	}
	if !resp.Ok {
		return fmt.Errorf("on command reported failure")
	}
	return nil
}

// Off re-enables system sleep (runs "nosleep.sh off").
func (c *Client) Off() error {
	out, err := c.run("off", "--json")
	if err != nil {
		return fmt.Errorf("failed to enable sleep: %w", err)
	}
	var resp actionResponse
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		return fmt.Errorf("failed to parse off JSON: %w", err)
	}
	if !resp.Ok {
		return fmt.Errorf("off command reported failure")
	}
	return nil
}

// SetupCommand returns an *exec.Cmd for the setup command.
// This is not run via Client.run because setup needs an interactive
// terminal for the sudo password prompt — the TUI uses tea.ExecProcess.
func (c *Client) SetupCommand() *exec.Cmd {
	return exec.Command(c.scriptPath, "setup")
}

// run executes the nosleep.sh script with the given arguments and returns
// its stdout output.
func (c *Client) run(args ...string) (string, error) {
	cmd := exec.Command(c.scriptPath, args...)
	out, err := cmd.Output()
	return string(out), err
}
