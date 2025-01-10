package steps

import tea "github.com/charmbracelet/bubbletea"

// Step interface defines what each wizard step must implement
type Step interface {
	tea.Model
	Title() string
	Description() string
	IsComplete() bool
	Value() interface{}
	Reset()
	View() string
}

// StepCompleteMsg is sent when a step is completed
type StepCompleteMsg struct {
	StepName string
	Value    interface{}
}

// StepBackMsg is sent when user wants to go back
type StepBackMsg struct{}

// StepSkipMsg is sent when a step should be skipped
type StepSkipMsg struct{}
