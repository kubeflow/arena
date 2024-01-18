package cron

import (
	"fmt"
	"io"
	"strings"
	"time"
)

const (
	formatLayout = "2006-01-02T15:04:05Z"
)

func printLine(w io.Writer, fields ...string) {
	buffer := strings.Join(fields, "\t")
	fmt.Fprintln(w, buffer)
}

func formatTime(t time.Time) string {
	return t.Format(formatLayout)
}
