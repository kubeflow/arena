package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	outputpkg "github.com/kubeflow/arena/pkg/output"
)

// CLIError is an error that teaches: it carries the accepted values and an
// optional hint so users and agents can self-correct without re-reading docs.
type CLIError struct {
	Message     string   `json:"message"`
	ValidValues []string `json:"validValues,omitempty"`
	Hint        string   `json:"hint,omitempty"`
}

// Error renders the message plus the enumerated valid values and hint.
func (e *CLIError) Error() string {
	msg := e.Message
	if len(e.ValidValues) > 0 {
		msg += "; must be one of: " + strings.Join(e.ValidValues, ", ")
	}
	if e.Hint != "" {
		msg += " (" + e.Hint + ")"
	}
	return msg
}

// Envelope renders the error as a JSON object for machine consumption.
func (e *CLIError) Envelope() string {
	b, err := json.MarshalIndent(struct {
		Error *CLIError `json:"error"`
	}{e}, "", "  ")
	if err != nil {
		return "Error: " + e.Error() + "\n"
	}
	return string(b) + "\n"
}

// ErrorEnvelope renders any error as the machine-readable JSON envelope.
// A CLIError (possibly wrapped) keeps its teaching fields; every other
// error is carried as a bare message so consumers parsing stderr always
// receive well-formed output.
func ErrorEnvelope(err error) string {
	var cliErr *CLIError
	if !errors.As(err, &cliErr) {
		cliErr = &CLIError{Message: err.Error()}
	}
	return cliErr.Envelope()
}

// JSONErrorRequested reports whether errors should be rendered as a JSON
// envelope, which is the case when the active output format is json or yaml.
// The shared outputFormat var is bound by every leaf command's local
// -o/--output registration via registerOutputFlag.
func JSONErrorRequested() bool {
	switch outputpkg.Format(outputFormat) {
	case outputpkg.FormatJSON, outputpkg.FormatYAML:
		return true
	}
	return false
}

// validateOutputFormatValue validates an output format value, returning a
// teaching error that enumerates the accepted formats.
func validateOutputFormatValue(format string) error {
	if err := outputpkg.Format(format).Validate(); err != nil {
		return &CLIError{
			Message:     fmt.Sprintf("invalid output format %q for --output", format),
			ValidValues: strings.Split(outputpkg.FormatSupported, ", "),
		}
	}
	return nil
}

// validateOutputFormat validates the shared outputFormat flag value.
func validateOutputFormat() error {
	return validateOutputFormatValue(outputFormat)
}
