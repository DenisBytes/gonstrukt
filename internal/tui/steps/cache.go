package steps

import (
	"github.com/DenisBytes/gonstrukt/internal/config"
	"github.com/DenisBytes/gonstrukt/internal/tui/components"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// CacheStep allows selection of the cache type
type CacheStep struct {
	selector components.Selector
	complete bool
	value    config.CacheType
}

// NewCacheStep creates a new cache selection step
func NewCacheStep() *CacheStep {
	options := []components.Option{
		{
			Label:       "Redis",
			Value:       string(config.CacheRedis),
			Description: "Industry standard. Distributed caching with persistence options.",
		},
		{
			Label:       "Valkey",
			Value:       string(config.CacheValkey),
			Description: "Redis-compatible fork. Drop-in replacement with community focus.",
		},
		{
			Label:       "In-Memory",
			Value:       string(config.CacheMemory),
			Description: "Local memory cache. Good for single-instance deployments.",
		},
	}

	return &CacheStep{
		selector: components.NewSelector("", options),
	}
}

// Init initializes the step
func (s *CacheStep) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (s *CacheStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		s.value = config.CacheType(s.selector.SelectedValue())
		return s, func() tea.Msg {
			return StepCompleteMsg{
				StepName: "cache",
				Value:    s.value,
			}
		}
	}

	return s, cmd
}

// View renders the step
func (s *CacheStep) View() string {
	return s.selector.View()
}

// Title returns the step title
func (s *CacheStep) Title() string {
	return "Select Cache"
}

// Description returns the step description
func (s *CacheStep) Description() string {
	return "Choose the caching backend for the gateway"
}

// IsComplete returns true if the step is complete
func (s *CacheStep) IsComplete() bool {
	return s.complete
}

// Value returns the selected value
func (s *CacheStep) Value() interface{} {
	return s.value
}

// Reset resets the step
func (s *CacheStep) Reset() {
	s.selector.Reset()
	s.complete = false
	s.value = ""
}
