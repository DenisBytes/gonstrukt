package steps

import (
	"fmt"
	"regexp"

	"github.com/DenisBytes/gonstrukt/internal/tui/components"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// DomainStep allows input of the local dev domain
type DomainStep struct {
	input    components.Input
	complete bool
	value    string
}

// NewDomainStep creates a new domain input step
func NewDomainStep(projectName string) *DomainStep {
	placeholder := "myapp.dev"
	if projectName != "" {
		placeholder = projectName + ".dev"
	}

	validator := func(value string) error {
		if value == "" {
			return fmt.Errorf("domain is required for k8s dev environment")
		}
		pattern := `^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z]{2,})+$`
		matched, _ := regexp.MatchString(pattern, value)
		if !matched {
			return fmt.Errorf("invalid domain format (e.g., myapp.dev)")
		}
		return nil
	}

	return &DomainStep{
		input: components.NewInput("", placeholder, validator),
	}
}

func (s *DomainStep) Init() tea.Cmd { return s.input.Init() }

func (s *DomainStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				StepName: "domain",
				Value:    s.value,
			}
		}
	}

	return s, cmd
}

func (s *DomainStep) View() string        { return s.input.View() }
func (s *DomainStep) Title() string        { return "Enter Dev Domain" }
func (s *DomainStep) Description() string  { return "Local HTTPS domain for k8s dev environment (mkcert will generate TLS certs)" }
func (s *DomainStep) IsComplete() bool     { return s.complete }
func (s *DomainStep) Value() interface{}   { return s.value }

func (s *DomainStep) Reset() {
	s.input.Reset()
	s.complete = false
	s.value = ""
}
