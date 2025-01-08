package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Option represents a single selectable option
type Option struct {
	Label       string
	Value       string
	Description string
	Disabled    bool
}

// Selector is a component for selecting from a list of options
type Selector struct {
	Options  []Option
	Cursor   int
	Selected int
	Title    string
	styles   Styles
}

// NewSelector creates a new selector with the given options
func NewSelector(title string, options []Option) Selector {
	return Selector{
		Options:  options,
		Cursor:   0,
		Selected: -1,
		Title:    title,
		styles:   DefaultStyles(),
	}
}

// Init initializes the selector
func (s Selector) Init() tea.Cmd {
	return nil
}

// Update handles key events
func (s Selector) Update(msg tea.Msg) (Selector, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			s.moveUp()
		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			s.moveDown()
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if !s.Options[s.Cursor].Disabled {
				s.Selected = s.Cursor
			}
		}
	}
	return s, nil
}

// View renders the selector
func (s Selector) View() string {
	var b strings.Builder

	if s.Title != "" {
		b.WriteString(s.styles.Title.Render(s.Title))
		b.WriteString("\n\n")
	}

	for i, opt := range s.Options {
		cursor := "  "
		if i == s.Cursor {
			cursor = Indicator()
		}

		var style = s.styles.Option
		if opt.Disabled {
			style = s.styles.DisabledOption
		} else if i == s.Cursor {
			style = s.styles.SelectedOption
		}

		line := fmt.Sprintf("%s%s", cursor, opt.Label)
		b.WriteString(style.Render(line))

		if opt.Description != "" && i == s.Cursor && !opt.Disabled {
			b.WriteString("\n")
			b.WriteString(s.styles.Description.PaddingLeft(4).Render(opt.Description))
		}

		b.WriteString("\n")
	}

	return b.String()
}

// moveUp moves the cursor up to the previous non-disabled option
func (s *Selector) moveUp() {
	for i := s.Cursor - 1; i >= 0; i-- {
		if !s.Options[i].Disabled {
			s.Cursor = i
			return
		}
	}
	// Wrap to bottom
	for i := len(s.Options) - 1; i > s.Cursor; i-- {
		if !s.Options[i].Disabled {
			s.Cursor = i
			return
		}
	}
}

// moveDown moves the cursor down to the next non-disabled option
func (s *Selector) moveDown() {
	for i := s.Cursor + 1; i < len(s.Options); i++ {
		if !s.Options[i].Disabled {
			s.Cursor = i
			return
		}
	}
	// Wrap to top
	for i := 0; i < s.Cursor; i++ {
		if !s.Options[i].Disabled {
			s.Cursor = i
			return
		}
	}
}

// IsSelected returns true if an option has been selected
func (s Selector) IsSelected() bool {
	return s.Selected >= 0 && s.Selected < len(s.Options)
}

// SelectedOption returns the selected option
func (s Selector) SelectedOption() *Option {
	if s.IsSelected() {
		return &s.Options[s.Selected]
	}
	return nil
}

// SelectedValue returns the value of the selected option
func (s Selector) SelectedValue() string {
	if opt := s.SelectedOption(); opt != nil {
		return opt.Value
	}
	return ""
}

// Reset resets the selector state
func (s *Selector) Reset() {
	s.Cursor = 0
	s.Selected = -1
}
