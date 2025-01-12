package steps

import (
	"github.com/DenisBytes/gonstrukt/internal/config"
	"github.com/DenisBytes/gonstrukt/internal/tui/components"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// RateLimiterStep allows selection of the rate limiter algorithm
type RateLimiterStep struct {
	selector components.Selector
	complete bool
	value    config.RateLimiterType
}

// NewRateLimiterStep creates a new rate limiter selection step
func NewRateLimiterStep() *RateLimiterStep {
	options := []components.Option{
		{
			Label:       "Token Bucket",
			Value:       string(config.RateLimiterTokenBucket),
			Description: "Classic algorithm. Allows burst traffic up to bucket capacity.",
		},
		{
			Label:       "Sliding Window",
			Value:       string(config.RateLimiterSlidingWindow),
			Description: "Approximated sliding window. Smoother rate limiting, less bursty.",
		},
	}

	return &RateLimiterStep{
		selector: components.NewSelector("", options),
	}
}

// Init initializes the step
func (s *RateLimiterStep) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (s *RateLimiterStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		s.value = config.RateLimiterType(s.selector.SelectedValue())
		return s, func() tea.Msg {
			return StepCompleteMsg{
				StepName: "rate_limiter",
				Value:    s.value,
			}
		}
	}

	return s, cmd
}

// View renders the step
func (s *RateLimiterStep) View() string {
	return s.selector.View()
}

// Title returns the step title
func (s *RateLimiterStep) Title() string {
	return "Select Rate Limiter"
}

// Description returns the step description
func (s *RateLimiterStep) Description() string {
	return "Choose the rate limiting algorithm for the gateway"
}

// IsComplete returns true if the step is complete
func (s *RateLimiterStep) IsComplete() bool {
	return s.complete
}

// Value returns the selected value
func (s *RateLimiterStep) Value() interface{} {
	return s.value
}

// Reset resets the step
func (s *RateLimiterStep) Reset() {
	s.selector.Reset()
	s.complete = false
	s.value = ""
}
