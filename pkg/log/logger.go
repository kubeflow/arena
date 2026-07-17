package log

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

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
// klog v2 does not provide a structured warning API (WarningS/WarningSDepth),
// so key-value pairs are formatted as "key=value" and appended to the message
// for consistency with Info and Error structured output.
func Warning(msg string, keysAndValues ...interface{}) {
	if len(keysAndValues) == 0 {
		klog.WarningDepth(1, msg)
		return
	}

	var b strings.Builder
	b.WriteString(msg)
	for i := 0; i+1 < len(keysAndValues); i += 2 {
		fmt.Fprintf(&b, " %v=%v", keysAndValues[i], keysAndValues[i+1])
	}
	klog.WarningDepth(1, b.String())
}

// SetVerbosity configures the klog verbosity level on the provided FlagSet.
// The FlagSet should be the same one passed to Init.
// Level 0 = errors/warnings only, 1 = info, 2+ = debug detail.
func SetVerbosity(fs *flag.FlagSet, level int32) error {
	return fs.Set("v", strconv.Itoa(int(level)))
}
