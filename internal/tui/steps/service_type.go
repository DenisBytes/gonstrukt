package steps

import (
	"github.com/DenisBytes/gonstrukt/internal/config"
	"github.com/DenisBytes/gonstrukt/internal/tui/components"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// ServiceTypeStep allows selection of the service type to generate
type ServiceTypeStep struct {
	selector components.Selector
	complete bool
	value    config.ServiceType
}

// NewServiceTypeStep creates a new service type selection step
func NewServiceTypeStep() *ServiceTypeStep {
	options := []components.Option{
		{
			Label:       "Gateway",
			Value:       string(config.ServiceGateway),
			Description: "API gateway with caching, rate limiting, and routing",
		},
		{
			Label:       "Auth Service",
			Value:       string(config.ServiceAuth),
			Description: "Authentication service with GDPR compliance, OAuth, MFA, and RBAC",
		},
		{
			Label:       "Both (Monorepo)",
			Value:       string(config.ServiceBoth),
			Description: "Generate both services in a monorepo structure with go.work",
		},
	}

	return &ServiceTypeStep{
		selector: components.NewSelector("", options),
	}
}

// Init initializes the step
func (s *ServiceTypeStep) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (s *ServiceTypeStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		s.value = config.ServiceType(s.selector.SelectedValue())
		return s, func() tea.Msg {
			return StepCompleteMsg{
				StepName: "service_type",
				Value:    s.value,
			}
		}
	}

	return s, cmd
}

// View renders the step
func (s *ServiceTypeStep) View() string {
	return s.selector.View()
}

// Title returns the step title
func (s *ServiceTypeStep) Title() string {
	return "Select Service Type"
}

// Description returns the step description
func (s *ServiceTypeStep) Description() string {
	return "Choose what type of service to generate"
}

// IsComplete returns true if the step is complete
func (s *ServiceTypeStep) IsComplete() bool {
	return s.complete
}

// Value returns the selected value
func (s *ServiceTypeStep) Value() interface{} {
	return s.value
}

// Reset resets the step
func (s *ServiceTypeStep) Reset() {
	s.selector.Reset()
	s.complete = false
	s.value = ""
}
