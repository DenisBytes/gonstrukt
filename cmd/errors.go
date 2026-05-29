package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// CliError wraps an error with an optional usage string for friendly CLI output.
type CliError struct {
	Err   error
	Usage string
}

func (e *CliError) Error() string {
	return e.Err.Error()
}

func (e *CliError) Unwrap() error {
	return e.Err
}

func NewCliError(err error, usage string) *CliError {
	return &CliError{
		Err:   err,
		Usage: usage,
	}
}

func FormatCliError(err error) string {
	red := color.New(color.FgRed, color.Bold)

	var cliErr *CliError
	if errors.As(err, &cliErr) {
		result := formatErrorLines(cliErr.Err, red) + "\n"
		if cliErr.Usage != "" {
			result += cliErr.Usage + "\n\n"
		}
		return result
	}

	return formatErrorLines(err, red) + "\n"
}

func formatErrorLines(err error, red *color.Color) string {
	var result strings.Builder
	result.WriteString("\n")

	for _, line := range strings.Split(err.Error(), "\n") {
		if strings.TrimSpace(line) != "" {
			fmt.Fprintf(&result, "%s %s\n", red.Sprint("error:"), line)
		}
	}

	return result.String()
}
