package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

// sampleJob is a simple struct used to exercise JSON and YAML marshalling paths.
type sampleJob struct {
	Name   string `json:"name" yaml:"name"`
	Status string `json:"status" yaml:"status"`
}

// --- Validate() tests ---

func TestValidateAcceptsValidFormats(t *testing.T) {
	formats := []Format{FormatTable, FormatWide, FormatJSON, FormatYAML}
	for _, f := range formats {
		err := f.Validate()
		assert.Nil(t, err, "expected nil error for format %q", f)
	}
}

func TestValidateRejectsInvalidFormat(t *testing.T) {
	err := Format("csv").Validate()
	assert.NotNil(t, err, "expected error for invalid format csv")
	if err != nil {
		msg := err.Error()
		assert.Contains(t, msg, "invalid output format", "error should use the spec-mandated prefix")
		assert.Contains(t, msg, "csv", "error should mention the invalid format")
		assert.Contains(t, msg, "table, wide, json, yaml", "error should list supported formats")
	}
}

// --- Render() tests ---

func TestRenderJSON(t *testing.T) {
	job := sampleJob{Name: "job-1", Status: "Running"}
	var buf bytes.Buffer
	err := FormatJSON.Render(&buf, job, RenderOptions{})
	assert.NoError(t, err)

	// Trim trailing newline added by Fprintln, then verify valid JSON.
	var got sampleJob
	err = json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &got)
	assert.NoError(t, err, "output should be valid JSON")
	assert.Equal(t, job, got)

	// 2-space indent: a multi-field JSON object will contain "\n  " (newline + 2 spaces).
	assert.Contains(t, buf.String(), "\n  ", "output should use 2-space indentation")
}

func TestRenderYAML(t *testing.T) {
	job := sampleJob{Name: "job-1", Status: "Running"}
	var buf bytes.Buffer
	err := FormatYAML.Render(&buf, job, RenderOptions{})
	assert.NoError(t, err)

	// Verify valid YAML by unmarshalling back.
	var got sampleJob
	err = yaml.Unmarshal(buf.Bytes(), &got)
	assert.NoError(t, err, "output should be valid YAML")
	assert.Equal(t, job, got)
}

func TestRenderTableCallsTableFn(t *testing.T) {
	called := false
	opts := RenderOptions{
		TableFn: func() string {
			called = true
			return "TABLE OUTPUT\n"
		},
	}
	var buf bytes.Buffer
	err := FormatTable.Render(&buf, nil, opts)
	assert.NoError(t, err)
	assert.True(t, called, "TableFn should be called for FormatTable")
	assert.Contains(t, buf.String(), "TABLE OUTPUT", "TableFn output should be printed")
}

func TestRenderWideCallsWideFn(t *testing.T) {
	wideCalled := false
	tableCalled := false
	opts := RenderOptions{
		TableFn: func() string {
			tableCalled = true
			return "TABLE OUTPUT\n"
		},
		WideFn: func() string {
			wideCalled = true
			return "WIDE OUTPUT\n"
		},
	}
	var buf bytes.Buffer
	err := FormatWide.Render(&buf, nil, opts)
	assert.NoError(t, err)
	assert.True(t, wideCalled, "WideFn should be called when provided")
	assert.False(t, tableCalled, "TableFn should NOT be called when WideFn is provided")
	assert.Contains(t, buf.String(), "WIDE OUTPUT", "WideFn output should be printed")
}

func TestRenderWideFallsBackToTableFnWhenWideFnNil(t *testing.T) {
	tableCalled := false
	opts := RenderOptions{
		TableFn: func() string {
			tableCalled = true
			return "TABLE OUTPUT\n"
		},
		WideFn: nil,
	}
	var buf bytes.Buffer
	err := FormatWide.Render(&buf, nil, opts)
	assert.NoError(t, err)
	assert.True(t, tableCalled, "TableFn should be called as fallback when WideFn is nil")
	assert.Contains(t, buf.String(), "TABLE OUTPUT", "TableFn output should be printed as fallback")
}

// --- Constant tests ---

func TestDefaultFormat(t *testing.T) {
	assert.Equal(t, FormatTable, DefaultFormat)
}

func TestFormatHelpText(t *testing.T) {
	assert.Equal(t, "Output format: table, wide, json, yaml", FormatHelpText)
}
