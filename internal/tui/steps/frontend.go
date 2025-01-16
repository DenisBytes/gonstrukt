package steps

import (
	"github.com/DenisBytes/gonstrukt/internal/config"
	"github.com/DenisBytes/gonstrukt/internal/tui/components"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// FrontendStep allows selection of frontend types (web, mobile, or both)
type FrontendStep struct {
	selector components.MultiSelector
	complete bool
	value    []config.FrontendType
}

// NewFrontendStep creates a new frontend selection step
func NewFrontendStep() *FrontendStep {
	options := []components.Option{
		{
			Label:       "Web",
			Value:       string(config.FrontendWeb),
			Description: "Web frontend with React, Next.js, or TanStack Start",
		},
		{
			Label:       "Mobile",
			Value:       string(config.FrontendMobile),
			Description: "Mobile app with React Native Expo",
		},
	}

	return &FrontendStep{
		selector: components.NewMultiSelector("", options),
	}
}

// Init initializes the step
func (s *FrontendStep) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (s *FrontendStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		s.value = make([]config.FrontendType, len(values))
		for i, v := range values {
			s.value[i] = config.FrontendType(v)
		}
		return s, func() tea.Msg {
			return StepCompleteMsg{
				StepName: "frontend",
				Value:    s.value,
			}
		}
	}

	return s, cmd
}

// View renders the step
func (s *FrontendStep) View() string {
	return s.selector.View()
}

// Title returns the step title
func (s *FrontendStep) Title() string {
	return "Frontend"
}

// Description returns the step description
func (s *FrontendStep) Description() string {
	return "Select frontends for your auth service (optional, press enter to skip)"
}

// IsComplete returns true if the step is complete
func (s *FrontendStep) IsComplete() bool {
	return s.complete
}

// Value returns the selected value
func (s *FrontendStep) Value() interface{} {
	return s.value
}

// Reset resets the step
func (s *FrontendStep) Reset() {
	s.selector.Reset()
	s.complete = false
	s.value = nil
}
