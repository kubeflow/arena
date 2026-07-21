package log

import (
	"bytes"
	"errors"
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/klog/v2"
)

// resetKlogState resets klog global state (verbosity and stderr output) after each test.
func resetKlogState(t *testing.T, flags *flag.FlagSet) {
	t.Helper()
	t.Cleanup(func() {
		_ = flags.Set("v", "0")
		klog.SetOutput(os.Stderr)
	})
}

// captureOutput redirects klog output to a buffer for test assertions.
// It disables logtostderr so output goes to the buffer instead of stderr.
func captureOutput(t *testing.T, flags *flag.FlagSet) *bytes.Buffer {
	t.Helper()
	_ = flags.Set("logtostderr", "false")
	var buf bytes.Buffer
	klog.SetOutput(&buf)
	return &buf
}

func TestInit_ConfiguresVerbosity(t *testing.T) {
	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	Init(flags)
	resetKlogState(t, flags)

	// Set verbosity to 2
	err := flags.Set("v", "2")
	assert.NoError(t, err)

	assert.True(t, V(1))
	assert.True(t, V(2))
	assert.False(t, V(3))
}

func TestInit_DefaultVerbosity(t *testing.T) {
	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	Init(flags)
	resetKlogState(t, flags)

	assert.True(t, V(0))
	assert.False(t, V(1))
}

func TestInfo_LogsMessage(t *testing.T) {
	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	Init(flags)
	resetKlogState(t, flags)

	buf := captureOutput(t, flags)

	Info("test message", "key", "value")
	klog.Flush()

	assert.Contains(t, buf.String(), "test message")
	assert.Contains(t, buf.String(), "key")
	assert.Contains(t, buf.String(), "value")
}

func TestError_LogsErrorAndMessage(t *testing.T) {
	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	Init(flags)
	resetKlogState(t, flags)

	buf := captureOutput(t, flags)

	err := errors.New("test error")
	Error(err, "operation failed", "operation", "create")
	klog.Flush()

	assert.Contains(t, buf.String(), "operation failed")
	assert.Contains(t, buf.String(), "test error")
	assert.Contains(t, buf.String(), "operation")
	assert.Contains(t, buf.String(), "create")
}

func TestDebug_OnlyLogsWhenVerbose(t *testing.T) {
	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	Init(flags)
	resetKlogState(t, flags)

	// Ensure verbosity is 0
	_ = flags.Set("v", "0")
	buf := captureOutput(t, flags)

	Debug("debug message")
	klog.Flush()

	assert.Empty(t, buf.String(), "debug should not log at v=0")

	// Now set high verbosity and verify debug logs
	buf.Reset()
	_ = flags.Set("v", "2")

	Debug("debug message")
	klog.Flush()

	assert.Contains(t, buf.String(), "debug message")
}

func TestWarning_LogsMessage(t *testing.T) {
	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	Init(flags)
	resetKlogState(t, flags)

	var buf bytes.Buffer
	output = &buf
	t.Cleanup(func() { output = os.Stderr })

	Warning("warning message", "resource", "configmap")

	assert.Contains(t, buf.String(), "Warning: warning message")
	assert.Contains(t, buf.String(), "resource=configmap")
}

func TestWarning_OddLengthKeysAndValues(t *testing.T) {
	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	Init(flags)
	resetKlogState(t, flags)

	var buf bytes.Buffer
	output = &buf
	t.Cleanup(func() { output = os.Stderr })

	Warning("warning message", "key", "value", "extra")

	assert.Contains(t, buf.String(), "Warning: warning message")
	assert.Contains(t, buf.String(), "key=value")
	assert.Contains(t, buf.String(), "extra")
}

func TestWarning_EscapesNewlines(t *testing.T) {
	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	Init(flags)
	resetKlogState(t, flags)

	var buf bytes.Buffer
	output = &buf
	t.Cleanup(func() { output = os.Stderr })

	Warning("msg", "key", "value\nFAKE LOG")

	assert.Contains(t, buf.String(), "value\\nFAKE LOG")
	assert.NotContains(t, buf.String(), "value\nFAKE LOG")
}
