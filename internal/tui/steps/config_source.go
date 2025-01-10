package steps

import (
	"github.com/DenisBytes/gonstrukt/internal/config"
	"github.com/DenisBytes/gonstrukt/internal/tui/components"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// ConfigSourceStep allows selection of the configuration source
type ConfigSourceStep struct {
	selector components.Selector
	complete bool
	value    config.ConfigSource
}

// NewConfigSourceStep creates a new config source selection step
func NewConfigSourceStep() *ConfigSourceStep {
	options := []components.Option{
		{
			Label:       "YAML File",
			Value:       string(config.ConfigYAML),
			Description: "Local YAML configuration file. Simple setup for development.",
		},
		{
			Label:       "Environment Variables",
			Value:       string(config.ConfigEnv),
			Description: "12-factor app style. Good for containers and cloud deployments.",
		},
		{
			Label:       "HashiCorp Vault",
			Value:       string(config.ConfigVault),
			Description: "Centralized secrets management. Recommended for production.",
		},
	}

	return &ConfigSourceStep{
		selector: components.NewSelector("", options),
	}
}

// Init initializes the step
func (s *ConfigSourceStep) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (s *ConfigSourceStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		s.value = config.ConfigSource(s.selector.SelectedValue())
		return s, func() tea.Msg {
			return StepCompleteMsg{
				StepName: "config_source",
				Value:    s.value,
			}
		}
	}

	return s, cmd
}

// View renders the step
func (s *ConfigSourceStep) View() string {
	return s.selector.View()
}

// Title returns the step title
func (s *ConfigSourceStep) Title() string {
	return "Select Configuration Source"
}

// Description returns the step description
func (s *ConfigSourceStep) Description() string {
	return "Choose where configuration will be loaded from"
}

// IsComplete returns true if the step is complete
func (s *ConfigSourceStep) IsComplete() bool {
	return s.complete
}

// Value returns the selected value
func (s *ConfigSourceStep) Value() interface{} {
	return s.value
}

// Reset resets the step
func (s *ConfigSourceStep) Reset() {
	s.selector.Reset()
	s.complete = false
	s.value = ""
}
