package steps

import (
	"github.com/DenisBytes/gonstrukt/internal/tui/components"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// AnalyticsSelection holds the selected analytics options
type AnalyticsSelection struct {
	PostHog bool
	Sentry  bool
}

// AnalyticsStep allows selection of analytics/monitoring tools
type AnalyticsStep struct {
	selector components.MultiSelector
	complete bool
	value    AnalyticsSelection
}

// NewAnalyticsStep creates a new analytics selection step
func NewAnalyticsStep() *AnalyticsStep {
	options := []components.Option{
		{
			Label:       "PostHog",
			Value:       "posthog",
			Description: "Product analytics, session replay, and feature flags",
		},
		{
			Label:       "Sentry",
			Value:       "sentry",
			Description: "Error tracking, performance monitoring, and session replay",
		},
	}

	return &AnalyticsStep{
		selector: components.NewMultiSelector("", options),
	}
}

// Init initializes the step
func (s *AnalyticsStep) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (s *AnalyticsStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			return s, func() tea.Msg { return StepBackMsg{} }
		}
	}

	var cmd tea.Cmd
	s.selector, cmd = s.selector.Update(msg)

	if s.selector.IsConfirmed() {
		s.complete = true
		values := s.selector.SelectedValues()
		s.value = AnalyticsSelection{}
		for _, v := range values {
			switch v {
			case "posthog":
				s.value.PostHog = true
			case "sentry":
				s.value.Sentry = true
			}
		}
		return s, func() tea.Msg {
			return StepCompleteMsg{
				StepName: "analytics",
				Value:    s.value,
			}
		}
	}

	return s, cmd
}

// View renders the step
func (s *AnalyticsStep) View() string {
	return s.selector.View()
}

// Title returns the step title
func (s *AnalyticsStep) Title() string {
	return "Select Analytics & Monitoring"
}

// Description returns the step description
func (s *AnalyticsStep) Description() string {
	return "Add optional analytics and error tracking (press enter to skip)"
}

// IsComplete returns true if the step is complete
func (s *AnalyticsStep) IsComplete() bool {
	return s.complete
}

// Value returns the selected value
func (s *AnalyticsStep) Value() interface{} {
	return s.value
}

// Reset resets the step
func (s *AnalyticsStep) Reset() {
	s.selector.Reset()
	s.complete = false
	s.value = AnalyticsSelection{}
}
