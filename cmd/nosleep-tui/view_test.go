package main

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/spinner"
)

// testModel returns a model with sensible defaults for view tests.
// Client is nil because View() never calls it.
func testModel(state SleepState, width, height int) model {
	return model{
		client:       nil,
		sleepState:   state,
		phase:        PhaseIdle,
		showHelp:     false,
		errorMessage: "",
		spinner:      spinner.New(spinner.WithSpinner(spinner.Line)),
		helpContent:  helpText,
		width:        width,
		height:       height,
	}
}

// TestView_PadsToTerminalHeight checks that View() fills the full terminal
// height when width/height are known, so old content is overwritten on resize.
func TestView_PadsToTerminalHeight(t *testing.T) {
	m := testModel(StateNormal, 80, 40)
	out := m.View()

	lines := strings.Split(out, "\n")
	// lipgloss.Place pads with blank lines to reach the target height.
	// The output should have at least `height` lines.
	if len(lines) < m.height {
		t.Errorf("expected at least %d lines to fill terminal height, got %d", m.height, len(lines))
	}
}

// TestView_NoPaddingBeforeFirstWindowSizeMsg checks that View() returns
// unpadded content when height == 0 (before the first WindowSizeMsg).
func TestView_NoPaddingBeforeFirstWindowSizeMsg(t *testing.T) {
	m := testModel(StateNormal, 0, 0)
	out := m.View()

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	// Without padding, the view is short — just the header + card + controls.
	// Certainly fewer than 40 lines.
	if len(lines) >= 40 {
		t.Errorf("expected short unpadded view, got %d lines", len(lines))
	}
}

// TestView_WarningShownWhenAwake checks the battery drain warning is present
// when sleep is disabled.
func TestView_WarningShownWhenAwake(t *testing.T) {
	m := testModel(StateAwake, 0, 0)
	out := m.View()

	if !strings.Contains(out, "Battery drain risk") {
		t.Error("expected battery drain warning when state is StateAwake, got none")
	}
}

// TestView_WarningAbsentWhenSleeping checks the battery drain warning is not
// shown when sleep is enabled, but the blank placeholder line is still rendered
// to keep layout height constant.
func TestView_WarningAbsentWhenSleeping(t *testing.T) {
	m := testModel(StateNormal, 0, 0)
	out := m.View()

	if strings.Contains(out, "Battery drain risk") {
		t.Error("did not expect battery drain warning when state is StateNormal")
	}

	controls := createControls(m)
	lines := strings.Split(controls, "\n")
	// First line is the reserved warning slot (blank space), followed by the
	// base controls. There should be more than just the controls lines.
	if len(lines) < 2 {
		t.Error("expected reserved blank line before controls, layout looks wrong")
	}
	// First line should be the blank placeholder (a single space), not empty.
	if strings.TrimSpace(lines[0]) != "" {
		t.Errorf("expected blank placeholder as first controls line, got %q", lines[0])
	}
}

// TestView_ControlsAlwaysPresent checks that all key bindings are rendered
// in both sleep states.
func TestView_ControlsAlwaysPresent(t *testing.T) {
	expectedKeys := []string{"Space", "s", "h", "esc", "r", "q"}

	for _, state := range []SleepState{StateAwake, StateNormal, StateUnknown} {
		m := testModel(state, 0, 0)
		out := m.View()
		for _, key := range expectedKeys {
			if !strings.Contains(out, key) {
				t.Errorf("state=%s: expected key %q in controls, not found", state, key)
			}
		}
	}
}
