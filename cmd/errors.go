package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fatih/color"
)

type ValidationError struct {
	Field   string
	Value   string
	Message string
	Err     error
}

func (e *ValidationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("validation failed for %s: %s (%v)", e.Field, e.Message, e.Err)
	}
	return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

type InvalidFormatError struct {
	Field          string
	Value          string
	ExpectedFormat string
}

func (e *InvalidFormatError) Error() string {
	return fmt.Sprintf("invalid format for %s: got '%s', expected format: %s",
		e.Field, e.Value, e.ExpectedFormat)
}

type InvalidOptionError struct {
	Field   string
	Value   string
	Options []string
}

func (e *InvalidOptionError) Error() string {
	return fmt.Sprintf("invalid %s '%s', valid options: %s",
		e.Field, e.Value, strings.Join(e.Options, ", "))
}

type MissingRequiredFieldError struct {
	Field       string
	ServiceType string
	Reason      string
}

func (e *MissingRequiredFieldError) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("%s is required for %s services: %s", e.Field, e.ServiceType, e.Reason)
	}
	return fmt.Sprintf("%s is required for %s services", e.Field, e.ServiceType)
}

var (
	ErrServiceNameRequired = errors.New("service name is required")
	ErrInvalidServiceName  = errors.New("service name format is invalid")
)

func NewValidationError(field, value, message string, err error) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
		Err:     err,
	}
}

func NewInvalidFormatError(field, value, expectedFormat string) *InvalidFormatError {
	return &InvalidFormatError{
		Field:          field,
		Value:          value,
		ExpectedFormat: expectedFormat,
	}
}

func NewInvalidOptionError(field, value string, options []string) *InvalidOptionError {
	return &InvalidOptionError{
		Field:   field,
		Value:   value,
		Options: options,
	}
}

func NewMissingRequiredFieldError(field, serviceType, reason string) *MissingRequiredFieldError {
	return &MissingRequiredFieldError{
		Field:       field,
		ServiceType: serviceType,
		Reason:      reason,
	}
}

func IsValidationError(err error) bool {
	var validationErr *ValidationError
	return errors.As(err, &validationErr)
}

func IsInvalidOptionError(err error) bool {
	var invalidOptionErr *InvalidOptionError
	return errors.As(err, &invalidOptionErr)
}

func IsMissingRequiredFieldError(err error) bool {
	var missingFieldErr *MissingRequiredFieldError
	return errors.As(err, &missingFieldErr)
}

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

	errorText := err.Error()
	lines := strings.Split(errorText, "\n")

	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			if i == 0 {
				result.WriteString(fmt.Sprintf("%s %s", red.Sprint("error:"), line))
			} else {
				result.WriteString(fmt.Sprintf("%s %s", red.Sprint("error:"), line))
			}
			result.WriteString("\n")
		}
	}

	return result.String()
}
