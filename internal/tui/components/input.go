package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Input is a text input component with validation
type Input struct {
	textinput  textinput.Model
	Title      string
	Error      string
	Validator  func(string) error
	styles     Styles
	Submitted  bool
}

// NewInput creates a new text input component
func NewInput(title, placeholder string, validator func(string) error) Input {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 256
	ti.Width = 50
	ti.Focus()

	return Input{
		textinput: ti,
		Title:     title,
		Validator: validator,
		styles:    DefaultStyles(),
	}
}

// Init initializes the input
func (i Input) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages
func (i Input) Update(msg tea.Msg) (Input, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if i.Validator != nil {
				if err := i.Validator(i.textinput.Value()); err != nil {
					i.Error = err.Error()
					return i, nil
				}
			}
			i.Error = ""
			i.Submitted = true
			return i, nil
		}
	}

	i.textinput, cmd = i.textinput.Update(msg)

	// Clear error on edit
	if i.Error != "" {
		i.Error = ""
	}

	return i, cmd
}

// View renders the input
func (i Input) View() string {
	var b strings.Builder

	if i.Title != "" {
		b.WriteString(i.styles.Title.Render(i.Title))
		b.WriteString("\n\n")
	}

	// Render input with appropriate style
	inputStyle := i.styles.Input
	if i.textinput.Focused() {
		inputStyle = i.styles.InputFocus
	}

	b.WriteString(inputStyle.Render(i.textinput.View()))
	b.WriteString("\n")

	if i.Error != "" {
		b.WriteString("\n")
		b.WriteString(i.styles.Error.Render("✗ " + i.Error))
	}

	return b.String()
}

// Value returns the current input value
func (i Input) Value() string {
	return i.textinput.Value()
}

// SetValue sets the input value
func (i *Input) SetValue(value string) {
	i.textinput.SetValue(value)
}

// Focus focuses the input
func (i *Input) Focus() tea.Cmd {
	return i.textinput.Focus()
}

// Blur removes focus from the input
func (i *Input) Blur() {
	i.textinput.Blur()
}

// Reset resets the input state
func (i *Input) Reset() {
	i.textinput.SetValue("")
	i.Error = ""
	i.Submitted = false
}

// IsSubmitted returns true if the input has been submitted
func (i Input) IsSubmitted() bool {
	return i.Submitted
}
