package steps

import (
	"github.com/DenisBytes/gonstrukt/internal/config"
	"github.com/DenisBytes/gonstrukt/internal/tui/components"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// GDPRStep allows selection of GDPR compliance features
type GDPRStep struct {
	selector components.MultiSelector
	complete bool
	value    []config.GDPRFeature
}

// NewGDPRStep creates a new GDPR feature selection step
func NewGDPRStep() *GDPRStep {
	options := []components.Option{
		{
			Label:       "Consent Management",
			Value:       string(config.GDPRConsent),
			Description: "Versioned consent records with history tracking (GDPR Article 7)",
		},
		{
			Label:       "Data Export",
			Value:       string(config.GDPRDataExport),
			Description: "Export all user data in JSON format (GDPR Article 20 - Data Portability)",
		},
		{
			Label:       "Account Deletion",
			Value:       string(config.GDPRDataDeletion),
			Description: "Soft-delete with PII anonymization (GDPR Article 17 - Right to Erasure)",
		},
		{
			Label:       "Processing Logs",
			Value:       string(config.GDPRProcessingLogs),
			Description: "Audit trail of all data processing with legal basis (GDPR Article 30)",
		},
	}

	return &GDPRStep{
		selector: components.NewMultiSelector("", options),
	}
}

// Init initializes the step
func (s *GDPRStep) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (s *GDPRStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		s.value = make([]config.GDPRFeature, len(values))
		for i, v := range values {
			s.value[i] = config.GDPRFeature(v)
		}
		return s, func() tea.Msg {
			return StepCompleteMsg{
				StepName: "gdpr_features",
				Value:    s.value,
			}
		}
	}

	return s, cmd
}

// View renders the step
func (s *GDPRStep) View() string {
	return s.selector.View()
}

// Title returns the step title
func (s *GDPRStep) Title() string {
	return "Select GDPR Compliance Features"
}

// Description returns the step description
func (s *GDPRStep) Description() string {
	return "Choose which GDPR features to implement (optional, press enter to skip)"
}

// IsComplete returns true if the step is complete
func (s *GDPRStep) IsComplete() bool {
	return s.complete
}

// Value returns the selected value
func (s *GDPRStep) Value() interface{} {
	return s.value
}

// Reset resets the step
func (s *GDPRStep) Reset() {
	s.selector.Reset()
	s.complete = false
	s.value = nil
}
