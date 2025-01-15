package steps

import (
	"github.com/DenisBytes/gonstrukt/internal/config"
	"github.com/DenisBytes/gonstrukt/internal/tui/components"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// EmailServiceStep allows selection of email service provider
type EmailServiceStep struct {
	selector components.Selector
	complete bool
	value    config.EmailService
}

// NewEmailServiceStep creates a new email service selection step
func NewEmailServiceStep() *EmailServiceStep {
	options := []components.Option{
		{
			Label:       "AWS SES",
			Value:       string(config.EmailSES),
			Description: "Amazon Simple Email Service - scalable, cost-effective for production",
		},
		{
			Label:       "SMTP",
			Value:       string(config.EmailSMTP),
			Description: "Generic SMTP - works with any email provider (SendGrid, Mailgun, etc.)",
		},
	}

	return &EmailServiceStep{
		selector: components.NewSelector("", options),
	}
}

// Init initializes the step
func (s *EmailServiceStep) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (s *EmailServiceStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		s.value = config.EmailService(s.selector.SelectedValue())
		return s, func() tea.Msg {
			return StepCompleteMsg{
				StepName: "email_service",
				Value:    s.value,
			}
		}
	}

	return s, cmd
}

// View renders the step
func (s *EmailServiceStep) View() string {
	return s.selector.View()
}

// Title returns the step title
func (s *EmailServiceStep) Title() string {
	return "Select Email Service"
}

// Description returns the step description
func (s *EmailServiceStep) Description() string {
	return "Choose the email service for verification, password reset, and notifications"
}

// IsComplete returns true if the step is complete
func (s *EmailServiceStep) IsComplete() bool {
	return s.complete
}

// Value returns the selected value
func (s *EmailServiceStep) Value() interface{} {
	return s.value
}

// Reset resets the step
func (s *EmailServiceStep) Reset() {
	s.selector.Reset()
	s.complete = false
	s.value = ""
}
