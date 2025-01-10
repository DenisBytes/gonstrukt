package steps

import (
	"github.com/DenisBytes/gonstrukt/internal/config"
	"github.com/DenisBytes/gonstrukt/internal/tui/components"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// DatabaseStep allows selection of the database type
type DatabaseStep struct {
	selector components.Selector
	complete bool
	value    config.DatabaseType
}

// NewDatabaseStep creates a new database selection step
func NewDatabaseStep() *DatabaseStep {
	options := []components.Option{
		{
			Label:       "PostgreSQL",
			Value:       string(config.DBPostgres),
			Description: "Recommended for production. Full ACID compliance, JSONB support.",
		},
		{
			Label:       "MySQL",
			Value:       string(config.DBMySQL),
			Description: "Popular relational database with wide ecosystem support.",
		},
		{
			Label:       "SQLite",
			Value:       string(config.DBSQLite),
			Description: "Lightweight, embedded database. Good for development/small apps.",
		},
		{
			Label:       "MongoDB",
			Value:       string(config.DBMongoDB),
			Description: "Document database. Flexible schema, good for rapid development.",
		},
		{
			Label:       "ArangoDB",
			Value:       string(config.DBArangoDB),
			Description: "Multi-model database. Graph, document, and key-value in one.",
		},
	}

	return &DatabaseStep{
		selector: components.NewSelector("", options),
	}
}

// Init initializes the step
func (s *DatabaseStep) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (s *DatabaseStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		s.value = config.DatabaseType(s.selector.SelectedValue())
		return s, func() tea.Msg {
			return StepCompleteMsg{
				StepName: "database",
				Value:    s.value,
			}
		}
	}

	return s, cmd
}

// View renders the step
func (s *DatabaseStep) View() string {
	return s.selector.View()
}

// Title returns the step title
func (s *DatabaseStep) Title() string {
	return "Select Database"
}

// Description returns the step description
func (s *DatabaseStep) Description() string {
	return "Choose the database for the auth service"
}

// IsComplete returns true if the step is complete
func (s *DatabaseStep) IsComplete() bool {
	return s.complete
}

// Value returns the selected value
func (s *DatabaseStep) Value() interface{} {
	return s.value
}

// Reset resets the step
func (s *DatabaseStep) Reset() {
	s.selector.Reset()
	s.complete = false
	s.value = ""
}
