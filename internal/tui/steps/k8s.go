package steps

import (
	"github.com/DenisBytes/gonstrukt/internal/tui/components"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// K8sStep allows enabling/disabling k3s dev environment
type K8sStep struct {
	selector components.Selector
	complete bool
	value    bool
}

// NewK8sStep creates a new k8s selection step
func NewK8sStep() *K8sStep {
	options := []components.Option{
		{
			Label:       "Enable K8s Dev Environment",
			Value:       "true",
			Description: "Generate k3s manifests with Nginx Ingress, mkcert TLS, observability stack",
		},
		{
			Label:       "Disable",
			Value:       "false",
			Description: "Run services directly on localhost without k3s",
		},
	}

	return &K8sStep{
		selector: components.NewSelector("", options),
	}
}

func (s *K8sStep) Init() tea.Cmd { return nil }

func (s *K8sStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				StepName: "k8s",
				Value:    s.value,
			}
		}
	}

	return s, cmd
}

func (s *K8sStep) View() string        { return s.selector.View() }
func (s *K8sStep) Title() string        { return "Enable K8s Dev Environment?" }
func (s *K8sStep) Description() string  { return "Generate a k3s-based local dev environment with TLS, ingress, and observability" }
func (s *K8sStep) IsComplete() bool     { return s.complete }
func (s *K8sStep) Value() interface{}   { return s.value }

func (s *K8sStep) Reset() {
	s.selector.Reset()
	s.complete = false
	s.value = false
}
