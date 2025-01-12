package steps

import (
	"github.com/DenisBytes/gonstrukt/internal/tui/components"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// GatewayFeaturesStep allows enabling gateway features (cache, rate limiting) for auth service
type GatewayFeaturesStep struct {
	selector components.Selector
	complete bool
	value    bool
}

// NewGatewayFeaturesStep creates a new gateway features selection step
func NewGatewayFeaturesStep() *GatewayFeaturesStep {
	options := []components.Option{
		{
			Label:       "Enable",
			Value:       "true",
			Description: "Add caching and rate limiting to your auth service",
		},
		{
			Label:       "Skip",
			Value:       "false",
			Description: "Auth service without gateway features (can be added later)",
		},
	}

	return &GatewayFeaturesStep{
		selector: components.NewSelector("", options),
	}
}

// Init initializes the step
func (s *GatewayFeaturesStep) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (s *GatewayFeaturesStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				StepName: "gateway_features",
				Value:    s.value,
			}
		}
	}

	return s, cmd
}

// View renders the step
func (s *GatewayFeaturesStep) View() string {
	return s.selector.View()
}

// Title returns the step title
func (s *GatewayFeaturesStep) Title() string {
	return "Enable Gateway Features?"
}

// Description returns the step description
func (s *GatewayFeaturesStep) Description() string {
	return "Add caching and rate limiting capabilities to your auth service"
}

// IsComplete returns true if the step is complete
func (s *GatewayFeaturesStep) IsComplete() bool {
	return s.complete
}

// Value returns the selected value
func (s *GatewayFeaturesStep) Value() any {
	return s.value
}

// Reset resets the step
func (s *GatewayFeaturesStep) Reset() {
	s.selector.Reset()
	s.complete = false
	s.value = false
}
