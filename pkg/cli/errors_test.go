package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLIError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *CLIError
		want string
	}{
		{
			name: "message only",
			err:  &CLIError{Message: "something failed"},
			want: "something failed",
		},
		{
			name: "message with valid values",
			err:  &CLIError{Message: `invalid value "js" for --output`, ValidValues: []string{"table", "wide", "json", "yaml"}},
			want: `invalid value "js" for --output; must be one of: table, wide, json, yaml`,
		},
		{
			name: "message with hint",
			err:  &CLIError{Message: "bad expression", Hint: "expected key=value"},
			want: "bad expression (expected key=value)",
		},
		{
			name: "message with both",
			err:  &CLIError{Message: "bad value", ValidValues: []string{"a", "b"}, Hint: "pick one"},
			want: "bad value; must be one of: a, b (pick one)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.err.Error())
		})
	}
}

func TestCLIError_Envelope(t *testing.T) {
	e := &CLIError{
		Message:     `invalid value "js" for --output`,
		ValidValues: []string{"table", "wide", "json", "yaml"},
		Hint:        "pick one",
	}

	var parsed struct {
		Error struct {
			Message     string   `json:"message"`
			ValidValues []string `json:"validValues"`
			Hint        string   `json:"hint"`
		} `json:"error"`
	}
	require.NoError(t, json.Unmarshal([]byte(e.Envelope()), &parsed))
	assert.Equal(t, `invalid value "js" for --output`, parsed.Error.Message)
	assert.Equal(t, []string{"table", "wide", "json", "yaml"}, parsed.Error.ValidValues)
	assert.Equal(t, "pick one", parsed.Error.Hint)
}

func TestCLIError_EnvelopeOmitsEmptyFields(t *testing.T) {
	e := &CLIError{Message: "boom"}
	assert.NotContains(t, e.Envelope(), "validValues")
	assert.NotContains(t, e.Envelope(), "hint")
}

func TestErrorEnvelope_PlainErrorCarriesMessageOnly(t *testing.T) {
	out := ErrorEnvelope(errors.New("api failure"))
	assert.Contains(t, out, `"message": "api failure"`)
	assert.NotContains(t, out, "validValues")
	assert.NotContains(t, out, "hint")
	assert.True(t, strings.HasSuffix(out, "\n"), "envelope ends with newline")
}

func TestErrorEnvelope_WrappedCLIErrorKeepsTeachingFields(t *testing.T) {
	err := fmt.Errorf("submit failed: %w", &CLIError{Message: "bad flag", ValidValues: []string{"a", "b"}, Hint: "pick one"})
	out := ErrorEnvelope(err)
	assert.Contains(t, out, `"message": "bad flag"`)
	assert.Contains(t, out, `"validValues": [`)
	assert.Contains(t, out, `"hint": "pick one"`)
	assert.NotContains(t, out, "submit failed", "envelope carries the CLIError, not the wrapper")
}

func TestJSONErrorRequested(t *testing.T) {
	orig := outputFormat
	t.Cleanup(func() { outputFormat = orig })

	for _, f := range []string{"json", "yaml"} {
		outputFormat = f
		assert.True(t, JSONErrorRequested(), "format %s should request JSON errors", f)
	}
	for _, f := range []string{"table", "wide", "", "bogus"} {
		outputFormat = f
		assert.False(t, JSONErrorRequested(), "format %q should not request JSON errors", f)
	}
}

func TestTopJobCmd_OutputFlagBindsSharedFormat(t *testing.T) {
	orig := outputFormat
	t.Cleanup(func() { outputFormat = orig })

	newTopJobCmd() // registerOutputFlag resets the shared var to the default
	require.NoError(t, newTopJobCmd().Flags().Set("output", "json"))
	assert.True(t, JSONErrorRequested(), "top -o json must request JSON errors via the shared format var")

	require.NoError(t, newTopJobCmd().Flags().Set("output", "table"))
	assert.False(t, JSONErrorRequested(), "top -o table must not request JSON errors")
}

func TestCLIError_ErrorsAs(t *testing.T) {
	var err error = &CLIError{Message: "wrapped"}
	var target *CLIError
	require.True(t, errors.As(err, &target))
	assert.Equal(t, "wrapped", target.Message)
}

