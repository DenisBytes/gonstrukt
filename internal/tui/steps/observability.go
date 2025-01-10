package steps

import (
	"github.com/DenisBytes/gonstrukt/internal/tui/components"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// ObservabilityStep allows enabling/disabling OTLP observability
type ObservabilityStep struct {
	selector components.Selector
	complete bool
	value    bool
}

// NewObservabilityStep creates a new observability selection step
func NewObservabilityStep() *ObservabilityStep {
	options := []components.Option{
		{
			Label:       "Enable OTLP",
			Value:       "true",
			Description: "Enable OpenTelemetry traces, metrics, and structured logging.",
		},
		{
			Label:       "Disable",
			Value:       "false",
			Description: "Basic logging only. Can be added later if needed.",
		},
	}

	return &ObservabilityStep{
		selector: components.NewSelector("", options),
	}
}

// Init initializes the step
func (s *ObservabilityStep) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (s *ObservabilityStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				StepName: "observability",
				Value:    s.value,
			}
		}
	}

	return s, cmd
}

// View renders the step
func (s *ObservabilityStep) View() string {
	return s.selector.View()
}

// Title returns the step title
func (s *ObservabilityStep) Title() string {
	return "Enable Observability?"
}

// Description returns the step description
func (s *ObservabilityStep) Description() string {
	return "Enable OpenTelemetry tracing and metrics"
}

// IsComplete returns true if the step is complete
func (s *ObservabilityStep) IsComplete() bool {
	return s.complete
}

// Value returns the selected value
func (s *ObservabilityStep) Value() interface{} {
	return s.value
}

// Reset resets the step
func (s *ObservabilityStep) Reset() {
	s.selector.Reset()
	s.complete = false
	s.value = false
}
