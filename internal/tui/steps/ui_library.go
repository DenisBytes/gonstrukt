package steps

import (
	"github.com/DenisBytes/gonstrukt/internal/config"
	"github.com/DenisBytes/gonstrukt/internal/tui/components"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// UILibraryStep allows selection of UI component library
type UILibraryStep struct {
	selector components.Selector
	complete bool
	value    config.UILibrary
}

// NewUILibraryStep creates a new UI library selection step
func NewUILibraryStep() *UILibraryStep {
	options := []components.Option{
		{
			Label:       "ShadcnUI",
			Value:       string(config.UILibShadcn),
			Description: "Copy-paste components with Tailwind CSS and Radix primitives",
		},
		{
			Label:       "BaseUI",
			Value:       string(config.UILibBaseUI),
			Description: "Uber's component library with Styletron CSS-in-JS",
		},
	}

	return &UILibraryStep{
		selector: components.NewSelector("", options),
	}
}

// Init initializes the step
func (s *UILibraryStep) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (s *UILibraryStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		s.value = config.UILibrary(s.selector.SelectedValue())
		return s, func() tea.Msg {
			return StepCompleteMsg{
				StepName: "ui_library",
				Value:    s.value,
			}
		}
	}

	return s, cmd
}

// View renders the step
func (s *UILibraryStep) View() string {
	return s.selector.View()
}

// Title returns the step title
func (s *UILibraryStep) Title() string {
	return "UI Library"
}

// Description returns the step description
func (s *UILibraryStep) Description() string {
	return "Choose the UI component library for your frontend"
}

// IsComplete returns true if the step is complete
func (s *UILibraryStep) IsComplete() bool {
	return s.complete
}

// Value returns the selected value
func (s *UILibraryStep) Value() interface{} {
	return s.value
}

// Reset resets the step
func (s *UILibraryStep) Reset() {
	s.selector.Reset()
	s.complete = false
	s.value = ""
}
