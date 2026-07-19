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
//
// klog v2 does not export a structured warning API (WarningS/WarningSDepth),
// unlike InfoS and ErrorS. To keep warnings parseable by log aggregation
// tools, we manually format key-value pairs as "key=value" segments and
// pass the combined string to WarningDepth. Odd-length key-value slices
// (a caller bug) include the orphaned key with an empty value rather than
// being silently dropped.
func Warning(msg string, keysAndValues ...interface{}) {
	if len(keysAndValues) == 0 {
		klog.WarningDepth(1, msg)
		return
	}

	var b strings.Builder
	b.WriteString(msg)
	for i := 0; i < len(keysAndValues); i += 2 {
		b.WriteByte(' ')
		fmt.Fprintf(&b, "%v", keysAndValues[i])
		if i+1 < len(keysAndValues) {
			b.WriteByte('=')
			fmt.Fprintf(&b, "%v", keysAndValues[i+1])
		}
	}
	klog.WarningDepth(1, b.String())
}

// SetVerbosity configures the klog verbosity level on the provided FlagSet.
// The FlagSet should be the same one passed to Init.
// Level 0 = errors/warnings only, 1 = info, 2+ = debug detail.
func SetVerbosity(fs *flag.FlagSet, level int32) error {
	return fs.Set("v", strconv.Itoa(int(level)))
}
