package output

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// Format represents a CLI output format (table, wide, json, or yaml).
type Format string

// Supported output formats.
const (
	FormatTable Format = "table"
	FormatWide  Format = "wide"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
)

// FormatSupported is the comma-separated list of supported format names,
// suitable for error messages and help text.
const FormatSupported = "table, wide, json, yaml"

// FormatHelpText is the help string shown next to the -o/--output flag.
const FormatHelpText = "Output format: " + FormatSupported

// DefaultFormat is the format used when the user does not pass -o/--output.
const DefaultFormat = FormatTable

// RenderOptions carries the callbacks used by the table-based render paths.
// TableFn is required for table and wide (fallback) rendering. WideFn, when
// non-nil, is preferred for the wide format.
type RenderOptions struct {
	TableFn func() string
	WideFn  func() string
}

// Validate returns nil when the format is one of the supported values,
// otherwise an error whose message names the offending value and lists the
// supported formats.
func (f Format) Validate() error {
	switch f {
	case FormatTable, FormatWide, FormatJSON, FormatYAML:
		return nil
	}
	return fmt.Errorf("invalid output format: %q (supported: %s)", f, FormatSupported)
}

// Render dispatches to the appropriate rendering path for this format.
//
//   - JSON:  marshals data with 2-space indentation and prints it followed
//     by a newline.
//   - YAML:  marshals data and prints it (yaml.Marshal already appends a
//     trailing newline).
//   - Wide:  calls WideFn when non-nil, otherwise falls back to TableFn.
//   - Table: calls TableFn.
func (f Format) Render(data interface{}, opts RenderOptions) error {
	switch f {
	case FormatJSON:
		b, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(b))
	case FormatYAML:
		b, err := yaml.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %w", err)
		}
		fmt.Print(string(b))
	case FormatWide:
		fn := opts.WideFn
		if fn == nil {
			fn = opts.TableFn
		}
		if fn != nil {
			fmt.Print(fn())
		}
	case FormatTable:
		if opts.TableFn != nil {
			fmt.Print(opts.TableFn())
		}
	}
	return nil
}
