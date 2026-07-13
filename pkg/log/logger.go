package log

import (
	"flag"
	"fmt"
	"strconv"

	"k8s.io/klog/v2"
)

// Init initializes the logging system with the given flag set.
// It registers klog flags and configures verbosity based on the -v flag.
func Init(flags *flag.FlagSet) {
	klog.InitFlags(flags)
}

// V returns true if the verbosity level is at least the given level.
// Higher levels mean more detailed logging.
func V(level int32) bool {
	return klog.V(klog.Level(level)).Enabled()
}

// Info logs an informational message with optional key-value pairs.
func Info(msg string, keysAndValues ...interface{}) {
	klog.InfoS(msg, keysAndValues...)
}

// Error logs an error message with the associated error and optional key-value pairs.
func Error(err error, msg string, keysAndValues ...interface{}) {
	klog.ErrorS(err, msg, keysAndValues...)
}

// Debug logs a debug message (only visible when verbosity >= 1).
func Debug(msg string, keysAndValues ...interface{}) {
	klog.V(1).InfoS(msg, keysAndValues...)
}

// Warning logs a warning message with optional key-value pairs.
// klog.WarningDepth does not support structured logging, so we format
// the key-value pairs into a readable message.
func Warning(msg string, keysAndValues ...interface{}) {
	if len(keysAndValues) == 0 {
		klog.WarningDepth(1, msg)
		return
	}

	// Format key-value pairs as "key=value"
	var formatted string
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key := fmt.Sprint(keysAndValues[i])
			value := fmt.Sprint(keysAndValues[i+1])
			formatted += fmt.Sprintf(", %s=%s", key, value)
		}
	}

	// Trim leading ", "
	if len(formatted) > 2 {
		formatted = formatted[2:]
	}

	klog.WarningDepth(1, fmt.Sprintf("%s: %s", msg, formatted))
}

// SetVerbosity configures the klog verbosity level.
// Level 0 = errors/warnings only, 1 = info, 2+ = debug detail.
func SetVerbosity(level int32) {
	// klog uses the standard flag package internally.
	// We set the -v flag value to control verbosity.
	flag.Set("v", strconv.Itoa(int(level)))
}
