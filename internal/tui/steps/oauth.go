package steps

import (
	"github.com/DenisBytes/gonstrukt/internal/config"
	"github.com/DenisBytes/gonstrukt/internal/tui/components"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// OAuthStep allows selection of OAuth providers
type OAuthStep struct {
	selector components.MultiSelector
	complete bool
	value    []config.OAuthProvider
}

// NewOAuthStep creates a new OAuth provider selection step
func NewOAuthStep() *OAuthStep {
	options := []components.Option{
		{
			Label:       "Google",
			Value:       string(config.OAuthGoogle),
			Description: "Sign in with Google using OAuth 2.0",
		},
		{
			Label:       "Microsoft",
			Value:       string(config.OAuthMicrosoft),
			Description: "Sign in with Microsoft/Azure AD",
		},
		{
			Label:       "Apple",
			Value:       string(config.OAuthApple),
			Description: "Sign in with Apple for iOS/macOS integration",
		},
	}

	return &OAuthStep{
		selector: components.NewMultiSelector("", options),
	}
}

// Init initializes the step
func (s *OAuthStep) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (s *OAuthStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		s.value = make([]config.OAuthProvider, len(values))
		for i, v := range values {
			s.value[i] = config.OAuthProvider(v)
		}
		return s, func() tea.Msg {
			return StepCompleteMsg{
				StepName: "oauth_providers",
				Value:    s.value,
			}
		}
	}

	return s, cmd
}

// View renders the step
func (s *OAuthStep) View() string {
	return s.selector.View()
}

// Title returns the step title
func (s *OAuthStep) Title() string {
	return "Select OAuth Providers"
}

// Description returns the step description
func (s *OAuthStep) Description() string {
	return "Choose which OAuth providers to integrate (optional, press enter to skip)"
}

// IsComplete returns true if the step is complete
func (s *OAuthStep) IsComplete() bool {
	return s.complete
}

// Value returns the selected value
func (s *OAuthStep) Value() interface{} {
	return s.value
}

// Reset resets the step
func (s *OAuthStep) Reset() {
	s.selector.Reset()
	s.complete = false
	s.value = nil
}
