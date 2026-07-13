// Package output provides table formatting for CLI display of training job information.
package output

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/kubeflow/arena/pkg/client"
)

// maxJobListWidths caps column widths in the job list table to prevent excessively wide output.
var maxJobListWidths = map[string]int{
	"NAME":     50,
	"STATUS":   20,
	"REPLICAS": 15,
	"AGE":      15,
}

// maxPodWidths caps column widths in the pod table to prevent excessively wide output.
var maxPodWidths = map[string]int{
	"NAME":   50,
	"STATUS": 20,
	"IP":     20,
	"NODE":   30,
}

// maxJobListWideWidths caps column widths in the wide job list table.
var maxJobListWideWidths = map[string]int{
	"NAME":       50,
	"NAMESPACE":  20,
	"STATUS":     20,
	"APIVERSION": 30,
	"FRAMEWORK":  15,
	"GPU":        10,
	"REPLICAS":   15,
	"AGE":        15,
}

// maxTopJobWidths caps column widths in the top job table.
var maxTopJobWidths = map[string]int{
	"NAME":          50,
	"STATUS":        20,
	"GPU_REQUESTED": 15,
	"REPLICAS":      15,
	"AGE":           15,
}

// maxTopJobWideWidths caps column widths in the wide top job table.
var maxTopJobWideWidths = map[string]int{
	"NAME":          50,
	"NAMESPACE":     20,
	"STATUS":        20,
	"APIVERSION":    30,
	"FRAMEWORK":     15,
	"GPU_REQUESTED": 15,
	"REPLICAS":      15,
	"AGE":           15,
}

// TableRenderer renders job information as aligned text tables for terminal display.
type TableRenderer struct{}

// NewTableRenderer returns a new TableRenderer instance.
func NewTableRenderer() *TableRenderer {
	return &TableRenderer{}
}

// calculateWidths determines the display width for each column by inspecting all cell values,
// expanding each column to fit the widest content it contains, capped by maxCaps where defined.
func calculateWidths(headers []string, rows [][]string, maxCaps map[string]int) []int {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
		if cap, ok := maxCaps[h]; ok {
			widths[i] = min(widths[i], cap)
		}
	}
	for _, row := range rows {
		for i, cell := range row {
			cellLen := len(cell)
			if cap, ok := maxCaps[headers[i]]; ok {
				cellLen = min(cellLen, cap)
			}
			if cellLen > widths[i] {
				widths[i] = cellLen
			}
		}
	}
	return widths
}

// truncate shortens s to maxLen characters, appending "..." when truncation occurs.
// If maxLen <= 3 the string is hard-cut without a marker.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// renderRow formats a single table row, truncating each cell to the corresponding width
// and left-padding with spaces so columns align.
func renderRow(cells []string, widths []int) string {
	var parts []string
	for i, cell := range cells {
		truncated := truncate(cell, widths[i])
		parts = append(parts, fmt.Sprintf("%-*s", widths[i], truncated))
	}
	return strings.Join(parts, " ") + "\n"
}

// indentLines prefixes every non-empty line of s with prefix, preserving the
// line structure (including trailing newlines) so the indented block aligns
// under a section header. Empty lines are left untouched to avoid introducing
// trailing whitespace.
func indentLines(s, prefix string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}

// RenderJobList renders a table of job statuses with columns: NAME, STATUS, REPLICAS, AGE.
// Column widths expand dynamically to fit content, up to the caps defined in maxJobListWidths.
// Returns "No jobs found" when the input slice is nil or empty.
func (r *TableRenderer) RenderJobList(jobs []client.JobStatus) string {
	if len(jobs) == 0 {
		return "No jobs found\n"
	}

	headers := []string{"NAME", "STATUS", "REPLICAS", "AGE"}
	var rows [][]string
	for _, job := range jobs {
		rows = append(rows, []string{
			job.Name, job.Status, fmt.Sprintf("%d/%d", job.Ready, job.Replicas), job.Age,
		})
	}

	widths := calculateWidths(headers, rows, maxJobListWidths)
	var sb strings.Builder
	sb.WriteString(renderRow(headers, widths))
	for _, row := range rows {
		sb.WriteString(renderRow(row, widths))
	}
	return sb.String()
}

