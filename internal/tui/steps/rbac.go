package steps

import (
	"github.com/DenisBytes/gonstrukt/internal/tui/components"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// RBACStep allows enabling/disabling Casbin RBAC
type RBACStep struct {
	selector components.Selector
	complete bool
	value    bool
}

// NewRBACStep creates a new RBAC selection step
func NewRBACStep() *RBACStep {
	options := []components.Option{
		{
			Label:       "Enable RBAC",
			Value:       "true",
			Description: "Add Casbin-based role-based access control with configurable policies",
		},
		{
			Label:       "Disable",
			Value:       "false",
			Description: "Basic authentication without fine-grained permissions. Can be added later.",
		},
	}

	return &RBACStep{
		selector: components.NewSelector("", options),
	}
}

// Init initializes the step
func (s *RBACStep) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (s *RBACStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				StepName: "rbac",
				Value:    s.value,
			}
		}
	}

	return s, cmd
}

// View renders the step
func (s *RBACStep) View() string {
	return s.selector.View()
}

// Title returns the step title
func (s *RBACStep) Title() string {
	return "Enable Role-Based Access Control?"
}

// Description returns the step description
func (s *RBACStep) Description() string {
	return "Add Casbin RBAC for fine-grained permission management"
}

// IsComplete returns true if the step is complete
func (s *RBACStep) IsComplete() bool {
	return s.complete
}

// Value returns the selected value
func (s *RBACStep) Value() interface{} {
	return s.value
}

// Reset resets the step
func (s *RBACStep) Reset() {
	s.selector.Reset()
	s.complete = false
	s.value = false
}
