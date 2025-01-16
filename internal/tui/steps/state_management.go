package steps

import (
	"github.com/DenisBytes/gonstrukt/internal/config"
	"github.com/DenisBytes/gonstrukt/internal/tui/components"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// StateManagementStep allows selection of state management approach
type StateManagementStep struct {
	selector components.Selector
	complete bool
	value    config.StateManagement
}

// NewStateManagementStep creates a new state management selection step
func NewStateManagementStep() *StateManagementStep {
	options := []components.Option{
		{
			Label:       "TanStack Query + Zustand",
			Value:       string(config.StateMgmtTanStack),
			Description: "Data fetching with TanStack Query, client state with Zustand",
		},
		{
			Label:       "Redux Toolkit + RTK Query",
			Value:       string(config.StateMgmtRedux),
			Description: "Centralized state with Redux Toolkit and RTK Query",
		},
	}

	return &StateManagementStep{
		selector: components.NewSelector("", options),
	}
}

// Init initializes the step
func (s *StateManagementStep) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (s *StateManagementStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		s.value = config.StateManagement(s.selector.SelectedValue())
		return s, func() tea.Msg {
			return StepCompleteMsg{
				StepName: "state_management",
				Value:    s.value,
			}
		}
	}

	return s, cmd
}

// View renders the step
func (s *StateManagementStep) View() string {
	return s.selector.View()
}

// Title returns the step title
func (s *StateManagementStep) Title() string {
	return "State Management"
}

// Description returns the step description
func (s *StateManagementStep) Description() string {
	return "Choose the state management approach for your frontend"
}

// IsComplete returns true if the step is complete
func (s *StateManagementStep) IsComplete() bool {
	return s.complete
}

// Value returns the selected value
func (s *StateManagementStep) Value() interface{} {
	return s.value
}

// Reset resets the step
func (s *StateManagementStep) Reset() {
	s.selector.Reset()
	s.complete = false
	s.value = ""
}