// RenderJobListWide renders a wide table with additional NAMESPACE, APIVERSION and FRAMEWORK columns.
// Columns: NAME, NAMESPACE, STATUS, APIVERSION, FRAMEWORK, REPLICAS, AGE.
// Returns "No jobs found" when the input slice is nil or empty.
func (r *TableRenderer) RenderJobListWide(jobs []client.JobStatus) string {
	if len(jobs) == 0 {
		return "No jobs found\n"
	}

	headers := []string{"NAME", "NAMESPACE", "STATUS", "APIVERSION", "FRAMEWORK", "GPU", "REPLICAS", "AGE"}
	var rows [][]string
	for _, job := range jobs {
		rows = append(rows, []string{
			job.Name, job.Namespace, job.Status, job.APIVersion, job.Framework,
			fmt.Sprintf("%d", job.GPURequested),
			fmt.Sprintf("%d/%d", job.Ready, job.Replicas), job.Age,
		})
	}

	widths := calculateWidths(headers, rows, maxJobListWideWidths)
	var sb strings.Builder
	sb.WriteString(renderRow(headers, widths))
	for _, row := range rows {
		sb.WriteString(renderRow(row, widths))
	}
	return sb.String()
}

// RenderTopJob renders a table of job GPU requests with columns: NAME, STATUS, GPU_REQUESTED, REPLICAS, AGE.
// Returns "No jobs found" when the input slice is nil or empty.
func (r *TableRenderer) RenderTopJob(jobs []client.JobStatus) string {
	if len(jobs) == 0 {
		return "No jobs found\n"
	}

	headers := []string{"NAME", "STATUS", "GPU_REQUESTED", "REPLICAS", "AGE"}
	var rows [][]string
	for _, job := range jobs {
		rows = append(rows, []string{
			job.Name, job.Status, fmt.Sprintf("%d", job.GPURequested),
			fmt.Sprintf("%d/%d", job.Ready, job.Replicas), job.Age,
		})
	}

	widths := calculateWidths(headers, rows, maxTopJobWidths)
	var sb strings.Builder
	sb.WriteString(renderRow(headers, widths))
	for _, row := range rows {
		sb.WriteString(renderRow(row, widths))
	}
	return sb.String()
}

// RenderTopJobWide renders a wide table with GPU info plus NAMESPACE, APIVERSION and FRAMEWORK columns.
// Columns: NAME, NAMESPACE, STATUS, APIVERSION, FRAMEWORK, GPU_REQUESTED, REPLICAS, AGE.
// Returns "No jobs found" when the input slice is nil or empty.
func (r *TableRenderer) RenderTopJobWide(jobs []client.JobStatus) string {
	if len(jobs) == 0 {
		return "No jobs found\n"
	}

	headers := []string{"NAME", "NAMESPACE", "STATUS", "APIVERSION", "FRAMEWORK", "GPU_REQUESTED", "REPLICAS", "AGE"}
	var rows [][]string
	for _, job := range jobs {
		rows = append(rows, []string{
			job.Name, job.Namespace, job.Status, job.APIVersion, job.Framework,
			fmt.Sprintf("%d", job.GPURequested),
			fmt.Sprintf("%d/%d", job.Ready, job.Replicas), job.Age,
		})
	}

	widths := calculateWidths(headers, rows, maxTopJobWideWidths)
	var sb strings.Builder
	sb.WriteString(renderRow(headers, widths))
	for _, row := range rows {
		sb.WriteString(renderRow(row, widths))
	}
	return sb.String()
}

// RenderJobDetail renders a detailed view of a single job, including its status fields
// and a pod table (when pods are present). The pod sub-table uses dynamic column widths
// capped by maxPodWidths.
func (r *TableRenderer) RenderJobDetail(info *client.JobInfo) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Name:      %s\n", info.Status.Name))
	sb.WriteString(fmt.Sprintf("Namespace: %s\n", info.Status.Namespace))
	sb.WriteString(fmt.Sprintf("Status:    %s\n", info.Status.Status))
	sb.WriteString(fmt.Sprintf("Replicas:  %d/%d\n", info.Status.Ready, info.Status.Replicas))
	sb.WriteString(fmt.Sprintf("Age:       %s\n", info.Status.Age))

	if len(info.Pods) > 0 {
		sb.WriteString("\nPods:\n")

		headers := []string{"NAME", "STATUS", "IP", "NODE"}
		var rows [][]string
		for _, pod := range info.Pods {
			rows = append(rows, []string{pod.Name, pod.Status, pod.IP, pod.Node})
		}

		widths := calculateWidths(headers, rows, maxPodWidths)

		var podSB strings.Builder
		podSB.WriteString(renderRow(headers, widths))
		for _, row := range rows {
			podSB.WriteString(renderRow(row, widths))
		}
		sb.WriteString(indentLines(podSB.String(), "  "))
	}

	if info.Configuration != nil {
		sb.WriteString("\nConfiguration:\n")
		yamlData, err := yaml.Marshal(info.Configuration)
		if err == nil {
			sb.WriteString(indentLines(string(yamlData), "  "))
		}
	}

	return sb.String()
}
