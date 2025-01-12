package steps

import (
	"github.com/DenisBytes/gonstrukt/internal/tui/components"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// AuthCacheStep allows enabling/disabling auth response caching
type AuthCacheStep struct {
	selector components.Selector
	complete bool
	value    bool
}

// NewAuthCacheStep creates a new auth cache selection step
func NewAuthCacheStep() *AuthCacheStep {
	options := []components.Option{
		{
			Label:       "Enable Auth Caching",
			Value:       "true",
			Description: "Cache auth service responses in Redis for faster session validation",
		},
		{
			Label:       "Disable",
			Value:       "false",
			Description: "Always forward auth requests to auth service. More consistent but slower.",
		},
	}

	return &AuthCacheStep{
		selector: components.NewSelector("", options),
	}
}

// Init initializes the step
func (s *AuthCacheStep) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (s *AuthCacheStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				StepName: "auth_cache",
				Value:    s.value,
			}
		}
	}

	return s, cmd
}

// View renders the step
func (s *AuthCacheStep) View() string {
	return s.selector.View()
}

// Title returns the step title
func (s *AuthCacheStep) Title() string {
	return "Enable Auth Response Caching?"
}

// Description returns the step description
func (s *AuthCacheStep) Description() string {
	return "Cache authentication responses in Redis for improved performance"
}

// IsComplete returns true if the step is complete
func (s *AuthCacheStep) IsComplete() bool {
	return s.complete
}

// Value returns the selected value
func (s *AuthCacheStep) Value() interface{} {
	return s.value
}

// Reset resets the step
func (s *AuthCacheStep) Reset() {
	s.selector.Reset()
	s.complete = false
	s.value = false
}
