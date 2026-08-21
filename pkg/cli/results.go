package cli

import (
	"fmt"
	"io"
	"strings"

	outputpkg "github.com/kubeflow/arena/pkg/output"
)

// SubmitResult is the machine-readable result of a successful submit.
type SubmitResult struct {
	Name       string `json:"name" yaml:"name"`
	Namespace  string `json:"namespace" yaml:"namespace"`
	Kind       string `json:"kind" yaml:"kind"`
	APIVersion string `json:"apiVersion" yaml:"apiVersion"`
}

// ActionResult is the machine-readable result of a successful mutation
// (delete, suspend, resume).
type ActionResult struct {
	Name      string `json:"name" yaml:"name"`
	Namespace string `json:"namespace" yaml:"namespace"`
	Kind      string `json:"kind" yaml:"kind"`
	Action    string `json:"action" yaml:"action"` // deleted | suspended | resumed
}

// printSubmitResult writes the submit result honoring -o/--output: the
// legacy one-liner unless json or yaml was explicitly requested.
func printSubmitResult(out io.Writer, name, namespace, kind, apiVersion string) error {
	switch outputpkg.Format(outputFormat) {
	case outputpkg.FormatJSON, outputpkg.FormatYAML:
		return outputpkg.Format(outputFormat).Render(out, &SubmitResult{
			Name:       name,
			Namespace:  namespace,
			Kind:       kind,
			APIVersion: apiVersion,
		}, outputpkg.RenderOptions{})
	}
	fmt.Fprintf(out, "Job %s submitted successfully\n", name)
	return nil
}

// printActionResult writes a mutation result honoring -o/--output: the
// legacy one-liner unless json or yaml was explicitly requested.
func printActionResult(out io.Writer, name, namespace, kind, action string) error {
	switch outputpkg.Format(outputFormat) {
	case outputpkg.FormatJSON, outputpkg.FormatYAML:
		return outputpkg.Format(outputFormat).Render(out, &ActionResult{
			Name:      name,
			Namespace: namespace,
			Kind:      kind,
			Action:    action,
		}, outputpkg.RenderOptions{})
	}
	fmt.Fprintf(out, "%s/%s %s\n", strings.ToLower(kind), name, action)
	return nil
}