func TestValidateOutputFormat(t *testing.T) {
	orig := outputFormat
	t.Cleanup(func() { outputFormat = orig })

	for _, f := range []string{"table", "wide", "json", "yaml"} {
		outputFormat = f
		assert.NoError(t, validateOutputFormat(), "format %q should be valid", f)
	}

	outputFormat = "js"
	err := validateOutputFormat()
	var cliErr *CLIError
	require.Error(t, err)
	require.True(t, errors.As(err, &cliErr))
	assert.Equal(t, []string{"table", "wide", "json", "yaml"}, cliErr.ValidValues)
	assert.Contains(t, err.Error(), "invalid output format")
	assert.Contains(t, err.Error(), "must be one of: table, wide, json, yaml")
}

func TestValidateOutputFormatValue(t *testing.T) {
	for _, f := range []string{"table", "wide", "json", "yaml"} {
		assert.NoError(t, validateOutputFormatValue(f), "format %q should be valid", f)
	}

	err := validateOutputFormatValue("js")
	var cliErr *CLIError
	require.Error(t, err)
	require.True(t, errors.As(err, &cliErr))
	assert.Equal(t, `invalid output format "js" for --output`, cliErr.Message)
	assert.Equal(t, []string{"table", "wide", "json", "yaml"}, cliErr.ValidValues)
	assert.Empty(t, cliErr.Hint)
}

func TestValidateOutputFormat_Integration(t *testing.T) {
	orig := outputFormat
	t.Cleanup(func() { outputFormat = orig })

	tmpFile := writeTestYAML(t, testRunYAML)
	err := ExecuteWithArgs([]string{"job", "run", "-f", tmpFile, "-o", "bogus"})
	var cliErr *CLIError
	require.Error(t, err)
	require.True(t, errors.As(err, &cliErr))
	assert.Contains(t, err.Error(), "must be one of: table, wide, json, yaml")
}

func TestDryRunFormat_IsTeachingError(t *testing.T) {
	orig := outputFormat
	t.Cleanup(func() { outputFormat = orig })

	outputFormat = "table"
	_, err := dryRunFormat()
	var cliErr *CLIError
	require.Error(t, err)
	require.True(t, errors.As(err, &cliErr))
	assert.Equal(t, []string{"json", "yaml"}, cliErr.ValidValues)
	assert.Contains(t, err.Error(), "dry-run only supports -o json or yaml")
}

func TestSubmitFrameworkErrorsAreTeaching(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantSubstr string
		wantValid  []string
	}{
		{
			name:       "v1 type lists v2 frameworks",
			args:       []string{"submit", "ray", "--name", "x", "--image", "img"},
			wantSubstr: "not supported by arena-v2 yet",
			wantValid:  []string{"pytorch", "tensorflow", "mpi", "horovod", "deepspeed"},
		},
		{
			name:       "unknown type lists accepted aliases",
			args:       []string{"submit", "jax", "--name", "x", "--image", "img"},
			wantSubstr: "unsupported framework type",
			wantValid:  []string{"pytorch", "pytorchjob", "tf", "tfjob", "tensorflow", "mpi", "mpijob", "mj", "horovod", "horovodjob", "hj", "deepspeed", "deepspeedjob", "dp"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ExecuteWithArgs(tt.args)
			var cliErr *CLIError
			require.Error(t, err)
			require.True(t, errors.As(err, &cliErr), "error should be a CLIError, got %T: %v", err, err)
			assert.Contains(t, err.Error(), tt.wantSubstr)
			assert.Equal(t, tt.wantValid, cliErr.ValidValues)
			assert.Contains(t, err.Error(), "must be one of:")
		})
	}
}

func TestRunCmd_SetSyntaxErrorHasHint(t *testing.T) {
	orig := outputFormat
	t.Cleanup(func() { outputFormat = orig })

	tmpFile := writeTestYAML(t, testRunYAML)
	err := ExecuteWithArgs([]string{"job", "run", "-f", tmpFile, "--set", "noequals"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to apply --set")
	assert.Contains(t, err.Error(), "failed to parse --set")
	assert.Contains(t, err.Error(), "worker.replicas=4")

	var cliErr *CLIError
	require.True(t, errors.As(err, &cliErr))
	assert.NotEmpty(t, cliErr.Hint)
}
