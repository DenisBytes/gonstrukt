package steps

import (
	"github.com/DenisBytes/gonstrukt/internal/config"
	"github.com/DenisBytes/gonstrukt/internal/tui/components"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// WebFrameworkStep allows selection of web framework
type WebFrameworkStep struct {
	selector components.Selector
	complete bool
	value    config.WebFramework
}

// NewWebFrameworkStep creates a new web framework selection step
func NewWebFrameworkStep() *WebFrameworkStep {
	options := []components.Option{
		{
			Label:       "React + Vite",
			Value:       string(config.FrameworkReact),
			Description: "React with Vite build tool and React Router",
		},
		// Next.js and TanStack Start templates are not yet implemented
		// {
		// 	Label:       "Next.js",
		// 	Value:       string(config.FrameworkNext),
		// 	Description: "Full-stack React framework with App Router",
		// },
		// {
		// 	Label:       "TanStack Start",
		// 	Value:       string(config.FrameworkTanStack),
		// 	Description: "Full-stack React framework with TanStack Router",
		// },
	}

	return &WebFrameworkStep{
		selector: components.NewSelector("", options),
	}
}

// Init initializes the step
func (s *WebFrameworkStep) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (s *WebFrameworkStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		s.value = config.WebFramework(s.selector.SelectedValue())
		return s, func() tea.Msg {
			return StepCompleteMsg{
				StepName: "web_framework",
				Value:    s.value,
			}
		}
	}

	return s, cmd
}

// View renders the step
func (s *WebFrameworkStep) View() string {
	return s.selector.View()
}

// Title returns the step title
func (s *WebFrameworkStep) Title() string {
	return "Web Framework"
}

// Description returns the step description
func (s *WebFrameworkStep) Description() string {
	return "Choose the web framework for your frontend"
}

// IsComplete returns true if the step is complete
func (s *WebFrameworkStep) IsComplete() bool {
	return s.complete
}

// Value returns the selected value
func (s *WebFrameworkStep) Value() interface{} {
	return s.value
}

// Reset resets the step
func (s *WebFrameworkStep) Reset() {
	s.selector.Reset()
	s.complete = false
	s.value = ""
}
