package steps

import (
	"github.com/DenisBytes/gonstrukt/internal/tui/components"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// TenancyStep allows enabling/disabling multi-tenancy
type TenancyStep struct {
	selector components.Selector
	complete bool
	value    bool
}

// NewTenancyStep creates a new tenancy selection step
func NewTenancyStep() *TenancyStep {
	options := []components.Option{
		{
			Label:       "Enable Multi-Tenancy",
			Value:       "true",
			Description: "Auth-first tenancy: workspace selection, invitations, tenant switching",
		},
		{
			Label:       "Disable",
			Value:       "false",
			Description: "Single-tenant mode. Can be added later.",
		},
	}

	return &TenancyStep{
		selector: components.NewSelector("", options),
	}
}

func (s *TenancyStep) Init() tea.Cmd { return nil }

func (s *TenancyStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				StepName: "tenancy",
				Value:    s.value,
			}
		}
	}

	return s, cmd
}

func (s *TenancyStep) View() string        { return s.selector.View() }
func (s *TenancyStep) Title() string        { return "Enable Multi-Tenancy?" }
func (s *TenancyStep) Description() string  { return "Add auth-first multi-tenant workspaces with invitations and tenant switching" }
func (s *TenancyStep) IsComplete() bool     { return s.complete }
func (s *TenancyStep) Value() interface{}   { return s.value }

func (s *TenancyStep) Reset() {
	s.selector.Reset()
	s.complete = false
	s.value = false
}
