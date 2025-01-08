package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// MultiSelector is a component for selecting multiple options from a list
type MultiSelector struct {
	Options   []Option
	Cursor    int
	Selected  map[int]bool
	Title     string
	styles    Styles
	confirmed bool
	minSelect int // minimum selections required (0 = optional)
	maxSelect int // maximum selections allowed (0 = unlimited)
}

// NewMultiSelector creates a new multi-selector with the given options
func NewMultiSelector(title string, options []Option) MultiSelector {
	return MultiSelector{
		Options:   options,
		Cursor:    0,
		Selected:  make(map[int]bool),
		Title:     title,
		styles:    DefaultStyles(),
		confirmed: false,
		minSelect: 0,
		maxSelect: 0,
	}
}

// NewMultiSelectorWithLimits creates a new multi-selector with selection limits
func NewMultiSelectorWithLimits(title string, options []Option, min, max int) MultiSelector {
	return MultiSelector{
		Options:   options,
		Cursor:    0,
		Selected:  make(map[int]bool),
		Title:     title,
		styles:    DefaultStyles(),
		confirmed: false,
		minSelect: min,
		maxSelect: max,
	}
}

// Init initializes the selector
func (s MultiSelector) Init() tea.Cmd {
	return nil
}

// Update handles key events
func (s MultiSelector) Update(msg tea.Msg) (MultiSelector, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			s.moveUp()
		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			s.moveDown()
		case key.Matches(msg, key.NewBinding(key.WithKeys(" "))):
			// Toggle selection with space
			if !s.Options[s.Cursor].Disabled {
				s.toggleSelection(s.Cursor)
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			// Confirm selection if minimum is met
			if s.canConfirm() {
				s.confirmed = true
			}
		}
	}
	return s, nil
}

// toggleSelection toggles the selection state of an option
func (s *MultiSelector) toggleSelection(idx int) {
	if s.Selected[idx] {
		delete(s.Selected, idx)
	} else {
		// Check max limit
		if s.maxSelect > 0 && len(s.Selected) >= s.maxSelect {
			return
		}
		s.Selected[idx] = true
	}
}

// canConfirm returns true if the minimum selection requirement is met
func (s MultiSelector) canConfirm() bool {
	return len(s.Selected) >= s.minSelect
}

// View renders the multi-selector
func (s MultiSelector) View() string {
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

		checkbox := "[ ]"
		if s.Selected[i] {
			checkbox = "[x]"
		}

		var style = s.styles.Option
		if opt.Disabled {
			style = s.styles.DisabledOption
		} else if i == s.Cursor {
			style = s.styles.SelectedOption
		}

		line := fmt.Sprintf("%s%s %s", cursor, checkbox, opt.Label)
		b.WriteString(style.Render(line))

		if opt.Description != "" && i == s.Cursor && !opt.Disabled {
			b.WriteString("\n")
			b.WriteString(s.styles.Description.PaddingLeft(4).Render(opt.Description))
		}

		b.WriteString("\n")
	}

	// Show help text
	b.WriteString("\n")
	helpText := "space toggle, enter confirm"
	if s.minSelect > 0 {
		helpText = fmt.Sprintf("space toggle, enter confirm (min %d)", s.minSelect)
	}
	b.WriteString(s.styles.Description.Render(helpText))

	return b.String()
}

// moveUp moves the cursor up to the previous non-disabled option
func (s *MultiSelector) moveUp() {
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
func (s *MultiSelector) moveDown() {
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

// IsConfirmed returns true if selection has been confirmed
func (s MultiSelector) IsConfirmed() bool {
	return s.confirmed
}

// SelectedOptions returns the selected options
func (s MultiSelector) SelectedOptions() []Option {
	var selected []Option
	for i, opt := range s.Options {
		if s.Selected[i] {
			selected = append(selected, opt)
		}
	}
	return selected
}

// SelectedValues returns the values of the selected options
func (s MultiSelector) SelectedValues() []string {
	var values []string
	for i, opt := range s.Options {
		if s.Selected[i] {
			values = append(values, opt.Value)
		}
	}
	return values
}

// Reset resets the selector state
func (s *MultiSelector) Reset() {
	s.Cursor = 0
	s.Selected = make(map[int]bool)
	s.confirmed = false
}
