package steps

import (
	"fmt"
	"regexp"

	"github.com/DenisBytes/gonstrukt/internal/tui/components"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// ProjectNameStep allows input of the module name
type ProjectNameStep struct {
	input    components.Input
	complete bool
	value    string
	styles   components.Styles
}

// NewProjectNameStep creates a new project name input step
func NewProjectNameStep() *ProjectNameStep {
	validator := func(value string) error {
		if value == "" {
			return fmt.Errorf("module name is required")
		}
		// Validate Go module name format
		pattern := `^[a-zA-Z0-9][a-zA-Z0-9._-]*(/[a-zA-Z0-9][a-zA-Z0-9._-]*)*$`
		matched, _ := regexp.MatchString(pattern, value)
		if !matched {
			return fmt.Errorf("invalid module format (e.g., github.com/user/project)")
		}
		return nil
	}

	return &ProjectNameStep{
		input:  components.NewInput("", "github.com/username/project", validator),
		styles: components.DefaultStyles(),
	}
}

// Init initializes the step
func (s *ProjectNameStep) Init() tea.Cmd {
	return s.input.Init()
}

// Update handles messages
func (s *ProjectNameStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			return s, func() tea.Msg { return StepBackMsg{} }
		}
	}

	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)

	if s.input.IsSubmitted() {
		s.complete = true
		s.value = s.input.Value()
		return s, func() tea.Msg {
			return StepCompleteMsg{
				StepName: "project_name",
				Value:    s.value,
			}
		}
	}

	return s, cmd
}

// View renders the step
func (s *ProjectNameStep) View() string {
	return s.input.View()
}

// Title returns the step title
func (s *ProjectNameStep) Title() string {
	return "Enter Module Name"
}

// Description returns the step description
func (s *ProjectNameStep) Description() string {
	return "Enter your Go module path"
}

// IsComplete returns true if the step is complete
func (s *ProjectNameStep) IsComplete() bool {
	return s.complete
}

// Value returns the entered value
func (s *ProjectNameStep) Value() interface{} {
	return s.value
}

// Reset resets the step
func (s *ProjectNameStep) Reset() {
	s.input.Reset()
	s.complete = false
	s.value = ""
}
