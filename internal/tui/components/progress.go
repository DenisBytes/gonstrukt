package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProgressState represents the state of a generation step
type ProgressState int

const (
	ProgressPending ProgressState = iota
	ProgressInProgress
	ProgressComplete
	ProgressError
)

// ProgressStep represents a single step in the generation process
type ProgressStep struct {
	Label   string
	State   ProgressState
	Message string
}

// Progress displays generation progress with a spinner
type Progress struct {
	spinner spinner.Model
	Steps   []ProgressStep
	Current int
	Done    bool
	Error   error
	styles  Styles
}

// NewProgress creates a new progress component
func NewProgress(steps []string) Progress {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(Primary)

	progressSteps := make([]ProgressStep, len(steps))
	for i, label := range steps {
		progressSteps[i] = ProgressStep{
			Label: label,
			State: ProgressPending,
		}
	}

	return Progress{
		spinner: s,
		Steps:   progressSteps,
		Current: 0,
		styles:  DefaultStyles(),
	}
}

// Init initializes the progress component
func (p Progress) Init() tea.Cmd {
	return p.spinner.Tick
}

// Update handles messages
func (p Progress) Update(msg tea.Msg) (Progress, tea.Cmd) {
	var cmd tea.Cmd
	p.spinner, cmd = p.spinner.Update(msg)
	return p, cmd
}

// View renders the progress
func (p Progress) View() string {
	var b strings.Builder

	b.WriteString(p.styles.Title.Render("Generating Project..."))
	b.WriteString("\n\n")

	for i, step := range p.Steps {
		var prefix string
		var style lipgloss.Style

		switch step.State {
		case ProgressPending:
			prefix = "○"
			style = lipgloss.NewStyle().Foreground(Muted)
		case ProgressInProgress:
			prefix = p.spinner.View()
			style = lipgloss.NewStyle().Foreground(Primary)
		case ProgressComplete:
			prefix = "✓"
			style = lipgloss.NewStyle().Foreground(Success)
		case ProgressError:
			prefix = "✗"
			style = lipgloss.NewStyle().Foreground(Error)
		}

		line := fmt.Sprintf("%s %s", prefix, step.Label)
		b.WriteString(style.Render(line))

		if step.Message != "" && (step.State == ProgressInProgress || step.State == ProgressError) {
			b.WriteString("\n")
			b.WriteString(lipgloss.NewStyle().
				Foreground(Muted).
				PaddingLeft(4).
				Render(step.Message))
		}

		if i < len(p.Steps)-1 {
			b.WriteString("\n")
		}
	}

	if p.Done && p.Error == nil {
		b.WriteString("\n\n")
		b.WriteString(p.styles.Success.Render("✓ Project generated successfully!"))
	}

	if p.Error != nil {
		b.WriteString("\n\n")
		b.WriteString(p.styles.Error.Render(fmt.Sprintf("✗ Error: %s", p.Error.Error())))
	}

	return b.String()
}

// StartStep marks a step as in progress
func (p *Progress) StartStep(index int, message string) {
	if index >= 0 && index < len(p.Steps) {
		p.Steps[index].State = ProgressInProgress
		p.Steps[index].Message = message
		p.Current = index
	}
}

// CompleteStep marks a step as complete
func (p *Progress) CompleteStep(index int) {
	if index >= 0 && index < len(p.Steps) {
		p.Steps[index].State = ProgressComplete
		p.Steps[index].Message = ""
	}
}

// FailStep marks a step as failed
func (p *Progress) FailStep(index int, err error) {
	if index >= 0 && index < len(p.Steps) {
		p.Steps[index].State = ProgressError
		p.Steps[index].Message = err.Error()
		p.Error = err
	}
}

// Complete marks the entire progress as done
func (p *Progress) Complete() {
	p.Done = true
}

// ProgressMsg is a message sent to update progress
type ProgressMsg struct {
	Step    int
	State   ProgressState
	Message string
	Error   error
}

// StartStepCmd creates a command to start a step
func StartStepCmd(step int, message string) tea.Cmd {
	return func() tea.Msg {
		return ProgressMsg{
			Step:    step,
			State:   ProgressInProgress,
			Message: message,
		}
	}
}

// CompleteStepCmd creates a command to complete a step
func CompleteStepCmd(step int) tea.Cmd {
	return func() tea.Msg {
		return ProgressMsg{
			Step:  step,
			State: ProgressComplete,
		}
	}
}

// FailStepCmd creates a command to fail a step
func FailStepCmd(step int, err error) tea.Cmd {
	return func() tea.Msg {
		return ProgressMsg{
			Step:  step,
			State: ProgressError,
			Error: err,
		}
	}
}
