package steps

import (
	"github.com/DenisBytes/gonstrukt/internal/tui/components"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// MFAStep allows enabling/disabling MFA/TOTP support
type MFAStep struct {
	selector components.Selector
	complete bool
	value    bool
}

// NewMFAStep creates a new MFA selection step
func NewMFAStep() *MFAStep {
	options := []components.Option{
		{
			Label:       "Enable MFA",
			Value:       "true",
			Description: "Add TOTP-based multi-factor authentication with backup codes",
		},
		{
			Label:       "Disable",
			Value:       "false",
			Description: "Basic password authentication only. Can be added later if needed.",
		},
	}

	return &MFAStep{
		selector: components.NewSelector("", options),
	}
}

// Init initializes the step
func (s *MFAStep) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (s *MFAStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			return s, func() tea.Msg { return StepBackMsg{} }
		}
	}

	var cmd tea.Cmd
	s.selector, cmd = s.selector.Update(msg)

	if s.selector.IsSelected() {
		s.complete = true
		s.value = s.selector.SelectedValue() == "true"
		return s, func() tea.Msg {
			return StepCompleteMsg{
				StepName: "mfa",
				Value:    s.value,
			}
		}
	}

	return s, cmd
}

// View renders the step
func (s *MFAStep) View() string {
	return s.selector.View()
}

// Title returns the step title
func (s *MFAStep) Title() string {
	return "Enable Multi-Factor Authentication?"
}

// Description returns the step description
func (s *MFAStep) Description() string {
	return "Add TOTP-based MFA with authenticator apps and backup codes"
}

// IsComplete returns true if the step is complete
func (s *MFAStep) IsComplete() bool {
	return s.complete
}

// Value returns the selected value
func (s *MFAStep) Value() interface{} {
	return s.value
}

// Reset resets the step
func (s *MFAStep) Reset() {
	s.selector.Reset()
	s.complete = false
	s.value = false
}
