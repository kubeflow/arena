package cron

import (
	"fmt"
	"io"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"
	"time"
)

const (
	formatLayout = "2006-01-02T15:04:05Z"
)

var gvr = schema.GroupVersionResource{
	Group:    "apps.kubedl.io",
	Version:  "v1alpha1",
	Resource: "crons",
}

func printLine(w io.Writer, fields ...string) {
	buffer := strings.Join(fields, "\t")
	fmt.Fprintln(w, buffer)
}

func parseTime(strTime string) (time.Time, error) {
	return time.Parse(formatLayout, strTime)
}

func formatTime(t time.Time) string {
	return t.Format(formatLayout)
}
